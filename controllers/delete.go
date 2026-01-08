package controllers

import (
	"net/http"
	"shift-manager/models"
	"strconv"
)

// Delete Handler
func DeleteAllocationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// 1. Check if Admin
		role, _ := r.Cookie("session_role")
		if role == nil || role.Value != "admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 2. Perform Delete
		id, _ := strconv.Atoi(r.FormValue("id"))
		models.DeleteAllocation(id)

		// 3. Refresh Page
		http.Redirect(w, r, "/dashboard?view=schedule", http.StatusSeeOther)
	}
}

// Edit Handler
func EditAllocationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// 1. Check if Admin
		role, _ := r.Cookie("session_role")
		if role == nil || role.Value != "admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 2. Get Data
		id, _ := strconv.Atoi(r.FormValue("id"))
		employee := r.FormValue("employee_name")
		shift := r.FormValue("shift_name")
		start := r.FormValue("start_date")
		end := r.FormValue("end_date")

		// 3. Perform Update
		models.UpdateAllocation(id, employee, shift, start, end)

		// 4. Refresh Page
		http.Redirect(w, r, "/dashboard?view=schedule", http.StatusSeeOther)
	}
}
