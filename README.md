# Actor Model Observability System

This is my thesis project where I built a ride-hailing system to compare two different ways of monitoring applications:

1. **Actor Model**: Each component (passenger, driver, trip) is an independent "actor" that monitors itself
2. **Traditional**: Standard REST API with external monitoring tools like Prometheus and Grafana

The goal is to see which approach works better for observability - basically, how well we can understand what's happening in our system when things go wrong.

## What's Inside

The system has actors for:
- **Passengers** - handle ride requests
- **Drivers** - manage availability and location
- **Trips** - track the actual rides
- **Matching** - connect passengers with drivers

For comparison, there's also a traditional REST API version that does the same thing but uses external tools for monitoring.

### Running It

For actor model:
```bash
# Set SYSTEM_MODE=actor in .env
go run cmd/main.go
```

For traditional approach:
```bash
# Set SYSTEM_MODE=traditional in .env
go run cmd/main.go
```

## Testing

Run tests:
```bash
go test ./...
```

Run benchmarks to compare actor vs traditional:
```bash
make bench-comparison
```

Load testing:
```bash
go run cmd/load-test/main.go -mode=actor -users=100 -duration=5m
go run cmd/load-test/main.go -mode=traditional -users=100 -duration=5m
```

## Monitoring

The whole point of this project is comparing how well we can monitor these two approaches:

**Actor Model**: Each actor monitors itself and reports what it's doing. You can see everything at:
- http://localhost:8080/api/metrics
- http://localhost:8080/api/health

**Traditional**: Uses external tools like Prometheus and Grafana for monitoring.

Both approaches track things like response times, error rates, and system performance, but they do it differently.

## What I'm Comparing

I'm measuring which approach is better at:

- **Performance**: How fast can each handle requests? How much memory do they use?
- **Monitoring Quality**: Which one gives better insights when something breaks?
- **Scalability**: How well do they handle increasing load?

To run the comparison:
```bash
go run scripts/benchmark.go
```

## Docker (Optional)

If you prefer using Docker:
```bash
docker-compose up -d
```

This starts everything you need: the app, PostgreSQL, Redis, and monitoring tools.

## Development Notes

If you want to add new actors:
1. Implement the `Actor` interface in `internal/actor/`
2. Add message handlers
3. Wire it up in the actor system

The code follows standard Go conventions. Tests are in the `tests/` folder.
