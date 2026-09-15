package routes

import (
	"net/http"
	"shift-manager/controllers"
	"shift-manager/middleware"
)

func SetupRoutes() {
	// 1. Landing & Authentication Routes
	http.HandleFunc("/", controllers.LandingHandler)
	http.HandleFunc("/login", controllers.LoginHandler)
	http.HandleFunc("/logout", controllers.LogoutHandler)

	// 2. Dashboard
	http.HandleFunc("/dashboard", middleware.IsLoggedIn(controllers.DashboardHandler))

	// 3. Admin Actions
	http.HandleFunc("/admin/add-user", controllers.AddUserHandler)
	http.HandleFunc("/assign", controllers.AssignShiftHandler)
	http.HandleFunc("/admin/delete", controllers.DeleteAllocationHandler)
	http.HandleFunc("/admin/edit-allocation", controllers.EditAllocationHandler)
	http.HandleFunc("/admin/delete-user", controllers.DeleteUserHandler)
	http.HandleFunc("/admin/add-shift", controllers.AdminAddShiftHandler)
	http.HandleFunc("/admin/update-quota", controllers.AdminUpdateQuotaHandler)
	http.HandleFunc("/admin/delete-shift", controllers.HandleDeleteShift)
	http.HandleFunc("/admin/approve", controllers.ApproveHandler)
	http.HandleFunc("/admin/replace", controllers.ReplaceHandler)
	http.HandleFunc("/admin/cron/toggle", controllers.ToggleCronHandler)
	http.HandleFunc("/admin/cron/run-now", controllers.RunCronNowHandler)

	// 4. User Actions
	http.HandleFunc("/user/add-days", controllers.AddDaysHandler)
	http.HandleFunc("/user/request-swap", controllers.RequestSwapHandler)

	// 5. API
	http.HandleFunc("/api/allocations", controllers.AllocationsAPIHandler)
	http.HandleFunc("/api/notifications/clear", controllers.ClearNotificationsAPIHandler)
	http.HandleFunc("/api/cron/status", controllers.CronStatusAPIHandler)

	// 6. Static Files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
}
