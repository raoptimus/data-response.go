package handler

import (
	"fmt"
	"net/http"

	dr "github.com/raoptimus/data-response.go/v2"
	"github.com/raoptimus/data-response.go/v2/response"
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

//nolint:ireturn,nolintlint // its ok
func AppVersion(data *VersionData) dr.Handler {
	return AppVersionFunc(data)
}

func AppVersionFunc(data *VersionData) dr.HandlerFunc {
	return func(r *http.Request, f *dr.Factory) *response.DataResponse {
		return f.Success(r.Context(), data.String()).
			WithHeader(response.HeaderXContentTypeOptions, response.ContentTypeOptionsNoSniff)
	}
}
