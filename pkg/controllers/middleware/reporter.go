package middleware

import (
	"context"
	"net/http"
)

type ReporterRequest struct {
	Name               string
	ReporterInstanceId string
	Type               string
	URL                string
}

// Reporter is part of some auth story.  We'll pull a token out of somewhere, validate it, and extract
// reporter information from it and maybe the local database.
func Reporter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// placeholder logic
		ua := r.Header.Get("User-Agent")
		reporterRequest := &ReporterRequest{
			Name:               ua,
			ReporterInstanceId: ua,
			Type:               ua,
			URL:                ua,
		}

		ctx := context.WithValue(r.Context(), ReporterRequestKey, reporterRequest)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

var (
	ReporterRequestKey = &contextKey{"reporterRequest"}
	GetReporter        = GetFromContext[ReporterRequest](ReporterRequestKey)
)
