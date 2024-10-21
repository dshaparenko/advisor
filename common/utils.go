package common

// import (
// 	"context"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"time"

// 	prommodel "github.com/prometheus/common/model"
// )

// type promClient interface {
// 	Query(ctx context.Context, query string, ts time.Time) (interface{}, error)
// }

// type PromClient struct {
// 	client    *http.Client
// 	thanosURL string
// }

// // Factory function to create a new PromClient
// func NewPromClient(thanosURL string) *PromClient {
// 	return &PromClient{
// 		client:    &http.Client{Timeout: 10 * time.Second}, // Adjust timeout as necessary
// 		thanosURL: thanosURL,
// 	}
// }

// // Do method for making HTTP requests with context cancellation, similar to the example
// func (c *PromClient) Do(ctx context.Context, req *http.Request) (*http.Response, []byte, error) {
// 	if ctx != nil {
// 		req = req.WithContext(ctx)
// 	}

// 	resp, err := c.client.Do(req)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	defer func() {
// 		if resp != nil {
// 			resp.Body.Close()
// 		}
// 	}()

// 	// Reading the response body asynchronously
// 	var body []byte
// 	done := make(chan struct{})
// 	go func() {
// 		body, err = io.ReadAll(resp.Body)
// 		close(done)
// 	}()

// 	select {
// 	case <-ctx.Done():
// 		<-done
// 		err = resp.Body.Close()
// 		if err == nil {
// 			err = ctx.Err()
// 		}
// 	case <-done:
// 	}

// 	return resp, body, err
// }

// // Query method to query Thanos/Prometheus using the Do method
// func (c *PromClient) Query(ctx context.Context, query string, ts time.Time) (interface{}, error) {
// 	// Construct the request URL
// 	url := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d", c.thanosURL, query, ts.Unix())

// 	// Create a new HTTP request
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create request: %w", err)
// 	}

// 	// Use the Do method to execute the request
// 	resp, body, err := c.Do(ctx, req)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to execute request: %w", err)
// 	}

// 	// Process response (you can unmarshal the JSON response here if needed)
// 	fmt.Printf("Response from Thanos: %s\n", body)

// 	// Simulated response for now (replace this with the actual response parsing)
// 	mockResponse := prommodel.Vector{
// 		&prommodel.Sample{
// 			Value: 10.5,
// 			Metric: prommodel.Metric{
// 				"container": "test-container",
// 			},
// 		},
// 	}

// 	return mockResponse, nil
// }
// func byteCountSI(b int64) string {
// 	const unit = 1000
// 	if b < unit {
// 		return fmt.Sprintf("%d B", b)
// 	}
// 	div, exp := int64(unit), 0
// 	for n := b / unit; n >= unit; n /= unit {
// 		div *= unit
// 		exp++
// 	}
// 	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
// }

// func Calculate(request int, limit int) int {

// 	return request + limit
// }
