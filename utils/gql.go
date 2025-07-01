package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var (
	ErrGraphQLConfigIsNil          = errors.New("GraphQL config is nil")
	ErrGraphQLQueryIsEmpty         = errors.New("Query is empty")
	ErrGraphQLUnexpectedStatusCode = errors.New("Unexpected status code")
	ErrGraphQLUserAgentIsEmpty     = errors.New("User agent is empty in GraphQL config")
	ErrGraphQLGhpTokenIsEmpty      = errors.New("Github token is empty in GraphQL config")
	ErrGraphQLUrlIsEmpty           = errors.New("URL is empty in GraphQL config")
	ErrGraphQLResponseErrors       = errors.New("GraphQL response contains errors")
)

type GQLC struct {
	URL        string
	client     *http.Client
	clientOnce sync.Once
}

type GraphQLRequest[VT any] struct {
	Query     string        `json:"query"`
	Variables map[string]VT `json:"variables"`
}

type GraphQLError struct {
	Message    string            `json:"message"`
	Type       string            `json:"type"`
	Locations  []map[string]int  `json:"locations"`
	Path       []any             `json:"path"`
	Extensions map[string]string `json:"extensions"`
}

type GraphQLResponse[R any] struct {
	Data   R              `json:"data"`
	Errors []GraphQLError `json:"errors"`
}

// getClient lazily initializes the http.Client.
func (g *GQLC) getClient() *http.Client {
	g.clientOnce.Do(func() {
		g.client = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		}
	})
	return g.client
}

func SendGQLRequest[R any, VT any](gqlc *GQLC, ctx context.Context, q string, variables map[string]VT, headers map[string]string) (GraphQLResponse[R], error) {
	var gqlRes GraphQLResponse[R]

	if gqlc == nil {
		return gqlRes, ErrGraphQLConfigIsNil
	} else if gqlc.URL == "" {
		return gqlRes, ErrGraphQLUrlIsEmpty
	} else if q == "" {
		return gqlRes, ErrGraphQLQueryIsEmpty
	}

	payload, err := json.Marshal(GraphQLRequest[VT]{Query: q, Variables: variables})
	if err != nil {
		return gqlRes, err
	}

	reader := bytes.NewReader(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, gqlc.URL, reader)
	if err != nil {
		return gqlRes, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	hc := gqlc.getClient()
	res, err := hc.Do(req)
	if err != nil {
		return gqlRes, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		subErr := fmt.Errorf("status code: %s", res.Status)
		return gqlRes, errors.Join(ErrGraphQLUnexpectedStatusCode, subErr)
	}

	err = json.NewDecoder(res.Body).Decode(&gqlRes)
	if err != nil {
		return gqlRes, err
	}

	if len(gqlRes.Errors) != 0 {
		subErrs := fmt.Errorf("%v", gqlRes.Errors)
		return gqlRes, errors.Join(ErrGraphQLResponseErrors, subErrs)
	}

	return gqlRes, nil
}
