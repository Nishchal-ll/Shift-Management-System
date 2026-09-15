package controllers

import (
	"net/http"
	"shift-manager/models"
	"shift-manager/services"
	"strconv"
)

func ApproveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// 1. Get the ID from the form
		id, _ := strconv.Atoi(r.FormValue("id"))

		var emp, newShift string
		models.DB.QueryRow("SELECT employee_name, COALESCE(new_requested_shift, '') FROM allocations WHERE id = $1", id).Scan(&emp, &newShift)

		// 2. Call the Model to update DB
		models.ApproveSwap(id)

		if emp != "" && newShift != "" {
			services.NotifyShiftSwapApproved(emp, newShift)
		}

		// 3. Refresh the Requests Page
		http.Redirect(w, r, "/dashboard?view=requests", http.StatusSeeOther)
	}
}
