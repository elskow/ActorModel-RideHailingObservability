# Load Testing Framework

This document describes the comprehensive load testing framework for comparing the actor model-based observability system against traditional monitoring approaches.

## Overview

The load testing framework provides multiple ways to benchmark and compare the performance characteristics of both approaches:

1. **Go Benchmark Tests** - Micro-benchmarks for specific components
2. **Load Testing Application** - HTTP load testing with configurable parameters
3. **Benchmark Comparison Script** - Automated comparison and reporting
4. **Shell Script Framework** - Complete end-to-end testing pipeline

## Quick Start

### Prerequisites

```bash
# Install load testing tools
make install-load-tools

# Build the application and load test tools
make build
make build-load-test
```

### Running Tests

#### 1. Quick Benchmark Comparison

```bash
# Run Go benchmarks comparing both approaches
make bench-comparison
```

#### 2. Comprehensive Load Testing

```bash
# Run the full load testing framework
make load-test
```

This will:
- Build both applications
- Start actor model and traditional servers
- Run load tests at multiple concurrency levels
- Generate a comprehensive comparison report

#### 3. Specific Benchmark Tests

```bash
# Actor model benchmarks only
make bench-actor

# Traditional approach benchmarks only
make bench-traditional

# Scalability testing
make bench-scalability

# Memory usage analysis
make bench-memory

# Observability overhead analysis
make bench-overhead
```

#### 4. Quick Load Test

```bash
# Simple load test (requires running server)
make load-test-quick
```

## Available Benchmark Tests

### Go Benchmark Tests (`tests/benchmark_test.go`)

1. **BenchmarkActorRideRequest** - Actor model HTTP request performance
2. **BenchmarkTraditionalRideRequest** - Traditional HTTP request performance
3. **BenchmarkActorSystemMessagePassing** - Actor message passing performance
4. **BenchmarkObservabilityOverhead** - Actor observability collection overhead
5. **BenchmarkTraditionalMonitoringOverhead** - Traditional monitoring overhead
6. **BenchmarkConcurrentActorOperations** - Concurrent actor operations
7. **BenchmarkMemoryUsage** - Memory usage patterns
8. **BenchmarkScalability** - Scalability characteristics at different user loads

### Load Testing Application (`cmd/load-test/main.go`)

Configurable HTTP load testing tool with:
- Adjustable concurrency levels
- Configurable test duration
- Multiple endpoint testing
- JSON result output
- Real-time metrics collection

### Benchmark Comparison Script (`scripts/benchmark.go`)

Automated comparison tool that:
- Runs both approaches with identical parameters
- Collects performance metrics
- Calculates comparison statistics
- Generates detailed reports

### Shell Script Framework (`scripts/run_load_test.sh`)

Complete testing pipeline that:
- Builds applications
- Starts servers
- Runs tests at multiple concurrency levels
- Generates comprehensive reports
- Handles cleanup automatically

## Configuration

### Load Test Parameters

Edit `scripts/run_load_test.sh` to modify:

```bash
# Test duration per concurrency level
LOAD_TEST_DURATION=60s

# Concurrency levels to test
CONCURRENCY_LEVELS=(1 5 10 25 50 100)

# Server ports
SERVER_PORT=8080
TRADITIONAL_PORT=8081
```

### Benchmark Configuration

Modify benchmark parameters in `tests/benchmark_test.go`:

```go
// Number of actors for concurrent testing
numActors := 100

// User counts for scalability testing
userCounts := []int{1, 10, 50, 100, 500}
```

## Understanding Results

### Go Benchmark Output

```
BenchmarkActorRideRequest-8         	    1000	   1234567 ns/op	    1024 B/op	      10 allocs/op
BenchmarkTraditionalRideRequest-8  	    1200	   1000000 ns/op	     512 B/op	       5 allocs/op
```

- **First number**: Number of iterations
- **ns/op**: Nanoseconds per operation (lower is better)
- **B/op**: Bytes allocated per operation (lower is better)
- **allocs/op**: Number of allocations per operation (lower is better)

### Load Test Report

The comprehensive report includes:

1. **Performance Metrics**
   - Requests per second
   - Average latency
   - Error rates
   - Throughput comparison

2. **Scalability Analysis**
   - Performance at different concurrency levels
   - Resource utilization patterns
   - Bottleneck identification

3. **Recommendations**
   - When to use each approach
   - Performance trade-offs
   - Optimization suggestions

## Interpreting Performance Differences

### Actor Model Advantages

- **Isolation**: Better fault tolerance and error handling
- **Concurrency**: Superior handling of concurrent operations
- **Observability**: More granular monitoring and tracing
- **Scalability**: Better performance under high load

### Traditional Approach Advantages

- **Simplicity**: Lower complexity and overhead
- **Latency**: Potentially lower latency for simple operations
- **Resource Usage**: Lower memory footprint for basic scenarios
- **Familiarity**: Established patterns and tooling

## Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure ports 8080 and 8081 are available
2. **Missing tools**: Run `make install-load-tools` to install dependencies
3. **Build failures**: Check Go version and dependencies with `make deps`
4. **Permission errors**: Ensure script is executable: `chmod +x scripts/run_load_test.sh`

### Debug Mode

Run individual components for debugging:

```bash
# Build and run actor server manually
go build -o server ./cmd
./server --port=8080 --mode=actor

# Run load test tool manually
go build -o load-test ./cmd/load-test
./load-test --url=http://localhost:8080/api/v1/rides/request --concurrency=10

# Run specific benchmark
go test -bench=BenchmarkActorRideRequest -v ./tests
```

## Extending the Framework

### Adding New Benchmarks

1. Add benchmark functions to `tests/benchmark_test.go`
2. Follow Go benchmark naming convention: `BenchmarkXxx`
3. Use `testing.B` parameter for timing control
4. Add corresponding Makefile targets

### Custom Load Tests

1. Modify `cmd/load-test/main.go` for new endpoints
2. Update `scripts/run_load_test.sh` for new test scenarios
3. Extend report generation in `scripts/benchmark.go`

### Additional Metrics

1. Add new metrics to observability collectors
2. Update benchmark tests to capture new metrics
3. Extend report templates to include new data

## Best Practices

1. **Consistent Environment**: Run tests on the same hardware/environment
2. **Multiple Runs**: Average results across multiple test runs
3. **Warm-up**: Allow applications to warm up before measuring
4. **Resource Monitoring**: Monitor CPU, memory, and network during tests
5. **Baseline Comparison**: Establish baseline performance metrics

## Report Generation

Reports are generated in the `./reports` directory with timestamps:

- `load_test_report_YYYYMMDD_HHMMSS.md` - Main comparison report
- `actor_server_YYYYMMDD_HHMMSS.log` - Actor server logs
- `traditional_server_YYYYMMDD_HHMMSS.log` - Traditional server logs
- `go_benchmarks_YYYYMMDD_HHMMSS.txt` - Go benchmark results
- `*_cN_YYYYMMDD_HHMMSS.json` - Load test results by concurrency level

## Contributing

When adding new tests or modifying existing ones:

1. Ensure tests are deterministic and repeatable
2. Document new configuration options
3. Update this README with new features
4. Add appropriate error handling and logging
5. Test on multiple environments when possible