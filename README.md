# Fake Traffic Generator (Go)

A tool to generate fake HTTP traffic by simulating users browsing the internet. This can be used for testing network traffic handling, analyzing traffic patterns, or benchmarking systems that process network traffic.

## Features

- Configurable traffic volume (requests per second)
- Adjustable number of concurrent users
- Simulated realistic browsing behavior
- Random IP spoofing to prevent caching
- Realistic user-agent and HTTP headers
- Supports reading URLs from a text file
- Manageable via command-line flags or config file

## Requirements

- Go 1.16 or higher

## Installation

Clone the repository and build the executable:

```bash
git clone https://github.com/your-username/fake-traffic-go.git
cd fake-traffic-go
go build
```

## Usage

### Basic Usage

```bash
# Start with 10 concurrent users
./fake-traffic-go

# Start with 100 concurrent users
./fake-traffic-go -users 100

# Start with 50 concurrent users and target 200 requests per second
./fake-traffic-go -users 50 -rps 200
```

### Command Line Options

```
  -config string
        Path to configuration file
  -create-sample
        Create a sample URL file if none exists
  -ip-end string
        End of IP range (default "192.168.1.254")
  -ip-start string
        Start of IP range (default "192.168.1.1")
  -rps int
        Target requests per second (default 50)
  -urls string
        Path to URL list file (default "urls/urls.txt")
  -users int
        Number of concurrent users (default 10)
```

### URL File Format

The URL file should contain one URL per line. For example:

```
https://www.example.com
https://www.google.com
https://www.github.com
```

You can create a sample URL file using the `-create-sample` flag.

## Configuration File

You can use a JSON configuration file instead of command-line arguments. Create a file like this:

```json
{
  "concurrent_users": 20,
  "requests_per_second": 100,
  "url_file_path": "urls/custom-urls.txt",
  "page_change_interval": 2.5,
  "ip_range_start": "10.0.0.1",
  "ip_range_end": "10.0.0.254",
  "enabled": true
}
```

Then use it with:

```bash
./fake-traffic-go -config config.json
```

## Note on IP Spoofing

The IP spoofing implementation in this tool is simulated and doesn't actually modify the network packets' source IP address at the OS level. In a real-world scenario, you would need root/admin privileges and additional OS-specific configuration to truly spoof source IPs.

The current implementation primarily serves as a demonstration of how such a system would be designed.

## License

MIT 