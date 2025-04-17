package config

import (
	"encoding/json"
	"os"
	"sync"
)

// Config represents the application configuration
type Config struct {
	// Number of concurrent users/clients
	ConcurrentUsers int `json:"concurrent_users"`

	// Target requests per second
	RequestsPerSecond int `json:"requests_per_second"`

	// URL file path
	URLFilePath string `json:"url_file_path"`

	// Rate at which to change pages (seconds)
	PageChangeInterval float64 `json:"page_change_interval"`

	// IP range to simulate traffic from
	IPRangeStart string `json:"ip_range_start"`
	IPRangeEnd   string `json:"ip_range_end"`

	// Enable/disable traffic
	Enabled bool `json:"enabled"`

	// Internal mutex for safe concurrent updates
	mu sync.RWMutex `json:"-"`
}

// Default configuration values
var DefaultConfig = &Config{
	ConcurrentUsers:    10,
	RequestsPerSecond:  50,
	URLFilePath:        "urls/urls.txt",
	PageChangeInterval: 2.0,
	IPRangeStart:       "192.168.1.1",
	IPRangeEnd:         "192.168.1.254",
	Enabled:            true,
}

// LoadFromFile loads configuration from a JSON file
func (c *Config) LoadFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return json.Unmarshal(data, c)
}

// SaveToFile saves current configuration to a JSON file
func (c *Config) SaveToFile(filePath string) error {
	c.mu.RLock()
	data, err := json.MarshalIndent(c, "", "  ")
	c.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// SetConcurrentUsers safely updates the number of concurrent users
func (c *Config) SetConcurrentUsers(num int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ConcurrentUsers = num
}

// SetRequestsPerSecond safely updates the target RPS
func (c *Config) SetRequestsPerSecond(rps int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.RequestsPerSecond = rps
}

// SetEnabled enables or disables traffic generation
func (c *Config) SetEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Enabled = enabled
}

// GetConcurrentUsers safely retrieves the concurrent users setting
func (c *Config) GetConcurrentUsers() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ConcurrentUsers
}

// GetRequestsPerSecond safely retrieves the target RPS
func (c *Config) GetRequestsPerSecond() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RequestsPerSecond
}

// IsEnabled safely checks if traffic generation is enabled
func (c *Config) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Enabled
}
