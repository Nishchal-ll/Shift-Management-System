package controllers

import (
	"net/http"
	"shift-manager/models"
	"strings"
)

func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("user_name")
		if username != "" {
			models.CreateUser(strings.ToLower(strings.TrimSpace(username)))
		}
	}
	http.Redirect(w, r, "/dashboard?view=assign", http.StatusSeeOther)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// 1. Check Admin Permissions
		role, _ := r.Cookie("session_role")
		if role == nil || role.Value != "admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 2. Get the username to delete
		username := r.FormValue("username")

		// 3. Run the complete delete logic
		err := models.DeleteUserComplete(username)
		if err != nil {
			http.Error(w, "Could not delete user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 4. Refresh the page
		http.Redirect(w, r, "/dashboard?view=users", http.StatusSeeOther)
	}
}
