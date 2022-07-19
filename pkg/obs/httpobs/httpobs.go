package httpobs

import (
	"context"
	"net/http"
)

// HTTPObs has values extracted from a request
type HTTPObs struct {
	Request *http.Request
	StrTags map[string]string
}

// ExtractTelemetryRequestAndFields creates a new request with fields
// that should be kept private removed, and a list of tags that
// can be fed to the Insighter interface
func ExtractTelemetryRequestAndFields(req *http.Request) (*HTTPObs, error) {
	// TODO: this probably needs some work :)
	return &HTTPObs{
		Request: req.Clone(context.Background()),
		StrTags: make(map[string]string),
	}, nil
}
