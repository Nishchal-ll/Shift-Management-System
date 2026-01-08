package main

import (
	"log"
	"net/http"
	"shift-manager/models"
	"shift-manager/routes"
)

func main() {
	models.InitDB()
	routes.SetupRoutes()

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
