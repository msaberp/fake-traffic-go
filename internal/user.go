package internal

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"fake-traffic-go/ipspoof"
	"fake-traffic-go/urls"
)

// BrowserUser represents a simulated user browsing the web
type BrowserUser struct {
	ID          int
	UserAgent   string
	SourceIP    string
	sessionTime float64
	thinkTime   float64
	urlManager  *urls.URLManager
	client      *HTTPClient
	stopChan    chan struct{}
	wg          *sync.WaitGroup
	rand        *rand.Rand
}

// NewBrowserUser creates a new simulated browser user
func NewBrowserUser(id int, urlManager *urls.URLManager, ipspoofer *ipspoof.IPSpoofer, wg *sync.WaitGroup, generator *TrafficGenerator) *BrowserUser {
	source := rand.NewSource(time.Now().UnixNano() + int64(id))
	r := rand.New(source)

	// Generate random think time (interval between page views) between 1-5 seconds
	thinkTime := 1.0 + r.Float64()*4.0

	// Generate random session time between 10-30 minutes
	sessionTime := 10.0 + r.Float64()*20.0

	// Create a callback function that records requests in the generator
	var requestCallback func()
	if generator != nil {
		requestCallback = generator.RecordRequest
	}

	return &BrowserUser{
		ID:          id,
		UserAgent:   ipspoof.GenerateRandomUserAgent(),
		SourceIP:    ipspoofer.GetRandomIP(),
		sessionTime: sessionTime,
		thinkTime:   thinkTime,
		urlManager:  urlManager,
		client:      NewHTTPClient(requestCallback),
		stopChan:    make(chan struct{}),
		wg:          wg,
		rand:        r,
	}
}

// Start begins the user's browsing session
func (u *BrowserUser) Start() {
	u.wg.Add(1)
	go func() {
		defer u.wg.Done()

		fmt.Printf("User %d started with IP %s and think time %.2fs\n",
			u.ID, u.SourceIP, u.thinkTime)

		// Set up client with our spoofed IP and user agent
		u.client.SetUserAgent(u.UserAgent)
		ipspoof.SetTransport(u.SourceIP)

		startTime := time.Now()
		sessionDuration := time.Duration(u.sessionTime * float64(time.Minute))

		for {
			select {
			case <-u.stopChan:
				fmt.Printf("User %d stopped\n", u.ID)
				return
			default:
				// Check if session time exceeded
				if time.Since(startTime) > sessionDuration {
					fmt.Printf("User %d session time exceeded\n", u.ID)
					return
				}

				// Get a random URL to "browse" to
				url := u.urlManager.GetRandomURL()

				// Make the request
				err := u.client.Get(url)
				if err != nil {
					fmt.Printf("User %d error requesting %s: %v\n", u.ID, url, err)
				} else {
					fmt.Printf("User %d visited %s\n", u.ID, url)
				}

				// Calculate think time with some randomness
				jitter := u.thinkTime * (0.5 + u.rand.Float64())
				thinkDuration := time.Duration(jitter * float64(time.Second))

				// Wait the think time before next request
				select {
				case <-u.stopChan:
					return
				case <-time.After(thinkDuration):
					// Continue to next URL
				}
			}
		}
	}()
}

// Stop halts the user's browsing session
func (u *BrowserUser) Stop() {
	close(u.stopChan)
}

// SimulatePageNavigation simulates a user clicking links and browsing around a site
// This is called internally by the browser session
func (u *BrowserUser) SimulatePageNavigation(baseURL string) []string {
	// Simulate clicking 1-5 links on the page
	numLinks := 1 + u.rand.Intn(5)
	visitedURLs := make([]string, 0, numLinks)

	// Add base URL as the first visited page
	visitedURLs = append(visitedURLs, baseURL)

	// This is a simplified simulation - in reality would parse the page and follow actual links
	for i := 0; i < numLinks; i++ {
		// Simulate a user clicking a link or navigating to a new path
		subpaths := []string{
			"/about",
			"/contact",
			"/products",
			"/services",
			"/blog",
			"/news",
			"/faq",
			"/login",
			"/register",
		}

		path := subpaths[u.rand.Intn(len(subpaths))]
		newURL := fmt.Sprintf("%s%s", baseURL, path)
		visitedURLs = append(visitedURLs, newURL)
	}

	return visitedURLs
}
