package controllers

import (
	"fmt"
	"net/http"
	"shift-manager/models"
	"shift-manager/services"
	"strconv"
)

func ReplaceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// ... permission checks ...

		id, _ := strconv.Atoi(r.FormValue("id"))
		newEmployee := r.FormValue("new_employee")

		var currentOwner, shiftName, startDate, endDate string
		models.DB.QueryRow(`SELECT employee_name, shift_name, start_date::text, end_date::text FROM allocations WHERE id = $1`, id).
			Scan(&currentOwner, &shiftName, &startDate, &endDate)

		// Call the replacement function
		err := models.ReplaceEmployee(id, newEmployee)

		if err != nil {
			// Print error to terminal so you can see WHY it failed
			fmt.Println("Replace Error:", err)
			http.Error(w, "Failed to replace: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if len(startDate) > 10 {
			startDate = startDate[:10]
		}
		if len(endDate) > 10 {
			endDate = endDate[:10]
		}
		services.NotifyShiftReplaced(currentOwner, newEmployee, shiftName, startDate, endDate)

		http.Redirect(w, r, "/dashboard?view=requests", http.StatusSeeOther)
	}
}
