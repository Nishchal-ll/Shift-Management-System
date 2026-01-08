package routes

import (
	"net/http"
	"shift-manager/controllers"
	"shift-manager/middleware"
	"shift-manager/models"
)

func SetupRoutes() {

	models.InitDB()
	// 2. Authentication Routes
	http.HandleFunc("/", controllers.LoginHandler)
	http.HandleFunc("/logout", controllers.LogoutHandler)

	http.HandleFunc("/dashboard", middleware.IsLoggedIn(controllers.DashboardHandler))

	// 4. Admin Actions
	http.HandleFunc("/admin/add-user", controllers.AddUserHandler)
	http.HandleFunc("/assign", controllers.AssignShiftHandler)

	// 5. User Actions
	http.HandleFunc("/user/add-days", controllers.AddDaysHandler)
	http.HandleFunc("/user/request-swap", controllers.RequestSwapHandler)

	http.HandleFunc("/api/allocations", controllers.AllocationsAPIHandler)

	http.HandleFunc("/admin/delete", controllers.DeleteAllocationHandler)
	http.HandleFunc("/admin/edit-allocation", controllers.EditAllocationHandler)

	http.HandleFunc("/admin/delete-user", controllers.DeleteUserHandler)

	http.HandleFunc("/admin/add-shift", controllers.AdminAddShiftHandler)
	http.HandleFunc("/admin/update-quota", controllers.AdminUpdateQuotaHandler)

	http.HandleFunc("/admin/delete-shift", controllers.HandleDeleteShift)

	http.HandleFunc("/admin/approve", controllers.ApproveHandler)
	http.HandleFunc("/admin/replace", controllers.ReplaceHandler)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
}
