-- Database initialization script for Actor Model Observability

-- Create database if it doesn't exist (this is handled by docker-compose)
-- CREATE DATABASE IF NOT EXISTS actor_observability;

-- Use the database
-- \c actor_observability;

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    name VARCHAR(255) NOT NULL,
    user_type VARCHAR(20) NOT NULL CHECK (user_type IN ('passenger', 'driver')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create drivers table
CREATE TABLE IF NOT EXISTS drivers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    license_number VARCHAR(50) UNIQUE NOT NULL,
    vehicle_type VARCHAR(50) NOT NULL,
    vehicle_plate VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'offline' CHECK (status IN ('online', 'offline', 'busy')),
    current_latitude DECIMAL(10, 8),
    current_longitude DECIMAL(11, 8),
    rating DECIMAL(3, 2) DEFAULT 0.00,
    total_trips INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create passengers table
CREATE TABLE IF NOT EXISTS passengers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating DECIMAL(3, 2) DEFAULT 0.00,
    total_trips INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create trips table
CREATE TABLE IF NOT EXISTS trips (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    passenger_id UUID NOT NULL REFERENCES passengers(id),
    driver_id UUID REFERENCES drivers(id),
    status VARCHAR(20) NOT NULL DEFAULT 'requested' CHECK (status IN ('requested', 'accepted', 'in_progress', 'completed', 'cancelled')),
    pickup_latitude DECIMAL(10, 8) NOT NULL,
    pickup_longitude DECIMAL(11, 8) NOT NULL,
    pickup_address TEXT,
    destination_latitude DECIMAL(10, 8) NOT NULL,
    destination_longitude DECIMAL(11, 8) NOT NULL,
    destination_address TEXT,
    estimated_fare DECIMAL(10, 2),
    actual_fare DECIMAL(10, 2),
    distance_km DECIMAL(8, 2),
    duration_minutes INTEGER,
    passenger_rating INTEGER CHECK (passenger_rating >= 1 AND passenger_rating <= 5),
    driver_rating INTEGER CHECK (driver_rating >= 1 AND driver_rating <= 5),
    requested_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    accepted_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create actor_instances table for observability
CREATE TABLE IF NOT EXISTS actor_instances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    actor_id VARCHAR(255) NOT NULL,
    actor_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    INDEX idx_actor_instances_actor_id (actor_id),
    INDEX idx_actor_instances_type (actor_type),
    INDEX idx_actor_instances_status (status),
    INDEX idx_actor_instances_created_at (created_at)
);

-- Create actor_messages table for observability
CREATE TABLE IF NOT EXISTS actor_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id VARCHAR(255) NOT NULL,
    sender_id VARCHAR(255),
    receiver_id VARCHAR(255) NOT NULL,
    message_type VARCHAR(100) NOT NULL,
    payload JSONB,
    status VARCHAR(50) NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE NOT NULL,
    received_at TIMESTAMP WITH TIME ZONE,
    processed_at TIMESTAMP WITH TIME ZONE,
    INDEX idx_actor_messages_message_id (message_id),
    INDEX idx_actor_messages_sender (sender_id),
    INDEX idx_actor_messages_receiver (receiver_id),
    INDEX idx_actor_messages_type (message_type),
    INDEX idx_actor_messages_sent_at (sent_at)
);

-- Create system_metrics table for observability
CREATE TABLE IF NOT EXISTS system_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    metric_name VARCHAR(255) NOT NULL,
    metric_value DECIMAL(15, 6) NOT NULL,
    metric_type VARCHAR(50) NOT NULL,
    labels JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    INDEX idx_system_metrics_name (metric_name),
    INDEX idx_system_metrics_type (metric_type),
    INDEX idx_system_metrics_timestamp (timestamp)
);

-- Create distributed_traces table for observability
CREATE TABLE IF NOT EXISTS distributed_traces (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trace_id VARCHAR(255) NOT NULL,
    span_id VARCHAR(255) NOT NULL,
    parent_span_id VARCHAR(255),
    operation_name VARCHAR(255) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    duration_ms BIGINT,
    tags JSONB,
    logs JSONB,
    INDEX idx_distributed_traces_trace_id (trace_id),
    INDEX idx_distributed_traces_span_id (span_id),
    INDEX idx_distributed_traces_operation (operation_name),
    INDEX idx_distributed_traces_start_time (start_time)
);

-- Create event_logs table for observability
CREATE TABLE IF NOT EXISTS event_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(100) NOT NULL,
    event_source VARCHAR(100) NOT NULL,
    event_data JSONB NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('debug', 'info', 'warn', 'error', 'fatal')),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    correlation_id VARCHAR(255),
    INDEX idx_event_logs_type (event_type),
    INDEX idx_event_logs_source (event_source),
    INDEX idx_event_logs_severity (severity),
    INDEX idx_event_logs_timestamp (timestamp),
    INDEX idx_event_logs_correlation_id (correlation_id)
);

-- Create traditional_metrics table for traditional monitoring
CREATE TABLE IF NOT EXISTS traditional_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    metric_name VARCHAR(255) NOT NULL,
    metric_value DECIMAL(15, 6) NOT NULL,
    metric_type VARCHAR(50) NOT NULL,
    component VARCHAR(100) NOT NULL,
    labels JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    INDEX idx_traditional_metrics_name (metric_name),
    INDEX idx_traditional_metrics_component (component),
    INDEX idx_traditional_metrics_timestamp (timestamp)
);

-- Create traditional_logs table for traditional monitoring
CREATE TABLE IF NOT EXISTS traditional_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    component VARCHAR(100) NOT NULL,
    context JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    INDEX idx_traditional_logs_level (level),
    INDEX idx_traditional_logs_component (component),
    INDEX idx_traditional_logs_timestamp (timestamp)
);

-- Create service_health table for traditional monitoring
CREATE TABLE IF NOT EXISTS service_health (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('healthy', 'unhealthy', 'degraded')),
    response_time_ms BIGINT,
    error_message TEXT,
    metadata JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    INDEX idx_service_health_service (service_name),
    INDEX idx_service_health_status (status),
    INDEX idx_service_health_timestamp (timestamp)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_type ON users(user_type);

CREATE INDEX IF NOT EXISTS idx_drivers_user_id ON drivers(user_id);
CREATE INDEX IF NOT EXISTS idx_drivers_status ON drivers(status);
CREATE INDEX IF NOT EXISTS idx_drivers_location ON drivers(current_latitude, current_longitude);

CREATE INDEX IF NOT EXISTS idx_passengers_user_id ON passengers(user_id);

CREATE INDEX IF NOT EXISTS idx_trips_passenger_id ON trips(passenger_id);
CREATE INDEX IF NOT EXISTS idx_trips_driver_id ON trips(driver_id);
CREATE INDEX IF NOT EXISTS idx_trips_status ON trips(status);
CREATE INDEX IF NOT EXISTS idx_trips_requested_at ON trips(requested_at);
CREATE INDEX IF NOT EXISTS idx_trips_pickup_location ON trips(pickup_latitude, pickup_longitude);
CREATE INDEX IF NOT EXISTS idx_trips_destination_location ON trips(destination_latitude, destination_longitude);

-- Create functions for updating timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updating timestamps
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_drivers_updated_at BEFORE UPDATE ON drivers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_passengers_updated_at BEFORE UPDATE ON passengers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_trips_updated_at BEFORE UPDATE ON trips
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert sample data for testing
INSERT INTO users (id, email, phone, name, user_type) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'john.driver@example.com', '+1234567890', 'John Driver', 'driver'),
    ('550e8400-e29b-41d4-a716-446655440002', 'jane.passenger@example.com', '+1234567891', 'Jane Passenger', 'passenger'),
    ('550e8400-e29b-41d4-a716-446655440003', 'bob.driver@example.com', '+1234567892', 'Bob Driver', 'driver'),
    ('550e8400-e29b-41d4-a716-446655440004', 'alice.passenger@example.com', '+1234567893', 'Alice Passenger', 'passenger')
ON CONFLICT (id) DO NOTHING;

INSERT INTO drivers (id, user_id, license_number, vehicle_type, vehicle_plate, status, current_latitude, current_longitude, rating, total_trips) VALUES
    ('660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'DL123456789', 'sedan', 'ABC123', 'online', -6.2088, 106.8456, 4.5, 150),
    ('660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440003', 'DL987654321', 'suv', 'XYZ789', 'offline', -6.2000, 106.8400, 4.8, 200)
ON CONFLICT (id) DO NOTHING;

INSERT INTO passengers (id, user_id, rating, total_trips) VALUES
    ('770e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', 4.7, 75),
    ('770e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440004', 4.9, 120)
ON CONFLICT (id) DO NOTHING;

-- Create a view for active trips
CREATE OR REPLACE VIEW active_trips AS
SELECT 
    t.*,
    u_p.name as passenger_name,
    u_p.phone as passenger_phone,
    u_d.name as driver_name,
    u_d.phone as driver_phone,
    d.vehicle_type,
    d.vehicle_plate
FROM trips t
JOIN passengers p ON t.passenger_id = p.id
JOIN users u_p ON p.user_id = u_p.id
LEFT JOIN drivers d ON t.driver_id = d.id
LEFT JOIN users u_d ON d.user_id = u_d.id
WHERE t.status IN ('requested', 'accepted', 'in_progress');

-- Create a view for driver statistics
CREATE OR REPLACE VIEW driver_stats AS
SELECT 
    d.id,
    u.name,
    d.license_number,
    d.vehicle_type,
    d.status,
    d.rating,
    d.total_trips,
    COUNT(t.id) as completed_trips_today,
    AVG(t.actual_fare) as avg_fare_today,
    SUM(t.actual_fare) as total_earnings_today
FROM drivers d
JOIN users u ON d.user_id = u.id
LEFT JOIN trips t ON d.id = t.driver_id 
    AND t.status = 'completed' 
    AND DATE(t.completed_at) = CURRENT_DATE
GROUP BY d.id, u.name, d.license_number, d.vehicle_type, d.status, d.rating, d.total_trips;

-- Grant permissions (adjust as needed for your security requirements)
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO postgres;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO postgres;

COMMIT;