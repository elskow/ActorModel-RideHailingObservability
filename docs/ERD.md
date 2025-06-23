# Entity Relationship Diagram (ERD)
# Actor Model Observability - Ride-Hailing System

## Database Schema Design

### 1. Core Business Entities

#### 1.1 Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    user_type VARCHAR(20) NOT NULL CHECK (user_type IN ('passenger', 'driver')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 1.2 Drivers Table
```sql
CREATE TABLE drivers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    license_number VARCHAR(50) UNIQUE NOT NULL,
    vehicle_type VARCHAR(50) NOT NULL,
    vehicle_plate VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'offline' CHECK (status IN ('online', 'offline', 'busy')),
    current_latitude DECIMAL(10, 8),
    current_longitude DECIMAL(11, 8),
    rating DECIMAL(3, 2) DEFAULT 5.00,
    total_trips INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 1.3 Passengers Table
```sql
CREATE TABLE passengers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating DECIMAL(3, 2) DEFAULT 5.00,
    total_trips INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 1.4 Trips Table
```sql
CREATE TABLE trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    passenger_id UUID NOT NULL REFERENCES passengers(id),
    driver_id UUID REFERENCES drivers(id),
    pickup_latitude DECIMAL(10, 8) NOT NULL,
    pickup_longitude DECIMAL(11, 8) NOT NULL,
    pickup_address TEXT,
    destination_latitude DECIMAL(10, 8) NOT NULL,
    destination_longitude DECIMAL(11, 8) NOT NULL,
    destination_address TEXT,
    status VARCHAR(20) DEFAULT 'requested' CHECK (status IN (
        'requested', 'matched', 'accepted', 'driver_arrived', 
        'in_progress', 'completed', 'cancelled'
    )),
    fare_amount DECIMAL(10, 2),
    distance_km DECIMAL(8, 2),
    duration_minutes INTEGER,
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    matched_at TIMESTAMP,
    accepted_at TIMESTAMP,
    pickup_at TIMESTAMP,
    completed_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 2. Observability Entities

#### 2.1 Actor Instances Table
```sql
CREATE TABLE actor_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_type VARCHAR(50) NOT NULL CHECK (actor_type IN (
        'passenger', 'driver', 'trip', 'matching', 'observability'
    )),
    actor_id VARCHAR(255) NOT NULL,
    entity_id UUID, -- References the business entity (user_id, trip_id, etc.)
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'error')),
    last_heartbeat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(actor_type, actor_id)
);
```

#### 2.2 Actor Messages Table
```sql
CREATE TABLE actor_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trace_id UUID NOT NULL,
    span_id UUID NOT NULL,
    parent_span_id UUID,
    sender_actor_type VARCHAR(50) NOT NULL,
    sender_actor_id VARCHAR(255) NOT NULL,
    receiver_actor_type VARCHAR(50) NOT NULL,
    receiver_actor_id VARCHAR(255) NOT NULL,
    message_type VARCHAR(100) NOT NULL,
    message_payload JSONB,
    status VARCHAR(20) DEFAULT 'sent' CHECK (status IN ('sent', 'received', 'processed', 'failed')),
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    received_at TIMESTAMP,
    processed_at TIMESTAMP,
    processing_duration_ms INTEGER,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 2.3 System Metrics Table
```sql
CREATE TABLE system_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_name VARCHAR(100) NOT NULL,
    metric_type VARCHAR(20) NOT NULL CHECK (metric_type IN ('counter', 'gauge', 'histogram')),
    metric_value DECIMAL(15, 6) NOT NULL,
    labels JSONB,
    actor_type VARCHAR(50),
    actor_id VARCHAR(255),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 2.4 Distributed Traces Table
```sql
CREATE TABLE distributed_traces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trace_id UUID NOT NULL,
    span_id UUID NOT NULL,
    parent_span_id UUID,
    operation_name VARCHAR(255) NOT NULL,
    actor_type VARCHAR(50),
    actor_id VARCHAR(255),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration_ms INTEGER,
    status VARCHAR(20) DEFAULT 'ok' CHECK (status IN ('ok', 'error', 'timeout')),
    tags JSONB,
    logs JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 2.5 Event Logs Table
```sql
CREATE TABLE event_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trace_id UUID,
    event_type VARCHAR(100) NOT NULL,
    event_category VARCHAR(50) NOT NULL CHECK (event_category IN (
        'business', 'system', 'error', 'performance', 'security'
    )),
    actor_type VARCHAR(50),
    actor_id VARCHAR(255),
    entity_type VARCHAR(50),
    entity_id UUID,
    event_data JSONB,
    severity VARCHAR(20) DEFAULT 'info' CHECK (severity IN ('debug', 'info', 'warn', 'error', 'fatal')),
    message TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 3. Indexes for Performance

```sql
-- Business entity indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_drivers_user_id ON drivers(user_id);
CREATE INDEX idx_drivers_status ON drivers(status);
CREATE INDEX idx_drivers_location ON drivers(current_latitude, current_longitude);
CREATE INDEX idx_passengers_user_id ON passengers(user_id);
CREATE INDEX idx_trips_passenger_id ON trips(passenger_id);
CREATE INDEX idx_trips_driver_id ON trips(driver_id);
CREATE INDEX idx_trips_status ON trips(status);
CREATE INDEX idx_trips_requested_at ON trips(requested_at);

-- Observability indexes
CREATE INDEX idx_actor_instances_type_id ON actor_instances(actor_type, actor_id);
CREATE INDEX idx_actor_instances_entity ON actor_instances(entity_id);
CREATE INDEX idx_actor_instances_status ON actor_instances(status);
CREATE INDEX idx_actor_instances_heartbeat ON actor_instances(last_heartbeat);

CREATE INDEX idx_actor_messages_trace ON actor_messages(trace_id);
CREATE INDEX idx_actor_messages_span ON actor_messages(span_id);
CREATE INDEX idx_actor_messages_sender ON actor_messages(sender_actor_type, sender_actor_id);
CREATE INDEX idx_actor_messages_receiver ON actor_messages(receiver_actor_type, receiver_actor_id);
CREATE INDEX idx_actor_messages_type ON actor_messages(message_type);
CREATE INDEX idx_actor_messages_sent_at ON actor_messages(sent_at);

CREATE INDEX idx_system_metrics_name ON system_metrics(metric_name);
CREATE INDEX idx_system_metrics_actor ON system_metrics(actor_type, actor_id);
CREATE INDEX idx_system_metrics_timestamp ON system_metrics(timestamp);

CREATE INDEX idx_distributed_traces_trace_id ON distributed_traces(trace_id);
CREATE INDEX idx_distributed_traces_span_id ON distributed_traces(span_id);
CREATE INDEX idx_distributed_traces_parent ON distributed_traces(parent_span_id);
CREATE INDEX idx_distributed_traces_actor ON distributed_traces(actor_type, actor_id);
CREATE INDEX idx_distributed_traces_start_time ON distributed_traces(start_time);

CREATE INDEX idx_event_logs_trace_id ON event_logs(trace_id);
CREATE INDEX idx_event_logs_type ON event_logs(event_type);
CREATE INDEX idx_event_logs_category ON event_logs(event_category);
CREATE INDEX idx_event_logs_actor ON event_logs(actor_type, actor_id);
CREATE INDEX idx_event_logs_entity ON event_logs(entity_type, entity_id);
CREATE INDEX idx_event_logs_timestamp ON event_logs(timestamp);
CREATE INDEX idx_event_logs_severity ON event_logs(severity);
```

### 4. Entity Relationships

#### 4.1 Core Business Relationships
- **Users** (1) → (0..1) **Drivers**: One user can be a driver
- **Users** (1) → (0..1) **Passengers**: One user can be a passenger
- **Passengers** (1) → (0..*) **Trips**: One passenger can have many trips
- **Drivers** (1) → (0..*) **Trips**: One driver can have many trips
- **Trips** (1) → (1) **Passengers**: Each trip belongs to one passenger
- **Trips** (1) → (0..1) **Drivers**: Each trip can be assigned to one driver

#### 4.2 Observability Relationships
- **Actor Instances** (1) → (0..*) **Actor Messages**: One actor can send/receive many messages
- **Actor Instances** (1) → (0..*) **System Metrics**: One actor can generate many metrics
- **Actor Instances** (1) → (0..*) **Distributed Traces**: One actor can participate in many traces
- **Actor Instances** (1) → (0..*) **Event Logs**: One actor can generate many events
- **Distributed Traces** (1) → (0..*) **Actor Messages**: One trace can contain many messages
- **Distributed Traces** (1) → (0..*) **Event Logs**: One trace can contain many events

### 5. Data Flow Patterns

#### 5.1 Business Data Flow
1. User registration creates entries in `users` and either `drivers` or `passengers`
2. Trip request creates entry in `trips` with status 'requested'
3. Matching process updates trip with driver assignment
4. Trip lifecycle updates trip status through various stages
5. Trip completion updates final metrics and ratings

#### 5.2 Observability Data Flow
1. Actor creation/destruction tracked in `actor_instances`
2. Message passing logged in `actor_messages` with trace context
3. Performance metrics collected in `system_metrics`
4. Distributed traces span across multiple actors in `distributed_traces`
5. Business and system events logged in `event_logs`

### 6. Data Retention Policies

- **Business Data**: Retain indefinitely for business analytics
- **Actor Messages**: Retain for 30 days (high volume)
- **System Metrics**: Retain for 90 days with aggregation
- **Distributed Traces**: Retain for 7 days (very high volume)
- **Event Logs**: Retain for 1 year with severity-based retention

### 7. Comparison Schema

For the traditional approach comparison, we'll use the same business tables but different observability tables:

#### 7.1 Traditional Monitoring Tables
```sql
CREATE TABLE traditional_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level VARCHAR(10) NOT NULL,
    message TEXT NOT NULL,
    service_name VARCHAR(100),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

CREATE TABLE traditional_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_name VARCHAR(100) NOT NULL,
    value DECIMAL(15, 6) NOT NULL,
    labels JSONB,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

This ERD provides a comprehensive foundation for both the actor model implementation and the traditional approach comparison, focusing on observability requirements while maintaining the core ride-hailing business logic.