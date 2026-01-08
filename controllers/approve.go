package controllers

import (
	"net/http"
	"shift-manager/models"
	"strconv"
)

func ApproveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// 1. Get the ID from the form
		id, _ := strconv.Atoi(r.FormValue("id"))

		// 2. Call the Model to update DB
		models.ApproveSwap(id)

		// 3. Refresh the Requests Page
		http.Redirect(w, r, "/dashboard?view=requests", http.StatusSeeOther)
	}
}
