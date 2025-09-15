# Load Balancer in Go

A simple load balancer implemented in Golang

- Support HTTP (layer 7) and TCP (layer 4) protocols
- Continuous Health Checks with specific time intervals
- Request with retry logic
- Different Load Balancing Algorithms (round robin and weighted round robin for now)

## Configuration
Add configuration in root of the project in file `config.json`

- example config.json:
```json
{
  "protocol": "tcp",
  "port": 8080,
  "algorithm": "Weighted Round Robin",
  "healthCheckInterval": 10,
  "servers": [
    {
      "addr": "127.0.0.1:8081",
      "healthCheckHTTPEndpoint": "/health",
      "weight": 2
    },
    {
      "addr": "127.0.0.1:8082",
      "healthCheckHTTPEndpoint": "/health",
      "weight": 1
    },
    {
      "addr": "127.0.0.1:8083",
      "healthCheckHTTPEndpoint": "/health",
      "weight": 1
    }
  ]
}
```

- **Supported values:**
- `protocol`: `tcp` | `http`
- `algorithm`: `Weighted Round Robin` | `Round Robin`
- `healthCheckInterval`: seconds (int)
- `addr`: backend server address, format: `ip:port`
- `healthCheckHTTPEndpoint`: for http mode, health check endpoints url, can omit in tcp mode

## Project Setup
- clone repository
- Install dependencies:
`go mod tidy`
- Add your configurations for load balancer
- Start Load Balancer:
`go run ./cmd/lb`

### Test

- You can also use test backend servers for testing: `PORT=8081 go run ./tools/tb`
- And after running your test servers and load balancers, you can make multiple concurrent requests with curl (given script): `./scripts/test/make-curl-requests.sh <path_to_urls.txt>` 

> Add your urls.txt and give path to the file
> example `./urls.txt` at the root of the project
