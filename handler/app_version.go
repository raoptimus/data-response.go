package handler

import (
	"fmt"
	"net/http"

	dr "github.com/raoptimus/data-response.go/v2"
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

func Version(data *VersionData) dr.Handler {
	return VersionFunc(data)
}

func VersionFunc(data *VersionData) dr.HandlerFunc {
	return func(r *http.Request, f *dr.Factory) dr.DataResponse {
		return f.Success(r.Context(), data.String()).
			WithHeader(dr.HeaderXContentTypeOptions, dr.ContentTypeOptionsNoSniff)
	}
}
