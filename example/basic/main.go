/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	adapterslog "github.com/raoptimus/data-response.go/pkg/logger/adapter/slog"
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	factory := dr.New(
		dr.WithLogger(adapterslog.New(logger)),
		dr.WithVerbosity(os.Getenv("APP_ENV") != "production"),
		dr.WithFormatter(formatter.NewJSON()),
	)

	mux := dr.NewServeMux(factory)

	// Setup middleware
	formatterMap := map[string]dr.Formatter{
		"application/json": formatter.NewJSON(),
		"application/xml":  formatter.NewXML(),
	}

	mux.WithMiddleware(
		middleware.DefaultCompression(),
		middleware.ContentNegotiator(formatterMap),
		middleware.Logging(),
	)

	mux.HandleFunc("/", func(r *http.Request, f *dr.Factory) dr.DataResponse {
		return f.NotFound(r.Context(), "resource not found")
	})

	mux.HandleFunc("GET /api/users", func(r *http.Request, f *dr.Factory) dr.DataResponse {
		users := []User{
			{ID: 1, Name: "Alice", Email: "alice@example.com"},
			{ID: 2, Name: "Bob", Email: "bob@example.com"},
		}

		return f.Success(r.Context(), users)
	})

	mux.HandleFunc("GET /api/users/{id}", func(r *http.Request, f *dr.Factory) dr.DataResponse {
		id := r.PathValue("id")
		if id == "999" {
			return f.NotFound(r.Context(), "User not found")
		}
		user := User{ID: 1, Name: "Alice", Email: "alice@example.com"}

		return f.Success(r.Context(), user)
	})

	mux.HandleFunc("POST /api/users", func(r *http.Request, f *dr.Factory) dr.DataResponse {
		attributeErrors := map[string][]string{
			"email": {"Email is required"},
			"name":  {"Name must be at least 3 characters"},
		}

		return f.ValidationError(r.Context(), "invalid request", attributeErrors)
	})

	//handler := dr.Chain(
	//	factory,
	//	dr.HandlerFunc(func(r *http.Request, f *dr.Factory) dr.DataResponse {
	//		user := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	//
	//		return f.Success(r.Context(), user)
	//	}),
	//	middleware.ContentNegotiator(formatterMap),
	//)


	log.Println("Server starting on :8080")
	http.ListenAndServe(":8080", mux)
}
