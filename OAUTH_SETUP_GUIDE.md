# OAuth Setup Guide for Directory Engine API

This guide will help you configure OAuth integration using your custom domain `api.engageautomations.com`.

## 🔧 Nango Configuration

### 1. Nango Dashboard Setup

1. **Login to your Nango Dashboard**: https://app.nango.dev
2. **Create or select your integration**
3. **Configure the redirect URI**:
   ```
   https://api.engageautomations.com/api/v1/auth/oauth/callback
   ```

### 2. Environment Variables

Set these environment variables in your Railway dashboard:

```bash
# Nango Configuration
NANGO_PUBLIC_KEY=your_nango_public_key_from_dashboard
NANGO_SECRET_KEY=your_nango_secret_key_from_dashboard
NANGO_SERVER_URL=https://api.nango.dev

# JWT Configuration
JWT_SECRET=your_secure_jwt_secret_at_least_32_characters

# Admin Configuration
ADMIN_TOKEN=your_secure_admin_token
```

### 3. OAuth Flow

Your OAuth flow will work as follows:

1. **Get OAuth URL**:
   ```http
   POST https://api.engageautomations.com/api/v1/auth/oauth/url
   Content-Type: application/json
   
   {
     "company_id": "your_company_id",
     "redirect_url": "https://api.engageautomations.com/api/v1/auth/oauth/callback"
   }
   ```

2. **Response**:
   ```json
   {
     "auth_url": "https://api.nango.dev/oauth/authorize?...",
     "state": "uuid-state-parameter",
     "expires_at": 1234567890
   }
   ```

3. **User Authorization**: Direct users to the `auth_url`

4. **Callback Handling**: Nango will redirect to:
   ```
   https://api.engageautomations.com/api/v1/auth/oauth/callback?code=auth_code&state=state_param
   ```

5. **Token Exchange**: Your API automatically processes the callback and returns:
   ```json
   {
     "message": "Authorization successful",
     "company_id": "company_123",
     "access_token": "jwt_token",
     "token_type": "Bearer",
     "expires_in": 3600,
     "locations_synced": 5
   }
   ```

## 🔐 Security Features

### State Parameter Validation
- Each OAuth request generates a unique state parameter
- State is cached for 10 minutes for security validation
- Prevents CSRF attacks

### JWT Token Generation
- Secure JWT tokens for API access
- 1-hour expiration by default
- Contains company_id and user context

### Automatic Token Refresh
- Background jobs refresh tokens before expiration
- Webhook support for real-time token updates
- Comprehensive error handling and retry logic

## 🚀 Testing Your Integration

### 1. Health Check
```bash
curl https://api.engageautomations.com/health
```

### 2. Get OAuth URL
```bash
curl -X POST https://api.engageautomations.com/api/v1/auth/oauth/url \
  -H "Content-Type: application/json" \
  -d '{
    "company_id": "test_company_123"
  }'
```

### 3. Test Protected Endpoint
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  https://api.engageautomations.com/api/v1/companies/test_company_123
```

## 🔧 Configuration Checklist

- [ ] Domain `api.engageautomations.com` is properly configured in Railway
- [ ] SSL certificate is active and valid
- [ ] Nango dashboard has the correct redirect URI
- [ ] Environment variables are set in Railway
- [ ] Database is connected and migrations are complete
- [ ] Redis is connected for caching and rate limiting
- [ ] Health endpoints are responding correctly

## 🐛 Troubleshooting

### Common Issues

1. **"Invalid redirect URI" error**:
   - Ensure Nango dashboard has: `https://api.engageautomations.com/api/v1/auth/oauth/callback`
   - Check for trailing slashes or typos

2. **"Invalid client_id" error**:
   - Verify `NANGO_PUBLIC_KEY` environment variable
   - Check Nango dashboard for correct public key

3. **"Invalid state parameter" error**:
   - State expires after 10 minutes
   - Ensure Redis is connected for state caching

4. **JWT token issues**:
   - Verify `JWT_SECRET` is set and secure
   - Check token expiration (default 1 hour)

### Debug Endpoints

```bash
# Check system health
curl https://api.engageautomations.com/api/v1/status

# Check admin health (requires admin token)
curl -H "X-Admin-Token: YOUR_ADMIN_TOKEN" \
  https://api.engageautomations.com/api/v1/admin/system/health
```

## 📚 Next Steps

1. **Configure your client application** to use the OAuth flow
2. **Set up webhooks** for real-time token updates
3. **Implement proper error handling** in your client
4. **Monitor token refresh** jobs and health endpoints
5. **Set up logging and monitoring** for production use

## 🔗 Related Documentation

- [Main README](README.md) - Complete API documentation
- [Setup Guide](SETUP_GUIDE.md) - Local development setup
- [Railway Documentation](https://docs.railway.app) - Deployment help
- [Nango Documentation](https://docs.nango.dev) - OAuth integration details

---

**Your Directory Engine API is now ready for OAuth integration with your custom domain!**