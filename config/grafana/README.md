# Grafana Configuration for Actor Model Observability

This directory contains the complete Grafana configuration for monitoring the Actor Model Ride Hailing system.

## Directory Structure

```
config/grafana/
├── grafana.ini                     # Main Grafana configuration
├── README.md                       # This file
└── provisioning/
    ├── dashboards/
    │   ├── dashboards.yml          # Dashboard provisioning config
    │   ├── actor-observability-dashboard.json
    │   ├── business-metrics-dashboard.json
    │   ├── comparison-dashboard.json
    │   └── system-health-dashboard.json
    ├── datasources/
    │   └── prometheus.yml          # Datasource configurations
    ├── alerting/
    │   ├── rules.yml              # Alert rules
    │   ├── contactpoints.yml      # Contact points for notifications
    │   └── policies.yml           # Notification policies
    └── notifiers/
        └── notifications.yml       # Legacy notification channels
```

## Features

### Datasources
- **Prometheus**: Primary metrics datasource with optimized query settings
- **Jaeger**: Distributed tracing with correlation to logs and metrics

### Dashboards
1. **Actor Observability Dashboard**: Core actor system metrics
   - Active actor instances by type
   - Message throughput and processing latency
   - Business metrics (ride requests, matches, completions)
   - System resource usage
   - Actor health status

2. **System Health Dashboard**: Infrastructure monitoring
   - Application, database, and cache status
   - System throughput metrics
   - Resource utilization

3. **Business Metrics Dashboard**: Business KPIs
   - Ride requests and successful matches
   - Trip completions and revenue metrics
   - Customer satisfaction indicators

4. **Comparison Dashboard**: Actor vs Traditional architecture
   - Performance comparisons
   - Observability quality metrics
   - System efficiency analysis

### Alerting

#### Alert Rules
- **High Actor Message Failure Rate**: Triggers when message failure rate > 5%
- **Actor System Down**: Alerts when the main application is unreachable
- **High HTTP Response Time**: Warns when 95th percentile > 500ms
- **Database Connection Issues**: Critical alert for database connectivity

#### Contact Points
- **Email**: Primary notification method for all alerts
- **Slack**: Real-time notifications for critical alerts
- **Webhook**: Integration with external systems
- **PagerDuty**: On-call escalation for critical issues

#### Notification Policies
- Critical alerts: Immediate notification via Slack + PagerDuty
- Warning alerts: Email notifications with longer intervals
- Component-specific routing (database, API, actor-system)
- Mute time intervals for maintenance windows

## Configuration Details

### Security Settings
- Admin credentials: `admin/admin` (change in production)
- Anonymous access disabled
- User registration disabled
- Secure cookie settings for HTTPS

### Performance Optimizations
- Query timeout: 60 seconds
- HTTP method: POST for large queries
- Connection pooling configured
- Caching enabled for better performance

### Tracing Integration
- OpenTelemetry OTLP endpoint configured
- Jaeger propagation enabled
- Trace-to-logs correlation
- Trace-to-metrics correlation
- Service map visualization

## Setup Instructions

### 1. Docker Compose Setup
The configuration is automatically mounted in the Docker Compose setup:

```yaml
grafana:
  image: grafana/grafana:latest
  volumes:
    - ./config/grafana/provisioning:/etc/grafana/provisioning
    - ./config/grafana/grafana.ini:/etc/grafana/grafana.ini
```

### 2. Access Grafana
- URL: http://localhost:3000
- Username: `admin`
- Password: `admin`

### 3. Verify Setup
1. Check datasources are connected (Settings > Data Sources)
2. Verify dashboards are loaded (Dashboards)
3. Test alert rules (Alerting > Alert Rules)
4. Configure notification channels (Alerting > Contact Points)

## Customization

### Adding New Dashboards
1. Create JSON dashboard file in `provisioning/dashboards/`
2. Dashboard will be automatically loaded on Grafana restart

### Modifying Alert Rules
1. Edit `provisioning/alerting/rules.yml`
2. Restart Grafana to apply changes

### Configuring Notifications
1. Update contact points in `provisioning/alerting/contactpoints.yml`
2. Modify routing in `provisioning/alerting/policies.yml`
3. Set up external integrations (Slack webhooks, PagerDuty keys)

### Environment-Specific Settings
For production deployment:

1. **Change default credentials**:
   ```ini
   [security]
   admin_password = your_secure_password
   ```

2. **Enable HTTPS**:
   ```ini
   [server]
   protocol = https
   cert_file = /path/to/cert.pem
   cert_key = /path/to/cert.key
   ```

3. **Configure external database**:
   ```ini
   [database]
   type = postgres
   host = your-db-host:5432
   name = grafana
   user = grafana_user
   password = grafana_password
   ```

4. **Set up SMTP for email alerts**:
   ```ini
   [smtp]
   enabled = true
   host = smtp.gmail.com:587
   user = your-email@gmail.com
   password = your-app-password
   ```

## Troubleshooting

### Common Issues

1. **Dashboards not loading**:
   - Check file permissions on provisioning directory
   - Verify JSON syntax in dashboard files
   - Check Grafana logs for errors

2. **Datasource connection failed**:
   - Verify Prometheus is running and accessible
   - Check network connectivity between containers
   - Validate datasource URL in configuration

3. **Alerts not firing**:
   - Check alert rule syntax and queries
   - Verify contact point configurations
   - Test notification channels manually

4. **Performance issues**:
   - Increase query timeout settings
   - Optimize dashboard queries
   - Check Prometheus retention settings

### Logs and Debugging

```bash
# View Grafana logs
docker logs actor-observability-grafana

# Check configuration
docker exec actor-observability-grafana grafana-cli admin data-migration list

# Test datasource connectivity
curl -u admin:admin http://localhost:3000/api/datasources/proxy/1/api/v1/query?query=up
```

## Monitoring Best Practices

1. **Dashboard Organization**:
   - Use folders to organize dashboards by domain
   - Create role-based access for different teams
   - Implement consistent naming conventions

2. **Alert Management**:
   - Set appropriate thresholds based on SLAs
   - Implement alert fatigue prevention
   - Use alert dependencies to reduce noise
   - Regular review and tuning of alert rules

3. **Performance**:
   - Use recording rules for expensive queries
   - Implement proper data retention policies
   - Monitor Grafana's own performance metrics

4. **Security**:
   - Regular credential rotation
   - Implement proper RBAC
   - Audit dashboard and alert changes
   - Secure external integrations

## Integration with Actor System

This Grafana setup is specifically designed to monitor the Actor Model architecture:

- **Actor Lifecycle**: Track actor creation, destruction, and health
- **Message Flow**: Monitor message passing between actors
- **Business Logic**: Observe ride-hailing specific metrics
- **Performance**: Compare actor model vs traditional approaches
- **Distributed Tracing**: Follow requests across actor boundaries

The configuration provides comprehensive observability for both technical and business stakeholders, enabling data-driven decisions about system performance and reliability.