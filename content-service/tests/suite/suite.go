package suite

import (
	"net/http"
	"testing"
	"time"
)

type ContentClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

type Suite struct {
	*testing.T
	ContentClient *ContentClient
}

func New(t *testing.T) *Suite {
	t.Helper()
	t.Parallel()

	contentClient := &ContentClient{
		BaseURL: "http://localhost:9089",
		HTTPClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}

	return &Suite{
		T:             t,
		ContentClient: contentClient,
	}
}
