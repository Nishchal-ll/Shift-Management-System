package controllers

import (
	"html/template"
	"net/http"
	"shift-manager/models"
	"time"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, _ := template.ParseFiles("templates/login.html")
		tmpl.Execute(w, nil)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	role, ok := models.GetUserRole(username, password)
	if ok {
		// Set cookies for session
		http.SetCookie(w, &http.Cookie{Name: "session_user", Value: username, Path: "/"})
		http.SetCookie(w, &http.Cookie{Name: "session_role", Value: role, Path: "/"})
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "session_user", Value: "", Expires: time.Now()})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Helper to get current user from cookie
func GetSession(r *http.Request) string {
	c, err := r.Cookie("session_user")
	if err != nil {
		return ""
	}
	return c.Value
}
