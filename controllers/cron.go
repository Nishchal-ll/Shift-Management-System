package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"shift-manager/services"
	"strings"
)

// ToggleCronHandler handles toggling the cron auto-assign state (Admin only)
func ToggleCronHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roleCookie, _ := r.Cookie("session_role")
	if roleCookie == nil || roleCookie.Value != "admin" {
		http.Error(w, "Unauthorized: Admin privileges required", http.StatusUnauthorized)
		return
	}

	// Parse desired state or toggle current state
	stateParam := r.FormValue("enabled")
	var newState bool
	if stateParam != "" {
		newState = (stateParam == "true" || stateParam == "1" || stateParam == "on")
	} else {
		// Toggle
		newState = !services.IsCronEnabled()
	}

	enabled := services.ToggleCron(newState)

	// If AJAX / JSON requested
	if strings.Contains(r.Header.Get("Accept"), "application/json") || r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"enabled": enabled,
			"status":  services.GetCronStatus(),
		})
		return
	}

	msg := url.QueryEscape(fmt.Sprintf("Auto-Scheduler is now %s", map[bool]string{true: "ACTIVE", false: "PAUSED"}[enabled]))
	http.Redirect(w, r, "/dashboard?view=schedule&success="+msg, http.StatusSeeOther)
}

// RunCronNowHandler triggers an immediate auto-assignment cycle (Admin only)
func RunCronNowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roleCookie, _ := r.Cookie("session_role")
	if roleCookie == nil || roleCookie.Value != "admin" {
		http.Error(w, "Unauthorized: Admin privileges required", http.StatusUnauthorized)
		return
	}

	count, summaries, err := services.RunAutoAssignCycle()
	if err != nil {
		if strings.Contains(r.Header.Get("Accept"), "application/json") || r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		http.Redirect(w, r, "/dashboard?view=schedule&error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}

	if strings.Contains(r.Header.Get("Accept"), "application/json") || r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"count":     count,
			"summaries": summaries,
			"status":    services.GetCronStatus(),
		})
		return
	}

	msg := url.QueryEscape(fmt.Sprintf("Auto-Scheduler successfully allocated %d new shifts across staff!", count))
	http.Redirect(w, r, "/dashboard?view=schedule&success="+msg, http.StatusSeeOther)
}

// CronStatusAPIHandler returns the live status of the cron runner
func CronStatusAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services.GetCronStatus())
}
