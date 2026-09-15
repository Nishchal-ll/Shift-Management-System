package controllers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"shift-manager/models"
	"shift-manager/services"
	"strconv"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	user := GetSession(r)
	if user == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	roleCookie, _ := r.Cookie("session_role")
	role := "user"
	if roleCookie != nil {
		role = roleCookie.Value
	}

	view := r.URL.Query().Get("view")
	if view == "" {
		view = "schedule"
	}

	errorMsg := r.URL.Query().Get("error")
	successMsg := r.URL.Query().Get("success")

	// --- 1. TEMPLATE PARSING UPDATE ---
	// Since we split the HTML, we must load the Base + All Partials.
	// Paths are relative to the "backend" folder where you run the command.
	files := []string{
		"templates/base.html",
		"templates/partials/nav.html",
		"templates/partials/view_schedule.html",
		"templates/partials/view_admin.html",
		"templates/partials/view_calendar.html",
		"templates/partials/modal_edit.html",
	}

	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		http.Error(w, "Template Parsing Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// --- 2. DATA FETCHING ---
	var groupedAllocations []models.GroupedAllocation
	var rawAllocations []models.Allocation

	if role == "admin" {
		rawAllocations = models.GetAllocations()
	} else {
		rawAllocations = models.GetUserAllocations(user)
	}

	// --- 3. GROUPING LOGIC (For Schedule View) ---
	if view == "schedule" {
		groupsMap := make(map[string]*models.GroupedAllocation)

		for _, alloc := range rawAllocations {
			detail := models.ShiftDetail{
				ID:                alloc.ID,
				ShiftName:         alloc.ShiftName,
				Status:            alloc.Status,
				NewRequestedShift: alloc.NewRequestedShift,
				StartDate:         alloc.StartDate,
				EndDate:           alloc.EndDate,
			}

			if entry, exists := groupsMap[alloc.EmployeeName]; exists {
				entry.Details = append(entry.Details, detail)
			} else {
				newGroup := &models.GroupedAllocation{
					EmployeeName: alloc.EmployeeName,
					Details:      []models.ShiftDetail{detail},
					StartDate:    alloc.StartDate.Format("2006-01-02"),
				}
				groupsMap[alloc.EmployeeName] = newGroup
			}
		}

		for _, group := range groupsMap {
			groupedAllocations = append(groupedAllocations, *group)
		}
	}

	// --- 4. REAL METRIC & CHART DATA PREPARATION ---
	dayNames := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	dayCounts := make(map[string]int)
	for _, d := range dayNames {
		dayCounts[d] = 0
	}

	shiftCounts := make(map[string]int)
	pendingCount := 0
	confirmedCount := 0

	for _, alloc := range rawAllocations {
		if alloc.Status == "Pending" {
			pendingCount++
		} else {
			confirmedCount++
		}
		shiftCounts[alloc.ShiftName]++

		// Calculate daily counts for the days spanned by this allocation
		cur := alloc.StartDate
		for !cur.After(alloc.EndDate) {
			weekday := cur.Weekday()
			var dayKey string
			switch weekday {
			case 1:
				dayKey = "Mon"
			case 2:
				dayKey = "Tue"
			case 3:
				dayKey = "Wed"
			case 4:
				dayKey = "Thu"
			case 5:
				dayKey = "Fri"
			case 6:
				dayKey = "Sat"
			case 0:
				dayKey = "Sun"
			}
			dayCounts[dayKey]++
			cur = cur.AddDate(0, 0, 1)
		}
	}

	maxCount := 1
	for _, d := range dayNames {
		if dayCounts[d] > maxCount {
			maxCount = dayCounts[d]
		}
	}

	type DayStat struct {
		Day           string
		Count         int
		HeightPercent int
	}

	var weeklyBars []DayStat
	for _, d := range dayNames {
		cnt := dayCounts[d]
		height := 12
		if cnt > 0 {
			height = 20 + int(float64(cnt)/float64(maxCount)*70)
		}
		weeklyBars = append(weeklyBars, DayStat{
			Day:           d,
			Count:         cnt,
			HeightPercent: height,
		})
	}

	totalAllocs := len(rawAllocations)
	coverageRate := 100.0
	if totalAllocs > 0 {
		coverageRate = (float64(confirmedCount) / float64(totalAllocs)) * 100.0
	}

	morningPct := 100
	afternoonPct := 100
	nightPct := 100
	if totalAllocs > 0 {
		if m, ok := shiftCounts["Morning"]; ok && m > 0 {
			morningPct = int(float64(m) / float64(totalAllocs) * 100)
		}
		if a, ok := shiftCounts["Afternoon"]; ok && a > 0 {
			afternoonPct = int(float64(a) / float64(totalAllocs) * 100)
		}
		if n, ok := shiftCounts["Night"]; ok && n > 0 {
			nightPct = int(float64(n) / float64(totalAllocs) * 100)
		}
	}

	employees := models.GetAllEmployees()
	notifications := models.GetNotifications(user)
	unreadCount := 0
	for _, n := range notifications {
		if !n.IsRead {
			unreadCount++
		}
	}

	data := map[string]interface{}{
		"CurrentUser":        user,
		"CurrentRole":        role,
		"CurrentView":        view,
		"ErrorMsg":           errorMsg,
		"SuccessMsg":         successMsg,
		"Employees":          employees,
		"Notifications":      notifications,
		"UnreadNotifCount":   unreadCount,
		"ShiftTypes":         models.GetShiftTypes(),
		"GroupedAllocations": groupedAllocations,
		"Allocations":        rawAllocations,
		"TotalStaffCount":    len(employees),
		"TotalShiftsCount":   totalAllocs,
		"PendingCount":       pendingCount,
		"ConfirmedCount":     confirmedCount,
		"CoverageRate":       strconv.FormatFloat(coverageRate, 'f', 1, 64),
		"MorningPct":         morningPct,
		"AfternoonPct":       afternoonPct,
		"NightPct":           nightPct,
		"WeeklyBars":         weeklyBars,
		"CronStatus":         services.GetCronStatus(),
	}

	// --- 5. RENDER ---
	// We use "base" here because your base.html starts with {{define "base"}}
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, "Render Error: "+err.Error(), http.StatusInternalServerError)
	}
}

// --- API & ACTION HANDLERS ---

func ClearNotificationsAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		user := GetSession(r)
		if user != "" {
			models.ClearNotifications(user)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}
}

func AllocationsAPIHandler(w http.ResponseWriter, r *http.Request) {
	user := GetSession(r)
	if user == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	roleCookie, _ := r.Cookie("session_role")
	role := "user"
	if roleCookie != nil {
		role = roleCookie.Value
	}

	var allocations []models.Allocation
	if role == "admin" {
		allocations = models.GetAllocations()
	} else {
		allocations = models.GetUserAllocations(user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allocations)
}

func AdminAddShiftHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		name := r.FormValue("name")
		start := r.FormValue("start_time")
		end := r.FormValue("end_time")
		quotaStr := r.FormValue("quota")

		quota, _ := strconv.Atoi(quotaStr)
		if quota == 0 {
			quota = 5
		}

		if name != "" && start != "" && end != "" {
			models.AddShift(name, start, end, quota)
		}
	}
	http.Redirect(w, r, "/dashboard?view=settings", http.StatusSeeOther)
}

func AdminUpdateQuotaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		id := r.FormValue("id")
		quota, _ := strconv.Atoi(r.FormValue("quota"))
		models.UpdateShiftQuota(id, quota)
	}
	http.Redirect(w, r, "/dashboard?view=settings", http.StatusSeeOther)
}
