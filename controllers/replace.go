package controllers

import (
	"fmt"
	"net/http"
	"shift-manager/models"
	"strconv"
)

func ReplaceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// ... permission checks ...

		id, _ := strconv.Atoi(r.FormValue("id"))
		newEmployee := r.FormValue("new_employee")

		// Call the new function
		err := models.ReplaceEmployee(id, newEmployee)

		if err != nil {
			// Print error to terminal so you can see WHY it failed
			fmt.Println("Replace Error:", err)
			http.Error(w, "Failed to replace: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/dashboard?view=requests", http.StatusSeeOther)
	}
}
