package internal

import (
	"fmt"
	"net/http"
	"time"
)

// HTTPClient wraps an http.Client with additional functionality
type HTTPClient struct {
	client          *http.Client
	userAgent       string
	requestCallback func() // Function to call when a request is made
}

// NewHTTPClient creates a new HTTP client with optional request callback
func NewHTTPClient(callback func()) *HTTPClient {
	client := &http.Client{
		Timeout: 10 * time.Second,
		// We don't follow redirects automatically as we want to simulate
		// user interaction for each navigation step
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &HTTPClient{
		client:          client,
		userAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		requestCallback: callback,
	}
}

// SetUserAgent sets the User-Agent header for all requests
func (c *HTTPClient) SetUserAgent(userAgent string) {
	c.userAgent = userAgent
}

// Get makes an HTTP GET request to the specified URL
func (c *HTTPClient) Get(url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set common headers to make the request look realistic
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Cache-Control", "max-age=0")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	// Log the response status
	fmt.Printf("Response status: %s\n", resp.Status)

	// Call the request callback if provided
	if c.requestCallback != nil {
		c.requestCallback()
	}

	return nil
}

// Post makes an HTTP POST request to the specified URL with form data
func (c *HTTPClient) Post(url string, contentType string, body []byte) error {
	// Implementation similar to Get but with POST method
	// This would be used for forms and login simulations
	// Left as an exercise or for future implementation
	return nil
}
