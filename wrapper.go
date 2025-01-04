package lambdaurl

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
)

// WrapHandler wraps a standard http.HandlerFunc to work with AWS Lambda Function URL.
func WrapHandler(handler http.HandlerFunc) func(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	return func(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
		// Convert LambdaFunctionURLRequest to http.Request
		httpRequest, err := convertLambdaRequestToHTTPRequest(request)
		if err != nil {
			return events.LambdaFunctionURLResponse{}, err
		}

		// Create a response recorder to capture the response
		responseRecorder := NewResponseRecorder()

		// Call the standard HTTP handler
		handler(responseRecorder, httpRequest)

		// Convert the recorded response to LambdaFunctionURLResponse
		lambdaResponse := convertHTTPResponseToLambdaResponse(responseRecorder)

		return lambdaResponse, nil
	}
}

// convertLambdaRequestToHTTPRequest converts a LambdaFunctionURLRequest to an http.Request.
func convertLambdaRequestToHTTPRequest(lambdaRequest events.LambdaFunctionURLRequest) (*http.Request, error) {
	// Parse the request URL
	requestURL, err := url.Parse(lambdaRequest.RequestContext.HTTP.Path)
	if err != nil {
		return nil, err
	}

	// Create the http.Request
	httpRequest := &http.Request{
		Method: lambdaRequest.RequestContext.HTTP.Method,
		URL:    requestURL,
		Header: make(http.Header),
		Body:   io.NopCloser(bytes.NewBufferString(lambdaRequest.Body)),
	}

	// Set headers
	for key, value := range lambdaRequest.Headers {
		httpRequest.Header.Set(key, value)
	}

	return httpRequest, nil
}

// convertHTTPResponseToLambdaResponse converts an http.ResponseWriter to a LambdaFunctionURLResponse.
func convertHTTPResponseToLambdaResponse(responseRecorder *ResponseRecorder) events.LambdaFunctionURLResponse {
	// Read the response body
	bodyBytes, _ := io.ReadAll(responseRecorder.Body)

	// Create the LambdaFunctionURLResponse
	lambdaResponse := events.LambdaFunctionURLResponse{
		StatusCode: responseRecorder.StatusCode,
		Headers:    make(map[string]string),
		Body:       string(bodyBytes),
	}

	// Set headers
	for key, values := range responseRecorder.Header() {
		lambdaResponse.Headers[key] = values[0]
	}

	return lambdaResponse
}

// ResponseRecorder is a custom http.ResponseWriter that records the response.
type ResponseRecorder struct {
	StatusCode int
	HeaderMap  http.Header
	Body       *bytes.Buffer
}

// NewResponseRecorder creates a new ResponseRecorder.
func NewResponseRecorder() *ResponseRecorder {
	return &ResponseRecorder{
		HeaderMap: make(http.Header),
		Body:      new(bytes.Buffer),
	}
}

// Header returns the header map.
func (r *ResponseRecorder) Header() http.Header {
	return r.HeaderMap
}

// Write writes the response body.
func (r *ResponseRecorder) Write(data []byte) (int, error) {
	return r.Body.Write(data)
}

// WriteHeader writes the status code.
func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
}
