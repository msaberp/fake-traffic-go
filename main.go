package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fake-traffic-go/config"
	"fake-traffic-go/internal"
	"fake-traffic-go/urls"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "", "Path to configuration file")
	users := flag.Int("users", 10, "Number of concurrent users")
	rps := flag.Int("rps", 50, "Target requests per second")
	urlFile := flag.String("urls", "urls/urls.txt", "Path to URL list file")
	createSample := flag.Bool("create-sample", false, "Create a sample URL file if none exists")
	ipStart := flag.String("ip-start", "192.168.1.1", "Start of IP range")
	ipEnd := flag.String("ip-end", "192.168.1.254", "End of IP range")

	flag.Parse()

	// Create config
	cfg := config.DefaultConfig

	// Load from file if specified
	if *configFile != "" {
		err := cfg.LoadFromFile(*configFile)
		if err != nil {
			fmt.Printf("Warning: Failed to load config file: %v\n", err)
		} else {
			fmt.Printf("Loaded configuration from %s\n", *configFile)
		}
	}

	// Override with command line arguments if they were provided
	// We check against default values to determine if flags were explicitly set
	if *users != 10 {
		cfg.SetConcurrentUsers(*users)
	}
	if *rps != 50 {
		cfg.SetRequestsPerSecond(*rps)
	}
	if *urlFile != "urls/urls.txt" {
		cfg.URLFilePath = *urlFile
	}
	if *ipStart != "192.168.1.1" {
		cfg.IPRangeStart = *ipStart
	}
	if *ipEnd != "192.168.1.254" {
		cfg.IPRangeEnd = *ipEnd
	}

	// Create URL sample file if requested and needed
	if *createSample {
		err := urls.CreateSampleURLFile(cfg.URLFilePath)
		if err != nil {
			fmt.Printf("Error creating sample URL file: %v\n", err)
		} else {
			fmt.Printf("Created sample URL file at: %s\n", cfg.URLFilePath)
		}
	}

	// Create and start traffic generator
	generator, err := internal.NewTrafficGenerator(cfg)
	if err != nil {
		fmt.Printf("Error initializing traffic generator: %v\n", err)
		os.Exit(1)
	}

	err = generator.Start()
	if err != nil {
		fmt.Printf("Error starting traffic generator: %v\n", err)
		os.Exit(1)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Fake traffic generator running. Press Ctrl+C to stop.")

	// Periodically print statistics
	statsTicker := time.NewTicker(5 * time.Second)
	defer statsTicker.Stop()

	// Main loop
	for {
		select {
		case <-sigChan:
			fmt.Println("\nReceived shutdown signal")
			generator.Stop()
			return

		case <-statsTicker.C:
			// Print current statistics
			stats := generator.GetStats()
			statsJSON, _ := json.MarshalIndent(stats, "", "  ")
			fmt.Println("Traffic Generator Stats:")
			fmt.Println(string(statsJSON))
		}
	}
}
