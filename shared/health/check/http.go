// Package check provides a way to do health checks
package check

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

var (
	// ErrCouldNotGetHost with HTTP
	ErrCouldNotGetHost = errors.New("could not get host")
)

type httpChecker struct {
	host   string
	client *http.Client
}

// NewHTTPChecker for health checks
func NewHTTPChecker(host string, client *http.Client) Checker {
	return &httpChecker{host, client}
}

func (c httpChecker) Check(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.host, nil)
	if err != nil {
		return errors.Wrapf(ErrCouldNotGetHost, "the host %s", c.host)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrapf(ErrCouldNotGetHost, "the host %s", c.host)
	}

	defer resp.Body.Close()

	return nil
}
