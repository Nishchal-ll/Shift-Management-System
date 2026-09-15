package controllers

import (
	"html/template"
	"net/http"
	"shift-manager/models"
	"time"
)

type LandingData struct {
	CurrentUser string
	CurrentRole string
}

func LandingHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	user := GetSession(r)
	roleCookie, _ := r.Cookie("session_role")
	role := ""
	if roleCookie != nil {
		role = roleCookie.Value
	}

	data := LandingData{
		CurrentUser: user,
		CurrentRole: role,
	}

	tmpl, err := template.ParseFiles("templates/landing.html")
	if err != nil {
		http.Error(w, "Error loading landing page: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		user := GetSession(r)
		if user != "" {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		errorMsg := r.URL.Query().Get("error")
		data := map[string]interface{}{
			"Error": errorMsg,
		}

		tmpl, err := template.ParseFiles("templates/login.html")
		if err != nil {
			http.Error(w, "Error loading login page: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, data)
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
		http.Redirect(w, r, "/login?error=Invalid+credentials.+Please+try+again.", http.StatusSeeOther)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "session_user", Value: "", Path: "/", Expires: time.Now().Add(-1 * time.Hour)})
	http.SetCookie(w, &http.Cookie{Name: "session_role", Value: "", Path: "/", Expires: time.Now().Add(-1 * time.Hour)})
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
