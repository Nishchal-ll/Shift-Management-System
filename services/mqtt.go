package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"shift-manager/models"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var client mqtt.Client

type NotificationPayload struct {
	Type      string `json:"type"`      // e.g. "shift_assigned", "swap_requested", "swap_approved", "shift_replaced"
	Title     string `json:"title"`     // Toast header title
	Message   string `json:"message"`   // Human readable description
	Employee  string `json:"employee"`  // Target employee
	ShiftName string `json:"shiftName"` // Shift type
	StartDate string `json:"startDate"` // e.g. "2026-09-15"
	EndDate   string `json:"endDate"`   // e.g. "2026-09-20"
	Timestamp string `json:"timestamp"` // ISO string
}

// InitMQTT initializes the MQTT client connection
func InitMQTT() {
	brokerHost := os.Getenv("MQTT_BROKER")
	if brokerHost == "" {
		brokerHost = "mosquitto:1883"
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", brokerHost))
	opts.SetClientID(fmt.Sprintf("shift-manager-backend-%d", time.Now().UnixNano()))
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.OnConnect = func(c mqtt.Client) {
		log.Println("[MQTT] Connected to broker:", brokerHost)
	}
	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		log.Printf("[MQTT] Connection lost: %v\n", err)
	}

	client = mqtt.NewClient(opts)
	go func() {
		token := client.Connect()
		if token.WaitTimeout(5*time.Second) && token.Error() != nil {
			log.Printf("[MQTT] Warning: initial connection failed (%v). Will retry in background.", token.Error())
		}
	}()
}

// PublishNotification persists to DB and sends an event to global shifts channel and user-specific channel
func PublishNotification(payload NotificationPayload) {
	if payload.Timestamp == "" {
		payload.Timestamp = time.Now().Format(time.RFC3339)
	}

	// 1. Always persist in PostgreSQL database
	if payload.Employee != "" {
		models.CreateNotification(payload.Employee, payload.Type, payload.Title, payload.Message, payload.ShiftName)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Println("[MQTT] Failed to marshal payload:", err)
		return
	}

	if client == nil || !client.IsConnected() {
		log.Println("[MQTT] Client not connected yet, skipping message:", string(data))
		return
	}

	// 2. Publish to general shifts topic for live calendar/feed updates
	client.Publish("shiftpro/shifts", 1, false, data)

	// 3. Publish to target employee's personal topic if specified
	if payload.Employee != "" {
		userTopic := fmt.Sprintf("shiftpro/notifications/%s", payload.Employee)
		client.Publish(userTopic, 1, false, data)
	}
}

// NotifyShiftAssigned fires when admin assigns a shift to an employee
func NotifyShiftAssigned(employee, shiftName, startDate, endDate string) {
	PublishNotification(NotificationPayload{
		Type:      "shift_assigned",
		Title:     "New Shift Assigned",
		Message:   fmt.Sprintf("You have been assigned to %s shift (%s to %s)", shiftName, startDate, endDate),
		Employee:  employee,
		ShiftName: shiftName,
		StartDate: startDate,
		EndDate:   endDate,
	})
}

// NotifyShiftSwapRequested fires when employee requests a shift swap
func NotifyShiftSwapRequested(employee, currentShift, requestedShift string) {
	PublishNotification(NotificationPayload{
		Type:      "swap_requested",
		Title:     "Swap Requested",
		Message:   fmt.Sprintf("%s requested to swap from %s to %s", employee, currentShift, requestedShift),
		Employee:  employee,
		ShiftName: requestedShift,
	})
}

// NotifyShiftSwapApproved fires when admin approves a shift swap
func NotifyShiftSwapApproved(employee, newShift string) {
	PublishNotification(NotificationPayload{
		Type:      "swap_approved",
		Title:     "Swap Request Approved",
		Message:   fmt.Sprintf("Your request for %s shift has been approved!", newShift),
		Employee:  employee,
		ShiftName: newShift,
	})
}

// NotifyShiftReplaced fires when an employee shift is replaced or swapped by admin
func NotifyShiftReplaced(oldEmployee, newEmployee, shiftName, startDate, endDate string) {
	// Notify new employee
	PublishNotification(NotificationPayload{
		Type:      "shift_assigned",
		Title:     "Shift Assigned (Replacement)",
		Message:   fmt.Sprintf("You have been assigned to %s shift from %s (%s to %s)", shiftName, oldEmployee, startDate, endDate),
		Employee:  newEmployee,
		ShiftName: shiftName,
		StartDate: startDate,
		EndDate:   endDate,
	})

	// Notify old employee
	if oldEmployee != "" {
		PublishNotification(NotificationPayload{
			Type:      "shift_replaced",
			Title:     "Shift Handed Over",
			Message:   fmt.Sprintf("Your %s shift (%s to %s) was transferred to %s", shiftName, startDate, endDate, newEmployee),
			Employee:  oldEmployee,
			ShiftName: shiftName,
			StartDate: startDate,
			EndDate:   endDate,
		})
	}
}
