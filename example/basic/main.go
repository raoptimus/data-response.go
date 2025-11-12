package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	dataresponse "github.com/raoptimus/data-response.go"
	"github.com/raoptimus/data-response.go/formatter"
	"github.com/raoptimus/data-response.go/middleware"
	adapterslog "github.com/raoptimus/data-response.go/pkg/logger/adapter/slog"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	factory := dataresponse.New(
		dataresponse.WithLogger(adapterslog.New(logger)),
		dataresponse.WithVerbosity(os.Getenv("APP_ENV") != "production"),
		dataresponse.WithFormatter(formatter.NewJSON()),
	)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/users", func(w http.ResponseWriter, r *http.Request) {
		users := []User{
			{ID: 1, Name: "Alice", Email: "alice@example.com"},
			{ID: 2, Name: "Bob", Email: "bob@example.com"},
		}

		resp := factory.Success(r.Context(), users)
		dataresponse.Write(w, r, resp, factory)
	})

	mux.HandleFunc("GET /api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if id == "999" {
			resp := factory.NotFound(r.Context(), "User not found")
			dataresponse.Write(w, r, resp, factory)
			return
		}

		user := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
		resp := factory.Success(r.Context(), user)
		dataresponse.Write(w, r, resp, factory)
	})

	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		attributeErrors := map[string][]string{
			"email": {"Email is required"},
			"name":  {"Name must be at least 3 characters"},
		}

		resp := factory.ValidationError(r.Context(), "invalid request", attributeErrors)
		dataresponse.Write(w, r, resp, factory)
	})

	// Setup middleware
	formatterMap := map[string]dataresponse.Formatter{
		"application/json": formatter.NewJSON(),
		"application/xml":  formatter.NewXML(),
	}

	negotiator := middleware.NewContentNegotiator(factory, formatterMap, formatter.NewJSON())
	recovery := middleware.NewRecover(factory)
	// todo: it does not work
	allowJSON := middleware.AllowContentType(factory, dataresponse.MimeTypeJSON.String())
	compressor := middleware.NewAutoCompressor(middleware.DefaultCompression)

	handler := compressor.Handler(
		recovery.Handler(
			allowJSON(
				negotiator.Handler(mux),
			),
		),
	)

	log.Println("Server starting on :8080")
	http.ListenAndServe(":8080", handler)
}
