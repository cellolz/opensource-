package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	//"os"
	"strconv"
	"strings"
	"time"
)

type ExecuteRequest struct {
	Query string `json:"query"`

	// Opcional
	RequestTimeoutMilliseconds int `json:"requestTimeoutMilliseconds,omitempty"`
}

type ExecuteResponse struct {
	State       string `json:"state"`
	RequestToken string `json:"requestToken"`
	Progress    int    `json:"progress"`
}

type PollResponse struct {
	State       string `json:"state"`
	RequestToken string `json:"requestToken"`
	Progress    int    `json:"progress"`

	Result *QueryResult `json:"result,omitempty"`

	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type QueryResult struct {
	Records []map[string]interface{} `json:"records"`
	Types   []TypeInfo                `json:"types,omitempty"`
}

type TypeInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Endpoint struct {
	Method   string
	Route    string
	Requests int64
}

type DynatraceClient struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

func NewDynatraceClient(baseURL, token string) *DynatraceClient {
	return &DynatraceClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (d *DynatraceClient) newRequest(
	ctx context.Context,
	method string,
	requestURL string,
	body io.Reader,
) (*http.Request, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		requestURL,
		body,
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Authorization",
		"Bearer "+d.Token,
	)

	req.Header.Set(
		"Accept",
		"application/json",
	)

	if body != nil {
		req.Header.Set(
			"Content-Type",
			"application/json",
		)
	}

	return req, nil
}

// ------------------------------------------------------------
// POST /query:execute
// ------------------------------------------------------------

func (d *DynatraceClient) ExecuteDQL(
	ctx context.Context,
	query string,
) (*ExecuteResponse, error) {

	payload := ExecuteRequest{
		Query: query,

		// 60 segundos
		RequestTimeoutMilliseconds: 60000,
	}

	body, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	endpoint := d.BaseURL +
		"/platform/storage/query/v1/query:execute"

	req, err := d.newRequest(
		ctx,
		http.MethodPost,
		endpoint,
		strings.NewReader(string(body)),
	)

	if err != nil {
		return nil, err
	}

	resp, err := d.Client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"execute DQL HTTP %d: %s",
			resp.StatusCode,
			string(responseBody),
		)
	}

	var result ExecuteResponse

	if err := json.Unmarshal(
		responseBody,
		&result,
	); err != nil {
		return nil, fmt.Errorf(
			"invalid execute response: %w\n%s",
			err,
			string(responseBody),
		)
	}

	return &result, nil
}

// ------------------------------------------------------------
// GET /query:poll
// ------------------------------------------------------------

func (d *DynatraceClient) PollDQL(
	ctx context.Context,
	requestToken string,
) (*PollResponse, error) {

	u := d.BaseURL +
		"/platform/storage/query/v1/query:poll"

	params := url.Values{}

	params.Set(
		"request-token",
		requestToken,
	)

	u += "?" + params.Encode()

	req, err := d.newRequest(
		ctx,
		http.MethodGet,
		u,
		nil,
	)

	if err != nil {
		return nil, err
	}

	resp, err := d.Client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"poll DQL HTTP %d: %s",
			resp.StatusCode,
			string(responseBody),
		)
	}

	var result PollResponse

	if err := json.Unmarshal(
		responseBody,
		&result,
	); err != nil {
		return nil, fmt.Errorf(
			"invalid poll response: %w\n%s",
			err,
			string(responseBody),
		)
	}

	return &result, nil
}

// ------------------------------------------------------------
// Executa e espera a consulta terminar
// ------------------------------------------------------------

func (d *DynatraceClient) Query(
	ctx context.Context,
	query string, tk string,
) (*QueryResult, error) {

	execute, err := d.ExecuteDQL(ctx, query)

	if err != nil {
		return nil, err
	}

	fmt.Println(
		"State:",
		execute.State,
	)
fmt.Println(execute)

	fmt.Println(
		"RequestToken:",
		execute.RequestToken,
	)

	// Em algumas situações a execução pode
	// já estar concluída.
	if execute.State == "SUCCEEDED" {
		poll, err := d.PollDQL(
			ctx,
			tk,
			//execute.RequestToken,
		)

		if err != nil {
			return nil, err
		}

		return poll.Result, nil
	}

	// Polling
	for {

		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		default:
		}

		time.Sleep(1 * time.Second)

		poll, err := d.PollDQL(
			ctx,
			execute.RequestToken,
		)

		if err != nil {
			return nil, err
		}

		fmt.Printf(
			"State=%s Progress=%d%%\n",
			poll.State,
			poll.Progress,
		)

		switch poll.State {

		case "RUNNING":
			continue

		case "SUCCEEDED":

			if poll.Result == nil {
				return nil, fmt.Errorf(
					"query succeeded but result is empty",
				)
			}

			return poll.Result, nil

		case "FAILED":

			if poll.Error != nil {
				return nil, fmt.Errorf(
					"DQL failed: %s",
					poll.Error.Message,
				)
			}

			return nil, fmt.Errorf(
				"DQL failed",
			)

		default:

			return nil, fmt.Errorf(
				"unexpected DQL state: %s",
				poll.State,
			)
		}
	}
}

// ------------------------------------------------------------
// Converte records para Endpoint
// ------------------------------------------------------------

func ParseEndpoints(
	result *QueryResult,
) []Endpoint {

	endpoints := make([]Endpoint, 0)

	for _, record := range result.Records {

		method, _ := record["http.request.method"].(string)

		route, _ := record["http.route"].(string)

		var requests int64

		switch value := record["requests"].(type) {

		case float64:
			requests = int64(value)

		case int64:
			requests = value

		case json.Number:
			requests, _ = value.Int64()

		case string:
			requests, _ = strconv.ParseInt(
				value,
				10,
				64,
			)
		}

		endpoints = append(
			endpoints,
			Endpoint{
				Method:   method,
				Route:    route,
				Requests: requests,
			},
		)
	}

	return endpoints
}

// ------------------------------------------------------------
// MAIN
// ------------------------------------------------------------

func main() {

	token := ""


	environmentID := "mbz63427"

	serviceID := "SERVICE-43D913196C625734"

	baseURL := fmt.Sprintf(
		"https://%s.apps.dynatrace.com",
		environmentID,
	)

	client := NewDynatraceClient(
		baseURL,
		token,
	)

	query := fmt.Sprintf(`

fetch spans, from:now() - 1h

| filter dt.entity.service == "%s"
  and request.is_root_span == true
  and isNotNull(http.request.method)
  and isNotNull(http.route)

| summarize
    requests = count(),
    by: {
        http.request.method,
        http.route
    }

| sort requests desc

| limit 500

`, serviceID)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Minute,
	)

	defer cancel()

	result, err := client.Query(
		ctx,
		query,
		token,
	)

	if err != nil {
		panic(err)
	}

	endpoints := ParseEndpoints(
		result,
	)

	fmt.Println()
	fmt.Println(
		"Endpoints:",
	)

	fmt.Println(
		"----------------------------------------------",
	)

	for _, endpoint := range endpoints {

		fmt.Printf(
			"%-8s %-50s %d\n",
			endpoint.Method,
			endpoint.Route,
			endpoint.Requests,
		)
	}
}
