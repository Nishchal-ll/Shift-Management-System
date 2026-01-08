package controllers

import (
	"log"
	"net/http"
	"net/url"
	"shift-manager/models"
	"strconv"
	"time"
)

// 2. Add Days (User)
func AddDaysHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		id, _ := strconv.Atoi(r.FormValue("id"))
		days, _ := strconv.Atoi(r.FormValue("days"))
		models.AddDaysToShift(id, days)
	}
	http.Redirect(w, r, "/dashboard?view=schedule", http.StatusSeeOther)
}

// 3. Request Swap (User)
func RequestSwapHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		id, _ := strconv.Atoi(r.FormValue("id"))
		newShift := r.FormValue("new_shift")
		models.RequestSwap(id, newShift)
	}
	http.Redirect(w, r, "/dashboard?view=schedule", http.StatusSeeOther)
}

// 4. Approve Request (Admin)
func ApproveRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		id, _ := strconv.Atoi(r.FormValue("id"))
		models.ApproveSwap(id)
	}
	http.Redirect(w, r, "/dashboard?view=requests", http.StatusSeeOther)
}

// 5 . Update Quota (Admin)
func UpdateQuotaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// 1. Security Check
		role, _ := r.Cookie("session_role")
		if role == nil || role.Value != "admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 2. Get Data
		shiftName := r.FormValue("shift_name")
		quotaStr := r.FormValue("quota")
		quota, err := strconv.Atoi(quotaStr)

		if err != nil {
			http.Error(w, "Invalid number", http.StatusBadRequest)
			return
		}

		// 3. Update Database
		err = models.UpdateShiftQuota(shiftName, quota)
		if err != nil {
			http.Error(w, "Database Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 4. Redirect back to Quota View
		http.Redirect(w, r, "/dashboard?view=quotas", http.StatusSeeOther)
	}
}
func AssignShiftHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		employee := r.FormValue("employee_name")
		shiftName := r.FormValue("shift_type")
		startStr := r.FormValue("start_date")
		endStr := r.FormValue("end_date")

		// Parse Dates
		start, _ := time.Parse("2006-01-02", startStr)
		end, _ := time.Parse("2006-01-02", endStr)

		// 1. Check Quota (This keeps shiftName because Quota depends on the shift type)
		if err := models.CheckQuotaAvailability(shiftName, startStr, endStr); err != nil {
			msg := url.QueryEscape(err.Error())
			http.Redirect(w, r, "/dashboard?view=schedule&error="+msg, http.StatusSeeOther)
			return
		}

		// 2. Check Conflict (FIXED: REMOVED shiftName)
		// We only pass 4 arguments here now:
		if err := models.CheckAvailability(employee, startStr, endStr, -1); err != nil {
			msg := url.QueryEscape(err.Error())
			http.Redirect(w, r, "/dashboard?view=schedule&error="+msg, http.StatusSeeOther)
			return
		}

		// 3. Create Allocation
		models.CreateAllocation(employee, shiftName, start, end)

		http.Redirect(w, r, "/dashboard?view=schedule&success=1", http.StatusSeeOther)
	}
}

func HandleDeleteShift(w http.ResponseWriter, r *http.Request) {
	// 1. Ensure this is a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Extract ID
	id := r.FormValue("id")
	if id == "" {
		http.Error(w, "Missing Shift ID", http.StatusBadRequest)
		return
	}

	// 3. Call DB (CHANGE "database" to "models")
	err := models.DeleteShiftType(id)

	if err != nil {
		log.Printf("Error deleting shift: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 4. Redirect
	http.Redirect(w, r, "/dashboard?view=settings", http.StatusSeeOther)
}
