// Package retry provides retry support for any http client or server.
package retry

import (
	"bytes"
	"context"
	"crypto/x509"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"
)

var (
	// A regular expression to match the error returned by net/http when the
	// configured number of redirects is exhausted. This error isn't typed
	// specifically so we resort to matching on the error string.
	redirectsErrorRe = regexp.MustCompile(`stopped after \d+ redirects\z`)

	// A regular expression to match the error returned by net/http when the
	// scheme specified in the URL is invalid. This error isn't typed
	// specifically so we resort to matching on the error string.
	schemeErrorRe = regexp.MustCompile(`unsupported protocol scheme`)
)

type retryRoundTripper struct {
	timeout time.Duration
	options []retry.Option

	next http.RoundTripper
}

func (h *retryRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	var (
		res       *http.Response
		err       error
		bodyBytes []byte
	)

	if r.Body != nil {
		// The request.Body is actually io.ReadCloser type which drains the body once it was read.
		// When the request body is read, it is drained and not available for the next call which uses the same request.
		// So make a copy of the request body and use the same for subsequent calls.
		bodyBytes, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		r.Body.Close()
	}

	ctx := r.Context()

	operation := func() error {
		tctx, cancel := context.WithTimeout(ctx, h.timeout)
		defer cancel()

		// Restore the request body.
		// Wrap the body with ioutil.NopCloser to get request body to its original state io.ReadCloser.
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		res, err = h.next.RoundTrip(r.WithContext(tctx)) // nolint:bodyclose

		if err != nil {
			if v, ok := err.(*url.Error); ok {
				// Don't retry if the error was due to too many redirects.
				if redirectsErrorRe.MatchString(v.Error()) {
					return nil
				}

				// Don't retry if the error was due to an invalid protocol scheme.
				if schemeErrorRe.MatchString(v.Error()) {
					return nil
				}

				// Don't retry if the error was due to TLS cert verification failure.
				if _, ok := v.Err.(x509.UnknownAuthorityError); ok {
					return nil
				}
			}

			// The error is likely recoverable so retry.
			return err
		}

		// Check the response code. We retry on 500-range responses to allow
		// the server time to recover, as 500's are typically not permanent
		// errors and may relate to outages on the server side. This will catch
		// invalid response codes as well, like 0 and 999.
		if res.StatusCode == 0 || (res.StatusCode >= 500 && res.StatusCode != 501) {
			return errors.Errorf("invalid status code %d", res.StatusCode)
		}

		return nil
	}

	// We don't need to check the error as it's only used to retry. We save the last error in err.
	retry.Do(operation, h.options...) // nolint:errcheck

	return res, err
}

// ClientRoundTripper returns a new round tripper that adds retry.
func ClientRoundTripper(timeout time.Duration, retries uint, next http.RoundTripper, options ...retry.Option) http.RoundTripper {
	options = append(options, retry.Attempts(retries))

	return &retryRoundTripper{timeout, options, next}
}
