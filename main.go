package main

import (
	"log"
	"net/http"
	"shift-manager/models"
	"shift-manager/routes"
	"shift-manager/services"
)

func main() {
	models.InitDB()
	services.InitMQTT()
	services.StartCronScheduler()
	routes.SetupRoutes()

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
