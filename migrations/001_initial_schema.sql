-- +migrate Up
-- This migration creates all the tables defined in the ERD

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    user_type VARCHAR(20) NOT NULL CHECK (user_type IN ('passenger', 'driver')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Drivers table
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

-- Passengers table
CREATE TABLE passengers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating DECIMAL(3, 2) DEFAULT 5.00,
    total_trips INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Trips table
CREATE TABLE trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    passenger_id UUID NOT NULL REFERENCES passengers(id),
    driver_id UUID REFERENCES drivers(id),
    pickup_latitude DECIMAL(10, 8) NOT NULL,
    pickup_longitude DECIMAL(11, 8) NOT NULL,
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

-- Actor instances table
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

-- Actor messages table
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

-- System metrics table
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

-- Distributed traces table
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

-- Event logs table
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

-- Traditional monitoring tables (for comparison)
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
    metric_type VARCHAR(50) NOT NULL,
    metric_value DECIMAL(15, 6) NOT NULL,
    labels JSONB,
    service_name VARCHAR(100),
    instance_id VARCHAR(100),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance optimization

-- Business entity indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_user_type ON users(user_type);

CREATE INDEX idx_drivers_user_id ON drivers(user_id);
CREATE INDEX idx_drivers_status ON drivers(status);
CREATE INDEX idx_drivers_location ON drivers(current_latitude, current_longitude);
CREATE INDEX idx_drivers_license ON drivers(license_number);

CREATE INDEX idx_passengers_user_id ON passengers(user_id);

CREATE INDEX idx_trips_passenger_id ON trips(passenger_id);
CREATE INDEX idx_trips_driver_id ON trips(driver_id);
CREATE INDEX idx_trips_status ON trips(status);
CREATE INDEX idx_trips_requested_at ON trips(requested_at);
CREATE INDEX idx_trips_pickup_location ON trips(pickup_latitude, pickup_longitude);
CREATE INDEX idx_trips_destination_location ON trips(destination_latitude, destination_longitude);

-- Observability indexes
CREATE INDEX idx_actor_instances_type_id ON actor_instances(actor_type, actor_id);
CREATE INDEX idx_actor_instances_entity ON actor_instances(entity_id);
CREATE INDEX idx_actor_instances_status ON actor_instances(status);
CREATE INDEX idx_actor_instances_heartbeat ON actor_instances(last_heartbeat);

CREATE INDEX idx_actor_messages_trace ON actor_messages(trace_id);
CREATE INDEX idx_actor_messages_span ON actor_messages(span_id);
CREATE INDEX idx_actor_messages_parent_span ON actor_messages(parent_span_id);
CREATE INDEX idx_actor_messages_sender ON actor_messages(sender_actor_type, sender_actor_id);
CREATE INDEX idx_actor_messages_receiver ON actor_messages(receiver_actor_type, receiver_actor_id);
CREATE INDEX idx_actor_messages_type ON actor_messages(message_type);
CREATE INDEX idx_actor_messages_sent_at ON actor_messages(sent_at);
CREATE INDEX idx_actor_messages_status ON actor_messages(status);

CREATE INDEX idx_system_metrics_name ON system_metrics(metric_name);
CREATE INDEX idx_system_metrics_type ON system_metrics(metric_type);
CREATE INDEX idx_system_metrics_actor ON system_metrics(actor_type, actor_id);
CREATE INDEX idx_system_metrics_timestamp ON system_metrics(timestamp);

CREATE INDEX idx_distributed_traces_trace_id ON distributed_traces(trace_id);
CREATE INDEX idx_distributed_traces_span_id ON distributed_traces(span_id);
CREATE INDEX idx_distributed_traces_parent ON distributed_traces(parent_span_id);
CREATE INDEX idx_distributed_traces_actor ON distributed_traces(actor_type, actor_id);
CREATE INDEX idx_distributed_traces_start_time ON distributed_traces(start_time);
CREATE INDEX idx_distributed_traces_operation ON distributed_traces(operation_name);
CREATE INDEX idx_distributed_traces_status ON distributed_traces(status);

CREATE INDEX idx_event_logs_trace_id ON event_logs(trace_id);
CREATE INDEX idx_event_logs_type ON event_logs(event_type);
CREATE INDEX idx_event_logs_category ON event_logs(event_category);
CREATE INDEX idx_event_logs_actor ON event_logs(actor_type, actor_id);
CREATE INDEX idx_event_logs_entity ON event_logs(entity_type, entity_id);
CREATE INDEX idx_event_logs_timestamp ON event_logs(timestamp);
CREATE INDEX idx_event_logs_severity ON event_logs(severity);

-- Traditional monitoring indexes
CREATE INDEX idx_traditional_logs_level ON traditional_logs(level);
CREATE INDEX idx_traditional_logs_service ON traditional_logs(service_name);
CREATE INDEX idx_traditional_logs_timestamp ON traditional_logs(timestamp);

CREATE INDEX idx_traditional_metrics_name ON traditional_metrics(metric_name);
CREATE INDEX idx_traditional_metrics_type ON traditional_metrics(metric_type);
CREATE INDEX idx_traditional_metrics_service ON traditional_metrics(service_name);
CREATE INDEX idx_traditional_metrics_timestamp ON traditional_metrics(timestamp);

-- +migrate StatementBegin
-- Functions for automatic timestamp updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';
-- +migrate StatementEnd

-- Triggers for automatic timestamp updates
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_drivers_updated_at BEFORE UPDATE ON drivers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_passengers_updated_at BEFORE UPDATE ON passengers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_trips_updated_at BEFORE UPDATE ON trips
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_actor_instances_updated_at BEFORE UPDATE ON actor_instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Sample data for testing (optional)
-- This can be removed in production

-- Insert sample users
INSERT INTO users (email, phone, name, user_type) VALUES
('passenger1@example.com', '+1234567890', 'John Doe', 'passenger'),
('passenger2@example.com', '+1234567891', 'Jane Smith', 'passenger'),
('driver1@example.com', '+1234567892', 'Mike Johnson', 'driver'),
('driver2@example.com', '+1234567893', 'Sarah Wilson', 'driver');

-- Insert sample passengers
INSERT INTO passengers (user_id)
SELECT id FROM users WHERE user_type = 'passenger';

-- Insert sample drivers
INSERT INTO drivers (user_id, license_number, vehicle_type, vehicle_plate, status, current_latitude, current_longitude)
SELECT 
    id,
    'DL' || LPAD((ROW_NUMBER() OVER())::text, 6, '0'),
    'Sedan',
    'ABC' || LPAD((ROW_NUMBER() OVER())::text, 3, '0'),
    'online',
    -6.2088 + (RANDOM() * 0.1 - 0.05), -- Jakarta area with some randomness
    106.8456 + (RANDOM() * 0.1 - 0.05)
FROM users WHERE user_type = 'driver';

-- +migrate Down
-- Drop all tables in reverse order of creation

-- Drop triggers first
DROP TRIGGER IF EXISTS update_actor_instances_updated_at ON actor_instances;
DROP TRIGGER IF EXISTS update_trips_updated_at ON trips;
DROP TRIGGER IF EXISTS update_passengers_updated_at ON passengers;
DROP TRIGGER IF EXISTS update_drivers_updated_at ON drivers;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS traditional_metrics;
DROP TABLE IF EXISTS traditional_logs;
DROP TABLE IF EXISTS event_logs;
DROP TABLE IF EXISTS distributed_traces;
DROP TABLE IF EXISTS system_metrics;
DROP TABLE IF EXISTS actor_messages;
DROP TABLE IF EXISTS actor_instances;
DROP TABLE IF EXISTS trips;
DROP TABLE IF EXISTS passengers;
DROP TABLE IF EXISTS drivers;
DROP TABLE IF EXISTS users;

-- Drop extension (optional, might be used by other schemas)
-- DROP EXTENSION IF EXISTS "uuid-ossp";