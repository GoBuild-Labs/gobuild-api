// @title gobuild-api
// @version 0.7.0-alpha
// @description API for automating operational inventory control at construction sites.
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
package main

import (
	"fmt"
	"net/http"

	"log"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/r4stypasty/gobuild-api/internal/handlers"
	"github.com/r4stypasty/gobuild-api/internal/middleware"
	"github.com/r4stypasty/gobuild-api/internal/repository"

	_ "github.com/r4stypasty/gobuild-api/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		apiKey = "dev-secret-key" // Default API key for development
	}

	db, err := repository.InitDB("materials.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()
	fmt.Println("Database connected successfully")

	materialHandler := &handlers.MaterialHandler{DB: db}

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)

	r.Get("/health", handlers.HealthCheckHandler) // Health check endpoint

	r.Get("/swagger/*", httpSwagger.WrapHandler) // Swagger documentation endpoint

	// API versioning
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.APIKeyAuthMiddleware(apiKey))

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("gobuild-api is running!"))
		})
		r.Get("/materials", materialHandler.GetMaterials)
		r.Post("/materials", materialHandler.CreateMaterial)
		r.Post("/materials/{id}/write-off", materialHandler.WriteOffMaterial)
		r.Get("/alerts", materialHandler.GetAlerts)
		r.Patch("/materials/{id}", materialHandler.UpdateMaterial)
		r.Delete("/materials/{id}", materialHandler.DeleteMaterial)
	})

	fmt.Println("Server is starting on http://localhost:8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
