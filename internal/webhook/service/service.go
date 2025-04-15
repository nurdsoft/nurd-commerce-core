package service

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/nurdsoft/nurd-commerce-core/internal/webhook/config"
	"github.com/nurdsoft/nurd-commerce-core/internal/webhook/entities"
	"github.com/cenkalti/backoff/v5"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type Service interface {
	NotifyOrderStatusChange(ctx context.Context, req *entities.NotifyOrderStatusChangeRequest) error
}

type service struct {
	log        *zap.SugaredLogger
	config     config.Config
	httpClient *http.Client
}

func New(
	log *zap.SugaredLogger,
	config config.Config,
	httpClient *http.Client,
) Service {
	return &service{
		log:        log,
		config:     config,
		httpClient: httpClient,
	}
}

func (s *service) NotifyOrderStatusChange(ctx context.Context, req *entities.NotifyOrderStatusChangeRequest) error {
	operation := func() (any, error) {
		// Marshal the request into JSON
		requestBody, err := json.Marshal(req)
		if err != nil {
			s.log.Errorf("Error marshaling request body: %v", err)
			return nil, backoff.Permanent(err) // Do not retry on JSON marshaling failure
		}

		// Create the HTTP request
		url := s.config.OrderURL
		s.log.Infof("Sending order status update: %v", url)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(requestBody))
		if err != nil {
			s.log.Errorf("Error creating request: %v", err)
			return nil, backoff.Permanent(err) // Do not retry on request creation failure
		}
		req.Header.Set("Authorization", s.config.Token)
		req.Header.Set("Content-Type", "application/json")

		// Execute the HTTP request
		resp, err := s.httpClient.Do(req)
		if err != nil {
			s.log.Errorf("Error making request to order: %v", err)
			return nil, err // Retry on transient network errors
		}
		defer resp.Body.Close()

		// Check for non-2xx status codes
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			s.log.Errorf("Non-2xx response: %d, body: %s", resp.StatusCode, string(body))
			return nil, errors.New("non-2xx response from server")
		}

		return nil, nil // Success
	}

	err := s.sendWebhookRequest(operation)
	if err != nil {
		return err
	}

	return nil
}

// sendWebhookRequest sends a webhook request with retry logic.
func (s *service) sendWebhookRequest(operation func() (any, error)) error {
	// Backoff retry logic:
	// Retry attempts will be sent based on an exponential backoff strategy.
	// The retry interval increases exponentially with each attempt, with a randomized factor applied.
	// For example, if the initial retry interval is 500ms, the sequence of retry intervals might look like this:
	//
	// Request #  RetryInterval (seconds)  Randomized Interval (seconds)
	//
	//  1          0.5                     [0.25,   0.75]
	//  2          0.75                    [0.375,  1.125]
	//  3          1.125                   [0.562,  1.687]
	//  4          1.687                   [0.8435, 2.53]
	//  5          2.53                    [1.265,  3.795]
	//  6          3.795                   [1.897,  5.692]
	//  7          5.692                   [2.846,  8.538]
	//  8          8.538                   [4.269, 12.807]
	//  9         12.807                   [6.403, 19.210]
	// ...
	// 60   								backoff.Stop

	maxTries := backoff.WithMaxTries(entities.MaxRetries)
	if _, err := backoff.Retry(
		context.TODO(),
		operation,
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		maxTries,
	); err != nil {
		return err
	}

	return nil
}
