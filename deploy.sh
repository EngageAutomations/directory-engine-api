#!/bin/bash

# Railway Deployment Script for Directory Engine API
# Project ID: 35f6aaac-48ab-424d-8e4c-d6148f490cdb
# Production URL: https://api.engageautomations.com

echo "🚀 Starting Railway deployment for Directory Engine API..."

# Set Railway project context
export RAILWAY_PROJECT_ID="35f6aaac-48ab-424d-8e4c-d6148f490cdb"
export RAILWAY_TOKEN="742ba18b-0328-4424-a170-545c7bb85c7a"

# Ensure we're linked to the correct project
echo "📡 Linking to Railway project..."
railway link -p $RAILWAY_PROJECT_ID

# Set essential environment variables
echo "⚙️ Setting up environment variables..."
railway variables set PORT=8080
railway variables set ENVIRONMENT=production
railway variables set DEBUG=false
railway variables set LOG_LEVEL=info
railway variables set LOG_FORMAT=json
railway variables set DB_SSL_MODE=require
railway variables set DB_MAX_OPEN_CONNS=25
railway variables set DB_MAX_IDLE_CONNS=5
railway variables set DB_CONN_MAX_LIFETIME=300s
railway variables set REDIS_DB=0
railway variables set REDIS_MAX_RETRIES=3
railway variables set REDIS_POOL_SIZE=10
railway variables set JWT_EXPIRATION=3600
railway variables set RATE_LIMIT_REQUESTS=100
railway variables set RATE_LIMIT_WINDOW=60s
railway variables set RATE_LIMIT_ENABLED=true
railway variables set CACHE_DEFAULT_TTL=300s
railway variables set CACHE_ENABLED=true
railway variables set CACHE_COMPRESSION=true
railway variables set SERVER_READ_TIMEOUT=30s
railway variables set SERVER_WRITE_TIMEOUT=30s
railway variables set SERVER_IDLE_TIMEOUT=60s

# Deploy the application
echo "🚢 Deploying to Railway..."
railway up

# Check deployment status
echo "📊 Checking deployment status..."
railway status

# Get the deployment URL
echo "🌐 Getting deployment URL..."
railway domain

echo "✅ Deployment complete!"
echo "🔗 Your API should be available at: https://api.engageautomations.com"
echo "📱 Railway Dashboard: https://railway.app/project/$RAILWAY_PROJECT_ID"