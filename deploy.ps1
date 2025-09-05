# Railway Deployment Script for Directory Engine API
# Project ID: 35f6aaac-48ab-424d-8e4c-d6148f490cdb
# Production URL: https://api.engageautomations.com

Write-Host "🚀 Starting Railway deployment for Directory Engine API..." -ForegroundColor Green

# Set Railway project context
$env:RAILWAY_PROJECT_ID = "35f6aaac-48ab-424d-8e4c-d6148f490cdb"
$env:RAILWAY_TOKEN = "742ba18b-0328-4424-a170-545c7bb85c7a"

# Ensure we're linked to the correct project
Write-Host "📡 Linking to Railway project..." -ForegroundColor Yellow
railway link -p $env:RAILWAY_PROJECT_ID

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to link to Railway project" -ForegroundColor Red
    exit 1
}

# Set essential environment variables
Write-Host "⚙️ Setting up environment variables..." -ForegroundColor Yellow
$variables = @{
    "PORT" = "8080"
    "ENVIRONMENT" = "production"
    "DEBUG" = "false"
    "LOG_LEVEL" = "info"
    "LOG_FORMAT" = "json"
    "DB_SSL_MODE" = "require"
    "DB_MAX_OPEN_CONNS" = "25"
    "DB_MAX_IDLE_CONNS" = "5"
    "DB_CONN_MAX_LIFETIME" = "300s"
    "REDIS_DB" = "0"
    "REDIS_MAX_RETRIES" = "3"
    "REDIS_POOL_SIZE" = "10"
    "JWT_EXPIRATION" = "3600"
    "RATE_LIMIT_REQUESTS" = "100"
    "RATE_LIMIT_WINDOW" = "60s"
    "RATE_LIMIT_ENABLED" = "true"
    "CACHE_DEFAULT_TTL" = "300s"
    "CACHE_ENABLED" = "true"
    "CACHE_COMPRESSION" = "true"
    "SERVER_READ_TIMEOUT" = "30s"
    "SERVER_WRITE_TIMEOUT" = "30s"
    "SERVER_IDLE_TIMEOUT" = "60s"
}

foreach ($key in $variables.Keys) {
    Write-Host "Setting $key = $($variables[$key])" -ForegroundColor Cyan
    railway variables set "$key=$($variables[$key])"
}

# Deploy the application
Write-Host "🚢 Deploying to Railway..." -ForegroundColor Yellow
railway up

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Deployment failed" -ForegroundColor Red
    exit 1
}

# Check deployment status
Write-Host "📊 Checking deployment status..." -ForegroundColor Yellow
railway status

# Get the deployment URL
Write-Host "🌐 Getting deployment URL..." -ForegroundColor Yellow
railway domain

Write-Host "✅ Deployment complete!" -ForegroundColor Green
Write-Host "🔗 Your API should be available at: https://api.engageautomations.com" -ForegroundColor Green
Write-Host "📱 Railway Dashboard: https://railway.app/project/$env:RAILWAY_PROJECT_ID" -ForegroundColor Green

# Test the deployment
Write-Host "🧪 Testing deployment..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "https://api.engageautomations.com/health" -Method GET -TimeoutSec 10
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ Health check passed!" -ForegroundColor Green
    } else {
        Write-Host "⚠️ Health check returned status: $($response.StatusCode)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "⚠️ Health check failed: $($_.Exception.Message)" -ForegroundColor Yellow
    Write-Host "This is normal if the service is still starting up." -ForegroundColor Gray
}