package services

import (
	"fmt"
	"log"
	"shift-manager/models"
	"sync"
	"time"
)

type CronStatus struct {
	Enabled        bool      `json:"enabled"`
	Interval       string    `json:"interval"`
	LastRun        time.Time `json:"last_run"`
	LastRunStr     string    `json:"last_run_str"`
	ShiftsAssigned int       `json:"shifts_assigned"`
	StatusText     string    `json:"status_text"`
}

type AutoScheduler struct {
	mu             sync.RWMutex
	enabled        bool
	ticker         *time.Ticker
	stopChan       chan struct{}
	interval       time.Duration
	lastRun        time.Time
	shiftsAssigned int
}

var scheduler *AutoScheduler
var schedulerOnce sync.Once

// StartCronScheduler initializes and starts the background auto-scheduler runner
func StartCronScheduler() {
	schedulerOnce.Do(func() {
		scheduler = &AutoScheduler{
			enabled:  false, // default disabled until admin toggles ON
			interval: 1 * time.Minute,
			stopChan: make(chan struct{}),
		}

		go scheduler.loop()
		log.Println("[Cron] Auto-Shift Scheduler initialized (Interval: 1m, Default: Paused)")
	})
}

func (s *AutoScheduler) loop() {
	s.ticker = time.NewTicker(s.interval)
	defer s.ticker.Stop()

	for {
		select {
		case <-s.ticker.C:
			s.mu.RLock()
			isEnabled := s.enabled
			s.mu.RUnlock()

			if isEnabled {
				log.Println("[Cron] Executing automated shift allocation cycle...")
				count, _, err := RunAutoAssignCycle()
				if err != nil {
					log.Printf("[Cron] Error during auto-assign cycle: %v\n", err)
				} else {
					log.Printf("[Cron] Cycle completed: %d shifts automatically assigned\n", count)
				}
			}
		case <-s.stopChan:
			return
		}
	}
}

// ToggleCron enables or disables the auto-scheduler
func ToggleCron(enable bool) bool {
	if scheduler == nil {
		StartCronScheduler()
	}
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()

	scheduler.enabled = enable
	log.Printf("[Cron] Auto-Scheduler state changed: Enabled = %v\n", enable)
	return scheduler.enabled
}

// IsCronEnabled returns whether the scheduler is currently active
func IsCronEnabled() bool {
	if scheduler == nil {
		return false
	}
	scheduler.mu.RLock()
	defer scheduler.mu.RUnlock()
	return scheduler.enabled
}

// GetCronStatus returns formatted snapshot for UI and APIs
func GetCronStatus() CronStatus {
	if scheduler == nil {
		return CronStatus{
			Enabled:    false,
			Interval:   "1 minute",
			StatusText: "Paused",
		}
	}

	scheduler.mu.RLock()
	defer scheduler.mu.RUnlock()

	lastRunStr := "Never"
	if !scheduler.lastRun.IsZero() {
		diff := time.Since(scheduler.lastRun)
		if diff < time.Minute {
			lastRunStr = "Just now"
		} else if diff < time.Hour {
			lastRunStr = fmt.Sprintf("%dm ago", int(diff.Minutes()))
		} else {
			lastRunStr = scheduler.lastRun.Format("Jan 02 15:04")
		}
	}

	statusText := "Paused"
	if scheduler.enabled {
		statusText = "Active (Running every 1m)"
	}

	return CronStatus{
		Enabled:        scheduler.enabled,
		Interval:       "1 minute",
		LastRun:        scheduler.lastRun,
		LastRunStr:     lastRunStr,
		ShiftsAssigned: scheduler.shiftsAssigned,
		StatusText:     statusText,
	}
}

// RunAutoAssignCycle performs the intelligent shift generation and allocation
func RunAutoAssignCycle() (int, []string, error) {
	if scheduler != nil {
		scheduler.mu.Lock()
		scheduler.lastRun = time.Now()
		scheduler.mu.Unlock()
	}

	employees := models.GetAllEmployees()
	if len(employees) == 0 {
		return 0, nil, fmt.Errorf("no registered employees found")
	}

	shiftTypes := models.GetShiftTypes()
	if len(shiftTypes) == 0 {
		return 0, nil, fmt.Errorf("no shift types configured in database")
	}

	assignedCount := 0
	var assignedSummaries []string

	// Plan shift allocation for the next 7 days starting from tomorrow
	now := time.Now()
	startDate := now.AddDate(0, 0, 1) // tomorrow
	daysToSchedule := 7

	for d := 0; d < daysToSchedule; d++ {
		targetDate := startDate.AddDate(0, 0, d)
		targetDateStr := targetDate.Format("2006-01-02")

		for empIdx, emp := range employees {
			// Check if employee already has a shift on targetDate
			err := models.CheckAvailability(emp, targetDateStr, targetDateStr, -1)
			if err != nil {
				// Already scheduled or conflict on this day, skip
				continue
			}

			// Rotate shift selection across available shift types
			shiftCandidate := shiftTypes[(empIdx+d)%len(shiftTypes)]

			// Check shift quota availability
			err = models.CheckQuotaAvailability(shiftCandidate.Name, targetDateStr, targetDateStr)
			if err != nil {
				// Quota full for this shift, try another shift type
				found := false
				for _, altShift := range shiftTypes {
					if altShift.Name != shiftCandidate.Name {
						if errAlt := models.CheckQuotaAvailability(altShift.Name, targetDateStr, targetDateStr); errAlt == nil {
							shiftCandidate = altShift
							found = true
							break
						}
					}
				}
				if !found {
					continue // All shift quotas filled for this day
				}
			}

			// Parse date objects for allocation insertion
			t1, _ := time.Parse("2006-01-02", targetDateStr)
			t2 := t1 // 1 day allocation

			// 1. Insert allocation into PostgreSQL
			if err := models.CreateAllocation(emp, shiftCandidate.Name, t1, t2); err != nil {
				log.Printf("[Cron] Failed to create allocation for %s: %v\n", emp, err)
				continue
			}

			// 2. Dispatch real-time MQTT notification & store alert in database
			NotifyShiftAssigned(emp, shiftCandidate.Name, targetDateStr, targetDateStr)

			assignedCount++
			summary := fmt.Sprintf("Assigned %s to %s on %s", emp, shiftCandidate.Name, targetDateStr)
			assignedSummaries = append(assignedSummaries, summary)
		}
	}

	if scheduler != nil {
		scheduler.mu.Lock()
		scheduler.shiftsAssigned += assignedCount
		scheduler.mu.Unlock()
	}

	// Broadcast an overall shifts update so connected clients refresh their calendar/views
	if assignedCount > 0 {
		PublishNotification(NotificationPayload{
			Type:      "shifts_auto_generated",
			Title:     "Automated Roster Generated",
			Message:   fmt.Sprintf("Cron job automatically generated %d shift allocations for the upcoming cycle.", assignedCount),
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}

	return assignedCount, assignedSummaries, nil
}
