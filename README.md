# Go Marketplace Application with Nango Integration

A high-performance Go marketplace application that integrates with Nango for OAuth management, designed to handle heavy traffic and data processing for marketplace installations.

## Features

- **Nango OAuth Integration**: Seamless company authorization and token management
- **Automatic Token Refresh**: Background jobs ensure tokens never expire
- **High Performance**: Built with Gin, Redis caching, and connection pooling
- **Scalable Architecture**: Microservices pattern with clean separation of concerns
- **Comprehensive API**: RESTful endpoints for all marketplace operations
- **Admin Dashboard**: Administrative endpoints for system management
- **Health Monitoring**: Detailed health checks and metrics
- **Production Ready**: Docker support, Railway deployment, and monitoring

## Architecture

```
├── cmd/                    # Application entry points
├── internal/
│   ├── api/               # HTTP handlers and routes
│   │   ├── handlers/      # Request handlers
│   │   └── middleware/    # HTTP middleware
│   ├── config/            # Configuration management
│   ├── database/          # Database connection and migrations
│   ├── models/            # Data models
│   └── services/          # Business logic services
├── scripts/               # Database and deployment scripts
├── docker-compose.yml     # Local development setup
├── Dockerfile            # Container configuration
└── railway.toml          # Railway deployment config
```

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Docker (optional)

### Local Development

1. **Clone and setup**:
   ```bash
   git clone <repository-url>
   cd marketplace-app
   cp .env.example .env
   ```

2. **Configure environment**:
   Edit `.env` with your settings:
   ```env
   # Database
   DB_HOST=localhost
   DB_USER=marketplace_user
   DB_PASSWORD=your_password
   
   # Nango
   NANGO_CLIENT_ID=your_client_id
   NANGO_CLIENT_SECRET=your_client_secret
   NANGO_API_KEY=your_api_key
   
   # JWT
   JWT_SECRET=your_jwt_secret
   ```

3. **Run with Docker Compose**:
   ```bash
   docker-compose up -d
   ```

4. **Or run locally**:
   ```bash
   go mod download
   go run main.go
   ```

### Railway Deployment

1. **Install Railway CLI**:
   ```bash
   npm install -g @railway/cli
   ```

2. **Deploy**:
   ```bash
   railway login
   railway init
   railway up
   ```

3. **Add services**:
   ```bash
   railway add postgresql
   railway add redis
   ```

4. **Set environment variables** in Railway dashboard:
   - `NANGO_CLIENT_ID`
   - `NANGO_CLIENT_SECRET`
   - `NANGO_API_KEY`
   - `JWT_SECRET`
   - `ADMIN_TOKEN`

## API Documentation

### Authentication Endpoints

#### Get OAuth URL
```http
GET /api/auth/oauth-url?integration_id=quickbooks&connection_id=company_123
```

#### Handle OAuth Callback
```http
GET /api/auth/callback?code=auth_code&state=connection_id
```

#### Exchange Token
```http
POST /api/auth/exchange
Content-Type: application/json

{
  "connection_id": "company_123",
  "integration_id": "quickbooks"
}
```

### Business Endpoints

#### Get Companies
```http
GET /api/companies
Authorization: Bearer <jwt_token>
```

#### Get Company Details
```http
GET /api/companies/{id}
Authorization: Bearer <jwt_token>
```

#### Sync Company Data
```http
POST /api/companies/{id}/sync
Authorization: Bearer <jwt_token>
```

#### Get Locations
```http
GET /api/companies/{company_id}/locations
Authorization: Bearer <jwt_token>
```

#### Get Contacts
```http
GET /api/companies/{company_id}/contacts
Authorization: Bearer <jwt_token>
```

#### Get Products
```http
GET /api/companies/{company_id}/products
Authorization: Bearer <jwt_token>
```

### Admin Endpoints

#### Get All Tokens
```http
GET /api/admin/tokens
X-Admin-Token: <admin_token>
```

#### Refresh All Tokens
```http
POST /api/admin/tokens/refresh-all
X-Admin-Token: <admin_token>
```

#### System Health
```http
GET /api/admin/system/health
X-Admin-Token: <admin_token>
```

### Health Endpoints

#### Basic Health Check
```http
GET /health
```

#### Detailed Health Check
```http
GET /health/detailed
```

#### Readiness Check
```http
GET /health/ready
```

#### Metrics
```http
GET /metrics
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|----------|
| `DB_HOST` | Database host | localhost |
| `DB_PORT` | Database port | 5432 |
| `DB_USER` | Database user | marketplace_user |
| `DB_PASSWORD` | Database password | - |
| `DB_NAME` | Database name | marketplace_db |
| `REDIS_HOST` | Redis host | localhost |
| `REDIS_PORT` | Redis port | 6379 |
| `NANGO_BASE_URL` | Nango API URL | https://api.nango.dev |
| `NANGO_CLIENT_ID` | Nango client ID | - |
| `NANGO_CLIENT_SECRET` | Nango client secret | - |
| `NANGO_API_KEY` | Nango API key | - |
| `JWT_SECRET` | JWT signing secret | - |
| `SERVER_PORT` | Server port | 8080 |
| `RATE_LIMIT_REQUESTS` | Rate limit per window | 100 |
| `CACHE_DEFAULT_TTL` | Default cache TTL | 300s |

### Database Models

#### Company
```go
type Company struct {
    ID              uuid.UUID `gorm:"type:uuid;primary_key"`
    NangoConnectionID string  `gorm:"uniqueIndex"`
    IntegrationID   string
    BusinessName    string
    AccessToken     string
    RefreshToken    string
    TokenExpiresAt  *time.Time
    // ... other fields
}
```

#### Location
```go
type Location struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key"`
    CompanyID uuid.UUID `gorm:"type:uuid;not null"`
    Name      string
    Address   string
    // ... other fields
}
```

## Performance Features

### Caching Strategy
- **Redis Primary**: Main caching layer with compression
- **In-Memory Fallback**: Local cache when Redis unavailable
- **Smart Invalidation**: Automatic cache clearing on data updates

### Rate Limiting
- **Redis-based**: Distributed rate limiting across instances
- **Configurable**: Adjustable limits per endpoint
- **Graceful Degradation**: Fallback when Redis unavailable

### Connection Pooling
- **Database**: Optimized PostgreSQL connection pool
- **Redis**: Connection pooling with retry logic
- **HTTP**: Keep-alive connections for external APIs

### Background Jobs
- **Token Refresh**: Hourly automatic token refresh
- **Cleanup**: Daily cleanup of expired tokens
- **Health Monitoring**: Continuous system health checks

## Monitoring and Observability

### Health Checks
- `/health` - Basic application health
- `/health/detailed` - Comprehensive system status
- `/health/ready` - Kubernetes readiness probe
- `/health/live` - Kubernetes liveness probe

### Metrics
- Request/response metrics
- Database connection stats
- Cache hit/miss ratios
- Token refresh success rates
- System resource usage

### Logging
- Structured JSON logging
- Request ID tracking
- Error stack traces
- Performance metrics

## Security Features

- **JWT Authentication**: Secure API access
- **Admin Token**: Protected administrative endpoints
- **CORS Configuration**: Configurable cross-origin policies
- **Rate Limiting**: DDoS protection
- **Input Validation**: Request sanitization
- **Security Headers**: Standard security headers

## Development

### Running Tests
```bash
go test ./...
```

### Code Generation
```bash
# Generate mocks
go generate ./...
```

### Database Migrations
```bash
# Auto-migration runs on startup
# Manual migration
go run scripts/migrate.go
```

### Local Development with Tools
```bash
# Start with management tools
docker-compose --profile tools up -d

# Access tools:
# - pgAdmin: http://localhost:8082
# - Redis Commander: http://localhost:8081
```

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check PostgreSQL is running
   - Verify connection string
   - Check firewall settings

2. **Redis Connection Failed**
   - Ensure Redis is running
   - Check Redis configuration
   - Verify network connectivity

3. **Nango API Errors**
   - Verify API credentials
   - Check Nango service status
   - Review webhook configuration

4. **Token Refresh Issues**
   - Check token expiration times
   - Verify refresh token validity
   - Review scheduler logs

### Debug Mode
```bash
# Enable debug logging
export DEBUG=true
export LOG_LEVEL=debug
go run main.go
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details.