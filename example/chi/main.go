package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	dataresponse "github.com/raoptimus/data-response.go"
	"github.com/raoptimus/data-response.go/formatter"
	adapterslog "github.com/raoptimus/data-response.go/pkg/logger/adapter/slog"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	factory := dataresponse.New(
		dataresponse.WithLogger(adapterslog.New(logger)),
		dataresponse.WithFormatter(formatter.NewJSON()),
	)

	// ✅ Create chi adapter
	chiAdapter := dataresponse.NewChiAdapter(factory)

	// Create chi router
	r := chi.NewRouter()

	// ✅ Use standard chi middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	//r.Use(middleware.Timeout(60 * time.Second))

	// ✅ Convert DataResponse middleware to chi middleware
	authMW := NewAuth(factory, func(token string) bool {
		return token == "secret-token"
	})
	loggerMW := NewLogging(adapterslog.New(logger), factory)

	// Health check - no auth required
	r.Get("/health", chiAdapter.HandlerFunc(func(r *http.Request) dataresponse.DataResponse {
		return factory.Success(r.Context(), map[string]string{"status": "ok"})
	}))

	// API routes with DataResponse middleware
	r.Route("/api", func(r chi.Router) {
		// ✅ Apply DataResponse middleware converted to chi middleware
		r.Use(chiAdapter.Middleware(loggerMW))

		// Public routes
		r.Group(func(r chi.Router) {
			r.Get("/public", chiAdapter.HandlerFunc(func(r *http.Request) dataresponse.DataResponse {
				return factory.Success(r.Context(), map[string]string{"data": "public"})
			}))
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(chiAdapter.Middleware(authMW))

			r.Get("/users", chiAdapter.HandlerFunc(func(r *http.Request) dataresponse.DataResponse {
				users := []User{
					{ID: 1, Name: "Alice", Email: "alice@example.com"},
					{ID: 2, Name: "Bob", Email: "bob@example.com"},
				}
				return factory.Success(r.Context(), users)
			}))

			r.Get("/users/{id}", chiAdapter.HandlerFunc(func(r *http.Request) dataresponse.DataResponse {
				id := chi.URLParam(r, "id")

				if id == "999" {
					return factory.NotFound(r.Context(), "User not found")
				}

				user := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
				return factory.Success(r.Context(), user)
			}))

			r.Post("/users", chiAdapter.HandlerFunc(func(r *http.Request) dataresponse.DataResponse {
				attributeErrors := map[string][]string{
					"email": {"Email is required"},
					"name":  {"Name must be at least 3 characters"},
				}

				return factory.ValidationError(r.Context(), "Validation failed", attributeErrors)
			}))
		})
	})

	log.Println("Server starting on :8080")
	log.Println("Try: curl http://localhost:8080/health")
	log.Println("Try: curl -H 'Authorization: Bearer secret-token' http://localhost:8080/api/users")
	http.ListenAndServe(":8080", r)
}
