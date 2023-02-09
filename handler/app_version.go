package handler

import (
	"fmt"
	"io"
	"net/http"
)

type VersionData struct {
	GitCommit string
	GitBranch string
	Version   string
	BuildDate string
	Name      string
}

func (d *VersionData) String() string {
	return fmt.Sprintf("Name: %s, Commit: %s, branch: %s, version: %s, build date: %s",
		d.Name,
		d.GitCommit,
		d.GitBranch,
		d.Version,
		d.BuildDate,
	)
}

func Version(data *VersionData) http.Handler {
	return VersionFunc(data)
}

func VersionFunc(data *VersionData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)

		if _, err := io.WriteString(w, data.String()); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}
