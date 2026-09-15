package middleware

import (
	"net/http"
	"shift-manager/controllers"
)

// IsLoggedIn checks if a user is logged in before letting them see protected pages
func IsLoggedIn(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := controllers.GetSession(r)
		if user == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}
