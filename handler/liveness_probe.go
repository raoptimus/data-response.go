package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	dr "github.com/raoptimus/data-response.go/v2"
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

func LivenessProbe(serv LivenessService) dr.HandlerFunc {
	return func(r *http.Request, f *dr.Factory) dr.DataResponse {
		if err := serv.Alive(r.Context()); err != nil {
			return f.ServiceUnavailable(r.Context(), err.Error())
		} else {
			return f.Success(r.Context(), nil)
		}
	}
}

