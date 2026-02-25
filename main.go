package main

import (
	"log"

	"github.com/joho/godotenv"

	_ "booknest/docs"
	"booknest/internal/http/database"
)

// @title           BookNest API
// @version         1.0
// @description     Online Bookstore backend (BookNest)
// @termsOfService  http://booknest.com/terms/

// @contact.name   Rakshitha
// @contact.email  kolimirakshitha@gmail.com

// @license.name  MIT
// @host          localhost:8080
// @BasePath      /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Cannot load the env file", "err", err)
	}

	// connect to database
	dbpool, err := database.Connect()
	if err != nil {
		log.Fatal("Cannot connect to database", "err", err)
	}

	defer dbpool.Close()

	// Set up server
	r, err := SetupServer(dbpool)
	if err != nil {
		log.Fatal(err)
	}

	StartHTTPServer(r)
}
