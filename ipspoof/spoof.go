package ipspoof

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

// IPSpoofer handles IP address spoofing
type IPSpoofer struct {
	startIP net.IP
	endIP   net.IP
	mu      sync.Mutex
	rand    *rand.Rand
}

// NewIPSpoofer creates a new IP spoofer within the given range
func NewIPSpoofer(startIPStr string, endIPStr string) (*IPSpoofer, error) {
	startIP := net.ParseIP(startIPStr).To4()
	if startIP == nil {
		return nil, fmt.Errorf("invalid start IP address: %s", startIPStr)
	}

	endIP := net.ParseIP(endIPStr).To4()
	if endIP == nil {
		return nil, fmt.Errorf("invalid end IP address: %s", endIPStr)
	}

	// Ensure startIP <= endIP
	for i := 0; i < 4; i++ {
		if startIP[i] > endIP[i] {
			return nil, fmt.Errorf("start IP (%s) must be less than or equal to end IP (%s)", startIPStr, endIPStr)
		} else if startIP[i] < endIP[i] {
			break
		}
	}

	source := rand.NewSource(time.Now().UnixNano())
	return &IPSpoofer{
		startIP: startIP,
		endIP:   endIP,
		rand:    rand.New(source),
	}, nil
}

// GetRandomIP returns a random IP address within the configured range
func (s *IPSpoofer) GetRandomIP() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Convert IPs to uint32 for easier random generation
	startInt := ipToUint32(s.startIP)
	endInt := ipToUint32(s.endIP)

	// Generate random IP in range
	randomInt := startInt + uint32(s.rand.Int63n(int64(endInt-startInt+1)))
	randomIP := uint32ToIP(randomInt)

	return randomIP.String()
}

// Helper function to convert IP to uint32
func ipToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

// Helper function to convert uint32 to IP
func uint32ToIP(ipInt uint32) net.IP {
	return net.IPv4(
		byte(ipInt>>24),
		byte(ipInt>>16),
		byte(ipInt>>8),
		byte(ipInt),
	)
}

// SetTransport modifies the HTTP transport to use a specific source IP (requires root privileges)
// This is a placeholder - in a real implementation, this would use raw sockets or similar
// Note: This functionality is limited and might not work without proper OS/networking setup
func SetTransport(sourceIP string) error {
	// This is a placeholder. In a real implementation, this would:
	// 1. Create raw sockets or use platform-specific methods to spoof the source IP
	// 2. Set up proper routing and packet handling

	// For demonstration purposes, just log that we're using a specific IP
	fmt.Printf("Using source IP: %s\n", sourceIP)

	return nil
}

// GenerateRandomUserAgent generates a random user agent string
// This helps with making traffic look more realistic
func GenerateRandomUserAgent() string {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	browsers := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%d.0.%d.%d Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:%d.0) Gecko/20100101 Firefox/%d.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_%d_%d) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/%d.%d Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%d.0.%d.%d Edge/%d.%d.%d.%d",
	}

	browser := browsers[r.Intn(len(browsers))]

	switch {
	case strings.Contains(browser, "Chrome") && !strings.Contains(browser, "Edge"):
		return fmt.Sprintf(browser, 70+r.Intn(30), r.Intn(9999), r.Intn(999))
	case strings.Contains(browser, "Firefox"):
		return fmt.Sprintf(browser, 70+r.Intn(30), 70+r.Intn(30))
	case strings.Contains(browser, "Safari") && !strings.Contains(browser, "Chrome"):
		return fmt.Sprintf(browser, 10+r.Intn(5), r.Intn(9), 12+r.Intn(8), r.Intn(9))
	case strings.Contains(browser, "Edge"):
		return fmt.Sprintf(browser, 70+r.Intn(30), r.Intn(9999), r.Intn(999), 15+r.Intn(10), r.Intn(999), r.Intn(999), r.Intn(999))
	default:
		return fmt.Sprintf(browser, 70+r.Intn(30), r.Intn(9999), r.Intn(999))
	}
}
