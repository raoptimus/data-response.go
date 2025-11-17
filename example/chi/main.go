package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	json "github.com/json-iterator/go"
	"github.com/raoptimus/data-response.go/pkg/chiadapter"
	slogadapter "github.com/raoptimus/data-response.go/pkg/logger/adapter/slog"
	dr "github.com/raoptimus/data-response.go/v2"
	"github.com/raoptimus/data-response.go/v2/formatter"
	"github.com/raoptimus/data-response.go/v2/middleware"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	// Create logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Create factory
	factory := dr.New(
		dr.WithLogger(slogadapter.New(logger)),
		dr.WithFormatter(formatter.NewJSON()),
		dr.WithVerbosity(true),
	)

	// Create chi router with DataResponse support
	r := chiadapter.NewRouter(factory)

	// Add global middleware
	r.WithMiddleware(
		middleware.Logging(),
		middleware.Recovery(),
		middleware.Compression(middleware.CompressionOptions{
			Level:   middleware.CompressionLevelDefault,
			MinSize: 1024,
		}),
		chiadapter.WrapChiMiddleware(
			chimiddleware.BasicAuth("", map[string]string{"user": "pass"}),
		),
	)

	// Health check
	r.Get("/health", func(r *http.Request, f *dr.Factory) dr.DataResponse {
		return f.Success(r.Context(), map[string]string{
			"status": "ok",
		})
	})

	// API routes
	r.Route("/api", func(api *chiadapter.Router) {
		// Add API-specific middleware
		api.WithMiddleware(
			middleware.ContentTypeValidator(middleware.ContentTypeValidatorOptions{
				AllowedTypes: []string{"application/json"},
				Methods:      []string{"POST", "PUT", "PATCH"},
			}),
		)

		// Users endpoints
		api.Get("/users", listUsers)
		api.Get("/users/{id}", getUser)
		api.Post("/users", createUser)
		api.Delete("/users/{id}", deleteUser)

		// Admin routes
		api.Route("/admin", func(admin *chiadapter.Router) {
			// Admin-specific middleware
			admin.WithMiddleware(authMiddleware())

			admin.Get("/stats", getStats)
		})
	})

	log.Println("Server starting on :8080")
	http.ListenAndServe(":8080", r)
}

func listUsers(r *http.Request, f *dr.Factory) dr.DataResponse {
	users := []User{
		{ID: 1, Name: "Alice", Email: "alice@example.com"},
		{ID: 2, Name: "Bob", Email: "bob@example.com"},
	}

	return f.Success(r.Context(), users)
}

func getUser(r *http.Request, f *dr.Factory) dr.DataResponse {
	// Get URL parameter using chi adapter
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return f.BadRequest(r.Context(), "invalid user ID")
	}

	if id != 1 {
		return f.NotFound(r.Context(), "user not found")
	}

	user := User{ID: id, Name: "Alice", Email: "alice@example.com"}
	return f.Success(r.Context(), user)
}

func createUser(r *http.Request, f *dr.Factory) dr.DataResponse {
	var user User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return f.BadRequest(r.Context(), "invalid request body")
	}

	user.ID = 3

	return f.Created(r.Context(), user, "/api/users/3")
}

func deleteUser(r *http.Request, f *dr.Factory) dr.DataResponse {
	idStr := chi.URLParam(r, "id")
	_, err := strconv.Atoi(idStr)
	if err != nil {
		return f.BadRequest(r.Context(), "invalid user ID")
	}

	return f.NoContent(r.Context())
}

func getStats(r *http.Request, f *dr.Factory) dr.DataResponse {
	stats := map[string]interface{}{
		"users":    100,
		"requests": 1000,
		"uptime":   "24h",
	}

	return f.Success(r.Context(), stats)
}

func authMiddleware() dr.Middleware {
	return func(next dr.Handler) dr.Handler {
		return dr.HandlerFunc(func(r *http.Request, f *dr.Factory) dr.DataResponse {
			token := r.Header.Get("Authorization")
			if token != "Bearer secret" {
				return f.Unauthorized(r.Context(), "invalid token")
			}

			return next.Handle(r, f)
		})
	}
}
