package middleware

import (
	"net/http"
	"shift-manager/controllers" // Import to access GetSession
)

// AuthMiddleware checks if a user is logged in before letting them see the page
func IsLoggedIn(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Check if user has a session
		user := controllers.GetSession(r)

		// 2. If no user, kick them to login page
		if user == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// 3. If user exists, let them pass to the next handler
		next(w, r)
	}
}
