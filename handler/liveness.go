package handler

import (
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
	Alive() error
}

type LivenessServiceRegistry struct {
	services map[string]LivenessService
}

func NewLivenessStackedService() *LivenessServiceRegistry {
	return &LivenessServiceRegistry{services: make(map[string]LivenessService)}
}

func (l *LivenessServiceRegistry) Register(name string, serv LivenessService) *LivenessServiceRegistry {
	l.services[name] = serv
	return l
}

func (l *LivenessServiceRegistry) Alive() error {
	result := NewDeadStackedErrors()

	for name, serv := range l.services {
		if err := serv.Alive(); err != nil {
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

		if err := serv.Alive(); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
}
