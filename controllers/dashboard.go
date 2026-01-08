package controllers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"shift-manager/models"
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

	// --- 4. DATA PREPARATION ---
	data := map[string]interface{}{
		"CurrentUser":        user,
		"CurrentRole":        role,
		"CurrentView":        view,
		"ErrorMsg":           errorMsg,
		"SuccessMsg":         successMsg,
		"Employees":          models.GetAllEmployees(),
		"ShiftTypes":         models.GetShiftTypes(),
		"GroupedAllocations": groupedAllocations,
		"Allocations":        rawAllocations,
	}

	// --- 5. RENDER ---
	// We use "base" here because your base.html starts with {{define "base"}}
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, "Render Error: "+err.Error(), http.StatusInternalServerError)
	}
}

// --- API & ACTION HANDLERS (No changes needed here) ---

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
