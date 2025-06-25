// K6 Load Testing Script for Actor Model Ride Hailing Observability
// This script tests various endpoints with realistic ride-hailing scenarios
// and generates metrics that are visible in Grafana dashboards

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { randomString, randomIntBetween, randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Custom metrics for better Grafana integration
const rideRequestRate = new Rate('ride_requests');
const rideRequestDuration = new Trend('ride_request_duration');
const driverLocationUpdates = new Counter('driver_location_updates');
const systemErrors = new Rate('system_errors');
const actorSystemRequests = new Counter('actor_system_requests');
const traditionalSystemRequests = new Counter('traditional_system_requests');

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 10 },   // Ramp up to 10 users
    { duration: '5m', target: 10 },   // Stay at 10 users
    { duration: '2m', target: 20 },   // Ramp up to 20 users
    { duration: '5m', target: 20 },   // Stay at 20 users
    { duration: '2m', target: 0 },    // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
    http_req_failed: ['rate<0.1'],    // Error rate should be below 10%
    ride_requests: ['rate>0.1'],      // Ride request rate should be above 10%
  },
};

// Base URL configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Pre-defined test data
const USER_IDS = [
  '550e8400-e29b-41d4-a716-446655440001',
  '550e8400-e29b-41d4-a716-446655440002',
  '550e8400-e29b-41d4-a716-446655440003',
  '550e8400-e29b-41d4-a716-446655440004',
  '550e8400-e29b-41d4-a716-446655440005',
];

const DRIVER_IDS = [
  '660e8400-e29b-41d4-a716-446655440001',
  '660e8400-e29b-41d4-a716-446655440002',
  '660e8400-e29b-41d4-a716-446655440003',
  '660e8400-e29b-41d4-a716-446655440004',
  '660e8400-e29b-41d4-a716-446655440005',
];

const PASSENGER_IDS = [
  '770e8400-e29b-41d4-a716-446655440010',
  '770e8400-e29b-41d4-a716-446655440011',
  '770e8400-e29b-41d4-a716-446655440012',
];

const RIDE_IDS = [
  '770e8400-e29b-41d4-a716-446655440001',
  '770e8400-e29b-41d4-a716-446655440002',
  '874915c5-744c-4ceb-9a3d-c718ab72eb0d',
];

// Helper functions
function randomLocation() {
  // Random coordinates around San Francisco area
  const lat = 37.7749 + (Math.random() - 0.5) * 0.1;
  const lng = -122.4194 + (Math.random() - 0.5) * 0.1;
  return { lat, lng };
}

function randomPhone() {
  return `+1${randomIntBetween(100, 999)}${randomIntBetween(100, 999)}${randomIntBetween(1000, 9999)}`;
}

function randomEmail() {
  return `${randomString(8)}@example.com`;
}

// Test scenarios with weighted distribution
const scenarios = [
  { name: 'rideRequest', weight: 30, func: testRequestRide },
  { name: 'getRideStatus', weight: 20, func: testGetRideStatus },
  { name: 'listRides', weight: 10, func: testListRides },
  { name: 'updateDriverLocation', weight: 15, func: testUpdateDriverLocation },
  { name: 'updateDriverStatus', weight: 10, func: testUpdateDriverStatus },
  { name: 'observabilityMetrics', weight: 5, func: testGetObservabilityMetrics },
  { name: 'traditionalMetrics', weight: 5, func: testGetTraditionalMetrics },
  { name: 'healthCheck', weight: 3, func: testHealthCheck },
  { name: 'systemMetrics', weight: 2, func: testSystemMetrics },
];

// Calculate cumulative weights for scenario selection
let cumulativeWeights = [];
let totalWeight = 0;
for (let scenario of scenarios) {
  totalWeight += scenario.weight;
  cumulativeWeights.push(totalWeight);
}

// Main test function
export default function () {
  // Select scenario based on weight
  const random = Math.random() * totalWeight;
  let selectedScenario = scenarios[0];
  
  for (let i = 0; i < cumulativeWeights.length; i++) {
    if (random <= cumulativeWeights[i]) {
      selectedScenario = scenarios[i];
      break;
    }
  }
  
  // Execute selected scenario
  selectedScenario.func();
  
  // Random sleep between requests (0.5-2 seconds)
  sleep(Math.random() * 1.5 + 0.5);
}

// Test scenario functions
function testRequestRide() {
  const passengerId = randomItem(PASSENGER_IDS);
  const pickup = randomLocation();
  const destination = randomLocation();
  
  const payload = JSON.stringify({
    passenger_id: passengerId,
    pickup_lat: pickup.lat,
    pickup_lng: pickup.lng,
    destination_lat: destination.lat,
    destination_lng: destination.lng,
    ride_type: randomItem(['standard', 'premium']),
  });
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    tags: {
      endpoint: 'ride_request',
      system_type: 'actor',
    },
  };
  
  const response = http.post(`${BASE_URL}/api/v1/rides/request`, payload, params);
  
  const success = check(response, {
    'ride request status is 200 or 201': (r) => r.status === 200 || r.status === 201,
    'ride request response time < 1000ms': (r) => r.timings.duration < 1000,
    'ride request contains trip_id': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.trip_id !== undefined;
      } catch {
        return false;
      }
    },
  });
  
  rideRequestRate.add(success);
  rideRequestDuration.add(response.timings.duration);
  actorSystemRequests.add(1);
  
  if (!success) {
    systemErrors.add(1);
  }
}

function testGetRideStatus() {
  const tripId = randomItem(RIDE_IDS);
  
  const params = {
    tags: {
      endpoint: 'ride_status',
      system_type: 'actor',
    },
  };
  
  const response = http.get(`${BASE_URL}/api/v1/rides/${tripId}/status`, params);
  
  const success = check(response, {
    'ride status check is 200': (r) => r.status === 200,
    'ride status response time < 500ms': (r) => r.timings.duration < 500,
  });
  
  actorSystemRequests.add(1);
  
  if (!success) {
    systemErrors.add(1);
  }
}

function testListRides() {
  const params = {
    tags: {
      endpoint: 'list_rides',
      system_type: 'actor',
    },
  };
  
  const response = http.get(`${BASE_URL}/api/v1/rides`, params);
  
  const success = check(response, {
    'list rides status is 200': (r) => r.status === 200,
    'list rides response time < 500ms': (r) => r.timings.duration < 500,
  });
  
  actorSystemRequests.add(1);
  
  if (!success) {
    systemErrors.add(1);
  }
}

function testUpdateDriverLocation() {
  const driverId = randomItem(DRIVER_IDS);
  const location = randomLocation();
  
  const payload = JSON.stringify({
    latitude: location.lat,
    longitude: location.lng,
  });
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    tags: {
      endpoint: 'driver_location',
      system_type: 'actor',
    },
  };
  
  const response = http.put(`${BASE_URL}/api/v1/drivers/${driverId}/location`, payload, params);
  
  const success = check(response, {
    'driver location update status is 200': (r) => r.status === 200,
    'driver location update response time < 300ms': (r) => r.timings.duration < 300,
  });
  
  driverLocationUpdates.add(1);
  actorSystemRequests.add(1);
  
  if (!success) {
    systemErrors.add(1);
  }
}

function testUpdateDriverStatus() {
  const driverId = randomItem(DRIVER_IDS);
  const statuses = ['online', 'busy', 'offline'];
  const status = randomItem(statuses);
  
  const payload = JSON.stringify({
    status: status,
  });
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    tags: {
      endpoint: 'driver_status',
      system_type: 'actor',
    },
  };
  
  const response = http.put(`${BASE_URL}/api/v1/drivers/${driverId}/status`, payload, params);
  
  const success = check(response, {
    'driver status update status is 200': (r) => r.status === 200,
    'driver status update response time < 300ms': (r) => r.timings.duration < 300,
  });
  
  actorSystemRequests.add(1);
  
  if (!success) {
    systemErrors.add(1);
  }
}

function testGetObservabilityMetrics() {
  const params = {
    tags: {
      endpoint: 'observability_metrics',
      system_type: 'actor',
    },
  };
  
  const response = http.get(`${BASE_URL}/api/v1/observability/prometheus`, params);
  
  const success = check(response, {
    'observability metrics status is 200': (r) => r.status === 200,
    'observability metrics response time < 1000ms': (r) => r.timings.duration < 1000,
  });
  
  actorSystemRequests.add(1);
  
  if (!success) {
    systemErrors.add(1);
  }
}

function testGetTraditionalMetrics() {
  const params = {
    tags: {
      endpoint: 'traditional_metrics',
      system_type: 'traditional',
    },
  };
  
  const response = http.get(`${BASE_URL}/api/v1/traditional/metrics`, params);
  
  const success = check(response, {
    'traditional metrics status is 200': (r) => r.status === 200,
    'traditional metrics response time < 1000ms': (r) => r.timings.duration < 1000,
  });
  
  traditionalSystemRequests.add(1);
  
  if (!success) {
    systemErrors.add(1);
  }
}

function testHealthCheck() {
  const params = {
    tags: {
      endpoint: 'health_check',
      system_type: 'system',
    },
  };
  
  const response = http.get(`${BASE_URL}/health/ping`, params);
  
  const success = check(response, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 100ms': (r) => r.timings.duration < 100,
  });
  
  if (!success) {
    systemErrors.add(1);
  }
}

function testSystemMetrics() {
  const params = {
    tags: {
      endpoint: 'system_metrics',
      system_type: 'actor',
    },
  };
  
  const response = http.get(`${BASE_URL}/api/v1/observability/metrics`, params);
  
  const success = check(response, {
    'system metrics status is 200': (r) => r.status === 200,
    'system metrics response time < 500ms': (r) => r.timings.duration < 500,
  });
  
  actorSystemRequests.add(1);
  
  if (!success) {
    systemErrors.add(1);
  }
}

// Setup function (runs once per VU)
export function setup() {
  console.log('Starting K6 load test for Actor Model Ride Hailing Observability');
  console.log(`Target URL: ${BASE_URL}`);
  
  // Verify server is accessible
  const response = http.get(`${BASE_URL}/health/ping`);
  if (response.status !== 200) {
    throw new Error(`Server not accessible at ${BASE_URL}. Status: ${response.status}`);
  }
  
  console.log('Server is accessible, starting load test...');
}

// Teardown function (runs once after all VUs finish)
export function teardown(data) {
  console.log('K6 load test completed');
}