package service

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/ups/errors"
)

type Session struct {
	AccessToken string
	ClientID    string
	TokenType   string
	IssuedAt    string
	ExpiresIn   string
	Status      string
	httpClient  *http.Client
}

// httpRequest executes an HTTP request to the salesforce server and returns the response data in byte buffer.
func (s *Session) httpRequest(method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("%s %s", s.TokenType, s.AccessToken))
	req.Header.Add("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Println(logPrefix, "request failed,", resp.StatusCode)
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(resp.Body)
		if err != nil {
			return nil, err
		}

		newStr := buf.String()

		sfErr := errors.ParseError(resp.StatusCode, buf.Bytes())

		log.Println(logPrefix, "Failed resp.body: ", newStr)
		return nil, sfErr
	}

	return io.ReadAll(resp.Body)
}
