package urls

import (
	"bufio"
	"math/rand"
	"os"
	"sync"
	"time"
)

// URLManager manages a list of URLs to be used for traffic generation
type URLManager struct {
	urls []string
	mu   sync.RWMutex
	rand *rand.Rand
}

// NewURLManager creates a new URL manager
func NewURLManager() *URLManager {
	source := rand.NewSource(time.Now().UnixNano())
	return &URLManager{
		urls: make([]string, 0),
		rand: rand.New(source),
	}
}

// LoadFromFile reads URLs from a file (one URL per line)
func (m *URLManager) LoadFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := scanner.Text()
		if url != "" {
			urls = append(urls, url)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	m.urls = urls
	m.mu.Unlock()

	return nil
}

// GetRandomURL returns a random URL from the loaded list
func (m *URLManager) GetRandomURL() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.urls) == 0 {
		return "https://example.com"
	}

	index := m.rand.Intn(len(m.urls))
	return m.urls[index]
}

// Count returns the number of loaded URLs
func (m *URLManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.urls)
}

// CreateSampleURLFile creates a sample URL file if none exists
func CreateSampleURLFile(filePath string) error {
	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return nil // File exists, no need to create
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write sample URLs
	sampleURLs := []string{
		"https://www.example.com",
		"https://www.google.com",
		"https://www.github.com",
		"https://www.reddit.com",
		"https://news.ycombinator.com",
		"https://www.wikipedia.org",
		"https://www.stackoverflow.com",
		"https://www.amazon.com",
		"https://www.nytimes.com",
		"https://www.cnn.com",
	}

	writer := bufio.NewWriter(file)
	for _, url := range sampleURLs {
		_, err := writer.WriteString(url + "\n")
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}
