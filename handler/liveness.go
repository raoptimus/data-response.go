package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type DeadStackedErrors struct {
	errors []error
}

func NewDeadStackedErrors() *DeadStackedErrors {
	return &DeadStackedErrors{errors: make([]error, 0)}
}

func (d *DeadStackedErrors) Add(err error) {
	if err == nil {
		return
	}
	d.errors = append(d.errors, err)
}

func (d *DeadStackedErrors) HasErrors() bool {
	return len(d.errors) > 0
}

func (d *DeadStackedErrors) Error() string {
	if len(d.errors) == 0 {
		return ""
	}
	sb := strings.Builder{}
	for i := range d.errors {
		sb.WriteString(d.errors[i].Error() + "\n")
	}
	return sb.String()
}

type LivenessService interface {
	Alive(ctx context.Context) error
}

type LivenessServiceRegistry struct {
	handles map[string]LivenessHandleFunc
}

type LivenessHandleFunc func(ctx context.Context) error

func NewLivenessServiceRegistry() *LivenessServiceRegistry {
	return &LivenessServiceRegistry{handles: make(map[string]LivenessHandleFunc)}
}

func (l *LivenessServiceRegistry) RegisterFunc(name string, serv LivenessHandleFunc) *LivenessServiceRegistry {
	l.handles[name] = serv
	return l
}

func (l *LivenessServiceRegistry) Register(name string, serv LivenessService) *LivenessServiceRegistry {
	l.handles[name] = func(ctx context.Context) error {
		return serv.Alive(ctx)
	}
	return l
}

func (l *LivenessServiceRegistry) Alive(ctx context.Context) error {
	result := NewDeadStackedErrors()

	for name, h := range l.handles {
		if err := h(ctx); err != nil {
			result.Add(errors.Wrap(err, name))
		}
	}

	if result.HasErrors() {
		return result
	}

	return nil
}

func Liveness(serv LivenessService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		if err := serv.Alive(req.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
}
