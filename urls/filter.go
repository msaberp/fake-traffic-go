package urls

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

// FilterOptions configures the URL filtering process
type FilterOptions struct {
	// Maximum time to wait when checking URL reachability (in seconds)
	Timeout int

	// Number of concurrent workers for checking URLs
	Workers int

	// Whether to check if the URL is reachable (makes actual HTTP requests)
	CheckReachability bool

	// Whether to validate URL syntax
	ValidateURL bool

	// Domains to exclude (e.g., "example.com")
	ExcludeDomains []string

	// Protocols to allow (e.g., "https")
	AllowProtocols []string
}

// DefaultFilterOptions returns sensible defaults for filtering
func DefaultFilterOptions() FilterOptions {
	return FilterOptions{
		Timeout:           5,
		Workers:           20,
		CheckReachability: true,
		ValidateURL:       true,
		ExcludeDomains:    []string{},
		AllowProtocols:    []string{"http", "https"},
	}
}

// FilterURLsFile reads, filters, and writes back a list of valid URLs
func FilterURLsFile(inputPath, outputPath string, options FilterOptions) (int, int, error) {
	// Read all URLs from file
	file, err := os.Open(inputPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url != "" {
			urls = append(urls, url)
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, fmt.Errorf("error reading input file: %w", err)
	}

	totalURLs := len(urls)
	fmt.Printf("Read %d URLs from %s\n", totalURLs, inputPath)

	// Filter the URLs
	validURLs, err := FilterURLs(urls, options)
	if err != nil {
		return 0, 0, fmt.Errorf("error filtering URLs: %w", err)
	}

	// Write filtered URLs back to file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	for _, u := range validURLs {
		if _, err := writer.WriteString(u + "\n"); err != nil {
			return 0, 0, fmt.Errorf("error writing to output file: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return 0, 0, fmt.Errorf("error flushing writer: %w", err)
	}

	validCount := len(validURLs)
	fmt.Printf("Filtered %d/%d URLs (%.1f%% removed)\n",
		validCount, totalURLs, 100.0-float64(validCount)/float64(totalURLs)*100.0)

	return totalURLs, validCount, nil
}

// FilterURLs processes a slice of URLs and returns only valid ones
func FilterURLs(urls []string, options FilterOptions) ([]string, error) {
	var validURLs []string
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// Create a channel for URLs to process
	urlChan := make(chan string)

	// Set up workers
	for i := 0; i < options.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Create an HTTP client with timeout
			client := &http.Client{
				Timeout: time.Duration(options.Timeout) * time.Second,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse // Don't follow redirects
				},
			}

			for urlStr := range urlChan {
				valid := true
				var reason string

				// Validate URL syntax
				if options.ValidateURL {
					parsedURL, err := url.Parse(urlStr)
					if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
						valid = false
						reason = "invalid URL format"
						continue
					}

					// Check protocol
					if len(options.AllowProtocols) > 0 {
						if !slices.Contains(options.AllowProtocols, parsedURL.Scheme) {
							valid = false
							reason = "protocol not allowed"
							continue
						}
					}

					// Check excluded domains
					isDomainExcluded := func(host string, excluded []string) bool {
						for _, domain := range excluded {
							if strings.Contains(host, domain) {
								return true
							}
						}
						return false
					}

					if isDomainExcluded(parsedURL.Host, options.ExcludeDomains) {
						valid = false
						reason = "domain excluded"
						continue
					}
				}

				// Check reachability
				if valid && options.CheckReachability {
					ctx, cancel := context.WithTimeout(context.Background(), time.Duration(options.Timeout)*time.Second)
					req, err := http.NewRequestWithContext(ctx, "HEAD", urlStr, nil)
					if err != nil {
						valid = false
						reason = "failed to create request"
						cancel()
						continue
					}

					// Add a user agent to avoid being blocked
					req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

					resp, err := client.Do(req)
					cancel()

					if err != nil {
						valid = false
						reason = "unreachable"
						continue
					}

					resp.Body.Close()

					// Consider non-success status codes as invalid
					if resp.StatusCode < 200 || resp.StatusCode >= 400 {
						valid = false
						reason = fmt.Sprintf("status code %d", resp.StatusCode)
					}
				}

				if valid {
					mutex.Lock()
					validURLs = append(validURLs, urlStr)
					mutex.Unlock()
				} else {
					fmt.Printf("Filtered out %s: %s\n", urlStr, reason)
				}
			}
		}()
	}

	// Send URLs to workers
	go func() {
		for _, u := range urls {
			urlChan <- u
		}
		close(urlChan)
	}()

	// Wait for all workers to finish
	wg.Wait()

	return validURLs, nil
}

// BuildFilterOptions creates a FilterOptions with custom settings
func BuildFilterOptions(timeout, workers int, checkReachability, validateURL bool,
	excludeDomains, allowProtocols []string) FilterOptions {

	return FilterOptions{
		Timeout:           timeout,
		Workers:           workers,
		CheckReachability: checkReachability,
		ValidateURL:       validateURL,
		ExcludeDomains:    excludeDomains,
		AllowProtocols:    allowProtocols,
	}
}
