package internal

import (
	"fmt"
	"sync"
	"time"

	"fake-traffic-go/config"
	"fake-traffic-go/ipspoof"
	"fake-traffic-go/urls"
)

// TrafficGenerator coordinates traffic generation
type TrafficGenerator struct {
	config        *config.Config
	urlManager    *urls.URLManager
	ipSpoofer     *ipspoof.IPSpoofer
	users         map[int]*BrowserUser
	usersMutex    sync.Mutex
	wg            sync.WaitGroup
	running       bool
	stopChan      chan struct{}
	requestCount  int64
	requestsMutex sync.Mutex
	requestsStart time.Time
}

// NewTrafficGenerator creates a new traffic generator
func NewTrafficGenerator(cfg *config.Config) (*TrafficGenerator, error) {
	// Create URL manager
	urlManager := urls.NewURLManager()
	err := urlManager.LoadFromFile(cfg.URLFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load URLs: %w", err)
	}

	// Create IP spoofer
	ipSpoofer, err := ipspoof.NewIPSpoofer(cfg.IPRangeStart, cfg.IPRangeEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to create IP spoofer: %w", err)
	}

	return &TrafficGenerator{
		config:        cfg,
		urlManager:    urlManager,
		ipSpoofer:     ipSpoofer,
		users:         make(map[int]*BrowserUser),
		stopChan:      make(chan struct{}),
		requestCount:  0,
		requestsStart: time.Now(),
	}, nil
}

// Start begins traffic generation
func (g *TrafficGenerator) Start() error {
	if g.running {
		return fmt.Errorf("traffic generator is already running")
	}

	g.running = true
	fmt.Println("Starting traffic generator...")

	// Start the user manager goroutine
	go g.manageUsers()

	return nil
}

// Stop halts traffic generation
func (g *TrafficGenerator) Stop() {
	if !g.running {
		return
	}

	fmt.Println("Stopping traffic generator...")
	close(g.stopChan)

	// Stop all users
	g.usersMutex.Lock()
	for _, user := range g.users {
		user.Stop()
	}
	g.usersMutex.Unlock()

	// Wait for all users to finish
	g.wg.Wait()

	g.running = false
	fmt.Println("Traffic generator stopped")
}

// manageUsers continuously adjusts the number of active users based on configuration
func (g *TrafficGenerator) manageUsers() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-g.stopChan:
			return
		case <-ticker.C:
			if !g.config.IsEnabled() {
				// Traffic generation disabled - stop all users
				g.adjustActiveUsers(0)
				continue
			}

			// Get current target for concurrent users
			targetUsers := g.config.GetConcurrentUsers()

			// Adjust number of active users
			g.adjustActiveUsers(targetUsers)
		}
	}
}

// adjustActiveUsers adds or removes users to match the target count
func (g *TrafficGenerator) adjustActiveUsers(targetCount int) {
	g.usersMutex.Lock()
	defer g.usersMutex.Unlock()

	currentCount := len(g.users)

	// Add users if needed
	if currentCount < targetCount {
		for i := currentCount; i < targetCount; i++ {
			user := NewBrowserUser(i, g.urlManager, g.ipSpoofer, &g.wg, g)
			g.users[i] = user
			user.Start()
		}
		fmt.Printf("Added %d users. Current user count: %d\n", targetCount-currentCount, targetCount)
	}

	// Remove users if needed
	if currentCount > targetCount {
		for i := currentCount - 1; i >= targetCount; i-- {
			if user, exists := g.users[i]; exists {
				user.Stop()
				delete(g.users, i)
			}
		}
		fmt.Printf("Removed %d users. Current user count: %d\n", currentCount-targetCount, targetCount)
	}
}

// RecordRequest increments the request counter
func (g *TrafficGenerator) RecordRequest() {
	g.requestsMutex.Lock()
	defer g.requestsMutex.Unlock()
	g.requestCount++
}

// GetActualRequestsPerSecond calculates the actual requests per second
func (g *TrafficGenerator) GetActualRequestsPerSecond() float64 {
	g.requestsMutex.Lock()
	defer g.requestsMutex.Unlock()

	elapsed := time.Since(g.requestsStart).Seconds()
	if elapsed < 1 {
		return 0 // Not enough time has passed for accurate measurement
	}

	rps := float64(g.requestCount) / elapsed

	// Reset counters every minute to avoid integer overflow and keep measurement recent
	if elapsed > 60 {
		g.requestCount = 0
		g.requestsStart = time.Now()
	}

	return rps
}

// GetStats returns statistics about the traffic generation
func (g *TrafficGenerator) GetStats() map[string]any {
	g.usersMutex.Lock()
	activeUsers := len(g.users)
	g.usersMutex.Unlock()

	return map[string]any{
		"active_users":            activeUsers,
		"target_users":            g.config.GetConcurrentUsers(),
		"target_requests_per_sec": g.config.GetRequestsPerSecond(),
		"actual_requests_per_sec": float64(int(g.GetActualRequestsPerSecond()*100)) / 100, // Round to 2 decimal places
		"url_count":               g.urlManager.Count(),
		"enabled":                 g.config.IsEnabled(),
	}
}
