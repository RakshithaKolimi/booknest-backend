package main

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/swaggo/swag"

	"booknest/docs"
	"booknest/internal/http/database"
)

const (
	swaggerV1InstanceName = "booknest-v1"
	swaggerV1DocURL       = "/swagger/v1/doc.json"
)

var configureSwaggerOnce sync.Once

// @title           BookNest API
// @version         1.0
// @description     Production API for BookNest online bookstore. v1 is stable; all endpoints require strict request validation and JWT auth where applicable.
// @termsOfService  http://booknest.com/terms/
// @contact.name    BookNest API Support
// @contact.url     https://booknest.com/support
// @contact.email   kolimirakshitha@gmail.com
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http https
// @accept          json
// @produce         json

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Enter JWT as: Bearer {token}
func configureSwagger() {
	configureSwaggerOnce.Do(func() {
		docs.SwaggerInfo.Title = "BookNest API"
		docs.SwaggerInfo.Description = "Production API for BookNest online bookstore. Versioned under /api/v1."
		docs.SwaggerInfo.Version = "v1"
		docs.SwaggerInfo.BasePath = "/api/v1"
		if host := os.Getenv("API_HOST"); host != "" {
			docs.SwaggerInfo.Host = host
		}
		docs.SwaggerInfo.InfoInstanceName = swaggerV1InstanceName

		// Register a named instance to support versioned Swagger UIs.
		swag.Register(swaggerV1InstanceName, docs.SwaggerInfo)
	})
}

func main() {
	configureSwagger()

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
