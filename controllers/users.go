package controllers

import (
	"net/http"
	"shift-manager/models"
	"strings"
)

func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// 1. Check Admin Permissions
		role, _ := r.Cookie("session_role")
		if role == nil || role.Value != "admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		username := strings.ToLower(strings.TrimSpace(r.FormValue("user_name")))
		password := strings.TrimSpace(r.FormValue("password"))

		// 2. Validate Username contains @
		if !strings.Contains(username, "@") || len(username) < 3 {
			http.Redirect(w, r, "/dashboard?view=add_user&error=Username+must+be+a+valid+email+containing+@", http.StatusSeeOther)
			return
		}

		// 3. Validate Password length
		if len(password) < 6 {
			http.Redirect(w, r, "/dashboard?view=add_user&error=Password+must+be+at+least+6+characters+long", http.StatusSeeOther)
			return
		}

		// 4. Create User
		err := models.CreateUser(username, password)
		if err != nil {
			http.Redirect(w, r, "/dashboard?view=add_user&error=User+already+exists+or+could+not+be+created", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/dashboard?view=users&success=User+account+created+successfully", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/dashboard?view=add_user", http.StatusSeeOther)
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
