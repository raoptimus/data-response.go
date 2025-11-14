package main

import (
	"net/http"
	"strings"

	dr "github.com/raoptimus/data-response.go/v2"
)

type Auth struct {
	factory *dr.Factory
	tokenChecker func(token string) bool
}

// NewAuth creates a new auth middleware.
func NewAuth(factory *dr.Factory, tokenChecker func(token string) bool) *Auth {
	return &Auth{
		factory:      factory,
		tokenChecker: tokenChecker,
	}
}

// ServeHTTP implements Middleware interface.
func (a *Auth) ServeHTTP(r *http.Request, next dr.Handler) dr.DataResponse {
	authHeader := r.Header.Get(dr.HeaderAuthorization)
	if authHeader == "" {
		return a.factory.Unauthorized(r.Context(), "Authorization required")
	}

	// Extract Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return a.factory.Unauthorized(r.Context(), "Invalid authorization header")
	}

	token := parts[1]
	if !a.tokenChecker(token) {
		return a.factory.Unauthorized(r.Context(), "Invalid token")
	}

	// Token is valid, proceed to next handler
	return next.Handle(r)
}
