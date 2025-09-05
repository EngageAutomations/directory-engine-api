# Docker Installation Guide for Windows

Since the automated CLI installation failed, here are several methods to install Docker on Windows:

## Method 1: Manual Docker Desktop Installation (Recommended)

### Step 1: Download Docker Desktop
1. Visit the official Docker website: https://www.docker.com/products/docker-desktop/
2. Click "Download for Windows"
3. Download the Docker Desktop Installer.exe file

### Step 2: Install Docker Desktop
1. **Run as Administrator**: Right-click the installer and select "Run as administrator"
2. Follow the installation wizard
3. When prompted, ensure "Use WSL 2 instead of Hyper-V" is checked (recommended)
4. Complete the installation and restart your computer if prompted

### Step 3: Start Docker Desktop
1. Launch Docker Desktop from the Start menu
2. Accept the service agreement
3. Wait for Docker to start (you'll see the Docker whale icon in the system tray)

### Step 4: Verify Installation
Open PowerShell and run:
```powershell
docker --version
docker-compose --version
```

## Method 2: Enable WSL 2 (If Required)

Docker Desktop requires WSL 2 on Windows. If you encounter issues:

### Enable WSL 2:
1. Open PowerShell as Administrator
2. Run these commands:
```powershell
# Enable WSL
dism.exe /online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart

# Enable Virtual Machine Platform
dism.exe /online /enable-feature /featurename:VirtualMachinePlatform /all /norestart

# Restart your computer
Restart-Computer
```

3. After restart, download and install the WSL 2 kernel update:
   - Visit: https://aka.ms/wsl2kernel
   - Download and install the update package

4. Set WSL 2 as default:
```powershell
wsl --set-default-version 2
```

## Method 3: Alternative - Podman (Docker Alternative)

If Docker Desktop continues to fail, you can use Podman as an alternative:

```powershell
# Install Podman via winget
winget install RedHat.Podman

# Or download from: https://podman.io/getting-started/installation
```

## Method 4: Chocolatey Installation (If Available)

If you want to install Chocolatey first, then Docker:

### Install Chocolatey:
```powershell
# Run as Administrator
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
```

### Install Docker via Chocolatey:
```powershell
choco install docker-desktop
```

## Troubleshooting Common Issues

### Issue 1: "Docker Desktop requires a newer WSL kernel version"
**Solution**: Update WSL kernel from https://aka.ms/wsl2kernel

### Issue 2: "Hardware assisted virtualization and data execution protection must be enabled"
**Solution**: Enable virtualization in BIOS/UEFI settings

### Issue 3: "Docker Desktop starting" hangs indefinitely
**Solutions**:
1. Restart Docker Desktop
2. Reset Docker Desktop: Settings → Troubleshoot → Reset to factory defaults
3. Reinstall Docker Desktop

### Issue 4: Permission errors
**Solution**: Add your user to the "docker-users" group:
1. Open Computer Management
2. Go to Local Users and Groups → Groups
3. Double-click "docker-users"
4. Add your user account
5. Log out and log back in

## System Requirements

- **Windows 10/11**: Version 2004 or higher (Build 19041 or higher)
- **RAM**: 4GB minimum (8GB recommended)
- **CPU**: 64-bit processor with Second Level Address Translation (SLAT)
- **Virtualization**: Hardware virtualization support must be enabled

## Next Steps After Installation

Once Docker is installed:

1. **Verify installation**:
   ```bash
   docker --version
   docker run hello-world
   ```

2. **Navigate to your project**:
   ```bash
   cd "C:\Users\Computer\Documents\Engage Automations\Directory Engine 2"
   ```

3. **Run the application**:
   ```bash
   # Build and start all services
   docker-compose up --build
   
   # Or run in background
   docker-compose up -d --build
   ```

4. **Access the application**:
   - API: http://localhost:8080
   - Health Check: http://localhost:8080/health
   - Redis Commander: http://localhost:8081
   - pgAdmin: http://localhost:5050

## Alternative: Run Without Docker

If Docker installation continues to fail, you can run the application locally:

1. **Install dependencies**:
   ```bash
   go mod tidy
   ```

2. **Set up local PostgreSQL and Redis** (or use cloud services)

3. **Update .env file** with local database connections

4. **Run the application**:
   ```bash
   go run cmd/main.go
   ```

## Support

If you continue to experience issues:
1. Check Docker Desktop logs: Settings → Troubleshoot → Show logs
2. Visit Docker documentation: https://docs.docker.com/desktop/windows/
3. Check Windows Event Viewer for system errors
4. Consider using Docker in a virtual machine as a last resort

---

**Note**: The most reliable method is usually the manual installation from the official Docker website. The CLI methods can fail due to various system configurations and permission requirements.