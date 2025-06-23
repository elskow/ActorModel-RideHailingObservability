# Product Requirements Document (PRD)
# Actor Model for Observability in Ride-Hailing System

## 1. Project Overview

### 1.1 Purpose
This thesis project demonstrates the implementation and evaluation of an actor model-based observability system for a simplified ride-hailing platform. The focus is on comparing actor model observability patterns against traditional monitoring approaches.

### 1.2 Scope
A simplified ride-hailing system with core functionalities:
- Driver registration and status management
- Ride request and matching
- Trip tracking and completion
- Real-time observability and monitoring

### 1.3 Target Use Case
Ride-hailing platform similar to Uber/Gojek but simplified to focus on observability aspects:
- Passengers request rides
- Drivers accept/reject requests
- System matches passengers with drivers
- Real-time tracking of trips
- Comprehensive observability of all system interactions

## 2. Functional Requirements

### 2.1 Core Business Logic
- **FR-001**: Passenger can request a ride with pickup and destination
- **FR-002**: Driver can set availability status (online/offline)
- **FR-003**: System matches available drivers with ride requests
- **FR-004**: Driver can accept or reject ride requests
- **FR-005**: System tracks trip progress (pickup, in-progress, completed)
- **FR-006**: System calculates trip fare based on distance and time

### 2.2 Observability Requirements
- **FR-007**: Real-time monitoring of all actor interactions
- **FR-008**: Distributed tracing of ride request lifecycle
- **FR-009**: Metrics collection for system performance
- **FR-010**: Event logging for audit and debugging
- **FR-011**: Health monitoring of individual actors
- **FR-012**: System-wide observability dashboard

## 3. Non-Functional Requirements

### 3.1 Performance
- **NFR-001**: System should handle 100 concurrent ride requests
- **NFR-002**: Ride matching should complete within 5 seconds
- **NFR-003**: Observability overhead should not exceed 10% of system resources

### 3.2 Scalability
- **NFR-004**: Actor system should support horizontal scaling
- **NFR-005**: Observability system should scale with business logic

### 3.3 Reliability
- **NFR-006**: System should maintain 99% uptime
- **NFR-007**: Observability data should be persistent and recoverable

## 4. System Architecture

### 4.1 Actor Model Components
- **Passenger Actor**: Manages passenger state and ride requests
- **Driver Actor**: Manages driver state and availability
- **Trip Actor**: Manages individual trip lifecycle
- **Matching Actor**: Handles ride request matching logic
- **Observability Actor**: Collects and processes monitoring data

### 4.2 Observability Components
- **Metrics Collector**: Gathers performance metrics
- **Trace Collector**: Handles distributed tracing
- **Event Logger**: Manages event logging
- **Health Monitor**: Monitors actor health
- **Dashboard**: Visualizes observability data

## 5. Comparison Framework

### 5.1 Traditional Approach
- Centralized logging system
- External monitoring tools (Prometheus, Grafana)
- Database-based state management
- Synchronous API calls

### 5.2 Actor Model Approach
- Distributed actor-based observability
- Built-in monitoring capabilities
- Actor-based state management
- Asynchronous message passing

### 5.3 Comparison Metrics
- **Performance**: Response time, throughput, resource usage
- **Scalability**: Horizontal scaling capabilities
- **Maintainability**: Code complexity, debugging ease
- **Observability**: Monitoring coverage, trace completeness
- **Fault Tolerance**: Error handling, system recovery

## 6. Technology Stack

### 6.1 Development
- **Language**: Go
- **Actor Framework**: Custom implementation or Asynq/similar
- **Database**: PostgreSQL for persistence
- **Message Queue**: Redis for actor communication

### 6.2 Observability
- **Metrics**: Custom metrics collection
- **Tracing**: OpenTelemetry integration
- **Logging**: Structured logging with logrus
- **Visualization**: Web-based dashboard

## 7. Success Criteria

### 7.1 Functional Success
- All core ride-hailing functionalities working
- Complete observability coverage of actor interactions
- Successful comparison between traditional and actor approaches

### 7.2 Technical Success
- System handles target load (100 concurrent requests)
- Observability overhead within acceptable limits
- Clear performance and maintainability advantages demonstrated

### 7.3 Academic Success
- Comprehensive analysis of actor model benefits for observability
- Quantitative comparison with traditional approaches
- Clear recommendations for when to use actor model observability

## 8. Timeline

### Phase 1: Foundation (Weeks 1-2)
- Project setup and basic actor framework
- Core data models and database schema
- Basic ride-hailing logic implementation

### Phase 2: Actor Implementation (Weeks 3-4)
- Complete actor system implementation
- Actor-based observability framework
- Message passing and state management

### Phase 3: Traditional Implementation (Weeks 5-6)
- Traditional approach implementation
- External monitoring integration
- Comparison baseline establishment

### Phase 4: Testing and Analysis (Weeks 7-8)
- Performance testing and benchmarking
- Observability analysis and comparison
- Documentation and thesis writing

## 9. Risks and Mitigation

### 9.1 Technical Risks
- **Risk**: Actor framework complexity
- **Mitigation**: Start with simple implementation, iterate

### 9.2 Scope Risks
- **Risk**: Feature creep beyond observability focus
- **Mitigation**: Maintain strict scope boundaries, focus on core comparison

### 9.3 Timeline Risks
- **Risk**: Implementation taking longer than expected
- **Mitigation**: Prioritize core features, have fallback simplified scenarios