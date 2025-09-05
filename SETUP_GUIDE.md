# Setup Guide - Go Marketplace Application

## ✅ **What's Already Complete**

✅ **Step 1**: Project structure and code implementation  
✅ **Step 2**: Environment configuration (`.env` file created from template)

## 🔧 **What You Need to Install**

### **Required Software:**

1. **Go 1.21+** - [Download from golang.org](https://golang.org/dl/)
   - Choose the Windows installer for your system (64-bit recommended)
   - After installation, restart your terminal/PowerShell
   - Verify with: `go version`

2. **Docker Desktop** - [Download from docker.com](https://www.docker.com/products/docker-desktop/)
   - Install Docker Desktop for Windows
   - Enable WSL 2 backend if prompted
   - Start Docker Desktop after installation
   - Verify with: `docker --version`

## 🚀 **Next Steps After Installation**

### **Option A: Run with Docker (Recommended for Development)**
```powershell
# Navigate to project directory
cd "C:\Users\Computer\Documents\Engage Automations\Directory Engine 2"

# Start all services (PostgreSQL, Redis, and the app)
docker-compose up -d

# Check if services are running
docker-compose ps

# View logs
docker-compose logs -f marketplace-app
```

### **Option B: Run Locally (Requires PostgreSQL and Redis)**
```powershell
# Install dependencies
go mod tidy

# Run the application
go run main.go
```

## 🔧 **Configuration**

### **Required Environment Variables in `.env`:**
Before running, update these values in your `.env` file:

```env
# Nango Configuration (Get from your Nango dashboard)
NANGO_CLIENT_ID=your_actual_nango_client_id
NANGO_CLIENT_SECRET=your_actual_nango_client_secret
NANGO_API_KEY=your_actual_nango_api_key
NANGO_WEBHOOK_SECRET=your_actual_webhook_secret

# JWT Secret (Generate a secure random string)
JWT_SECRET=your_secure_jwt_secret_at_least_32_characters_long

# Admin Token (For accessing admin endpoints)
ADMIN_TOKEN=your_secure_admin_token
```

### **Generate Secure Secrets:**
```powershell
# Generate JWT Secret (32+ characters)
[System.Web.Security.Membership]::GeneratePassword(64, 10)

# Or use online generator: https://generate-secret.vercel.app/32
```

## 🌐 **Testing the Application**

Once running, test these endpoints:

### **Health Check:**
```
GET http://localhost:8080/health
```

### **Get OAuth URL:**
```
GET http://localhost:8080/api/auth/oauth-url?integration_id=quickbooks&connection_id=test_company
```

### **Admin Health Check:**
```
GET http://localhost:8080/api/admin/system/health
Headers: X-Admin-Token: your_admin_token
```

## 🐳 **Docker Services**

When using Docker Compose, these services will be available:

- **Application**: http://localhost:8080
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379
- **pgAdmin** (optional): http://localhost:8082 (admin@marketplace.com / admin)
- **Redis Commander** (optional): http://localhost:8081

### **Docker Commands:**
```powershell
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# Restart a specific service
docker-compose restart marketplace-app

# Start with management tools
docker-compose --profile tools up -d
```

## 🚀 **Railway Deployment**

For production deployment to Railway:

1. **Install Railway CLI:**
   ```powershell
   npm install -g @railway/cli
   ```

2. **Deploy:**
   ```powershell
   railway login
   railway init
   railway add postgresql
   railway add redis
   railway up
   ```

3. **Set Environment Variables** in Railway dashboard:
   - All the Nango credentials
   - JWT_SECRET
   - ADMIN_TOKEN

## 🔍 **Troubleshooting**

### **Common Issues:**

1. **Port 8080 already in use:**
   ```powershell
   # Find process using port 8080
   netstat -ano | findstr :8080
   # Kill the process (replace PID)
   taskkill /PID <PID> /F
   ```

2. **Docker not starting:**
   - Ensure Docker Desktop is running
   - Check Windows features: Hyper-V and WSL 2
   - Restart Docker Desktop

3. **Database connection issues:**
   - Ensure PostgreSQL container is running: `docker-compose ps`
   - Check database logs: `docker-compose logs postgres`

4. **Go module issues:**
   ```powershell
   go clean -modcache
   go mod download
   go mod tidy
   ```

## 📚 **API Documentation**

Full API documentation is available in the main [README.md](README.md) file.

## 🎯 **What's Ready**

Your marketplace application includes:
- ✅ Complete Nango OAuth integration
- ✅ Automatic token refresh system
- ✅ Company, location, contact, and product management
- ✅ High-performance caching and rate limiting
- ✅ Comprehensive health monitoring
- ✅ Production-ready Docker configuration
- ✅ Railway deployment setup

**Once you install Go and Docker, everything is ready to run!**