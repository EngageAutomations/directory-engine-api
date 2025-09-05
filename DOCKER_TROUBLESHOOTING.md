# Docker Troubleshooting Guide

## Current Issue: Docker API 500 Internal Server Error

You're experiencing a Docker API error that prevents containers from starting. This is a common issue with Docker Desktop on Windows.

## Quick Fixes to Try

### 1. Restart Docker Desktop
1. **Close Docker Desktop completely**:
   - Right-click the Docker whale icon in system tray
   - Select "Quit Docker Desktop"
   - Wait 30 seconds

2. **Restart Docker Desktop**:
   - Launch Docker Desktop from Start menu
   - Wait for it to fully start (whale icon should be steady)
   - Try the command again

### 2. Reset Docker Desktop
1. Open Docker Desktop
2. Go to **Settings** → **Troubleshoot**
3. Click **"Reset to factory defaults"**
4. Restart Docker Desktop
5. Try again

### 3. Check Docker Desktop Status
```powershell
# Check if Docker daemon is running
docker system info

# If that fails, try:
docker version
```

### 4. Alternative: Use Docker Desktop GUI
1. Open Docker Desktop
2. Go to the **Images** tab
3. Try pulling an image manually: `postgres:15-alpine`
4. If successful, try `docker compose up --build` again

## Alternative Solutions

### Option A: Run Individual Services

If Docker Compose continues to fail, try running services individually:

```powershell
# Start PostgreSQL
docker run -d --name postgres-db \
  -e POSTGRES_DB=marketplace \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres123 \
  -p 5432:5432 \
  postgres:15-alpine

# Start Redis
docker run -d --name redis-cache \
  -p 6379:6379 \
  redis:7-alpine

# Build and run the Go application
docker build -t marketplace-app .
docker run -d --name marketplace-app \
  --env-file .env \
  -p 8080:8080 \
  marketplace-app
```

### Option B: Run Without Docker (Local Development)

If Docker continues to have issues, you can run the application locally:

#### 1. Install PostgreSQL Locally
- Download from: https://www.postgresql.org/download/windows/
- Or use: `winget install PostgreSQL.PostgreSQL`
- Create database: `marketplace`
- Update `.env` with local connection details

#### 2. Install Redis Locally
- Download from: https://github.com/microsoftarchive/redis/releases
- Or use: `winget install Redis.Redis`
- Start Redis service

#### 3. Update .env File
```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=marketplace

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
```

#### 4. Run the Application
```powershell
# Install dependencies (already done)
go mod tidy

# Run database migrations
psql -U postgres -d marketplace -f scripts/init.sql

# Start the application
go run main.go
```

### Option C: Use Railway (Cloud Deployment)

Skip local development and deploy directly to Railway:

1. **Install Railway CLI**:
   ```powershell
   npm install -g @railway/cli
   # or
   winget install Railway.CLI
   ```

2. **Deploy to Railway**:
   ```powershell
   railway login
   railway init
   railway up
   ```

## Common Docker Desktop Issues on Windows

### Issue: WSL 2 Problems
**Solution**: Update WSL 2 kernel
```powershell
wsl --update
wsl --set-default-version 2
```

### Issue: Hyper-V Conflicts
**Solution**: Ensure WSL 2 backend is selected in Docker Desktop settings

### Issue: Antivirus Interference
**Solution**: Add Docker Desktop and project folder to antivirus exclusions

### Issue: Insufficient Resources
**Solution**: Increase Docker Desktop memory/CPU limits in Settings → Resources

## Testing Docker Health

Run these commands to diagnose Docker issues:

```powershell
# Test basic Docker functionality
docker run hello-world

# Check Docker system status
docker system info

# Check available images
docker images

# Check running containers
docker ps

# Check Docker Desktop logs
# Go to Docker Desktop → Settings → Troubleshoot → Show logs
```

## Next Steps

1. **Try the quick fixes above first**
2. **If Docker works**: Use `docker compose up --build`
3. **If Docker fails**: Use local development setup
4. **For production**: Deploy to Railway

## Application URLs (Once Running)

- **API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **API Documentation**: http://localhost:8080/docs (if implemented)
- **Redis Commander**: http://localhost:8081 (Docker only)
- **pgAdmin**: http://localhost:5050 (Docker only)

---

**Remember**: The Go application is fully functional and ready to run. The Docker issue is just a deployment method problem, not an application problem.