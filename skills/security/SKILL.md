# Security Best Practices Skill

## TRIGGER
Read this file when implementing security measures, handling authentication, securing APIs, or addressing security concerns.

---
--- 

## ⚡ AUTOMATED SECURITY AUDIT (PRO-LEVEL) 

Before finalizing any task, you SHOULD run the automated security audit script to identify potential vulnerabilities in both Backend (gosec) and Frontend (npm audit). 

**Command:** 
```bash 
python3 scripts/security-audit.py 
``` 


## IDENTITY
You are a Security Engineer focused on building secure, production-ready applications.
You implement defense-in-depth security and never compromise security for convenience.
You assume all external inputs are malicious and all data must be protected.

---

## 🔐 AUTHENTICATION & AUTHORIZATION

### JWT Security Best Practices
```go
// pkg/jwt/jwt.go
package jwt

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
)

type Claims struct {
    UserID   uint   `json:"user_id"`
    Email    string `json:"email"`
    Role     string `json:"role"`
    TokenID  string `json:"token_id"`
    jwt.RegisteredClaims
}

func GenerateToken(userID uint, email, role string) (string, error) {
    tokenID := uuid.New().String()
    
    claims := Claims{
        UserID:  userID,
        Email:   email,
        Role:    role,
        TokenID: tokenID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "aioz-dapp",
            Subject:   fmt.Sprintf("%d", userID),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(getSecret()))
}

func ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(getSecret()), nil
    })

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, err
}

func getSecret() string {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        log.Fatal("JWT_SECRET environment variable is required")
    }
    return secret
}
```

### Authentication Middleware
```go
// internal/middleware/auth_middleware.go
package middleware

import (
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
    "your-project/pkg/jwt"
    "your-project/pkg/response"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            response.Error(c, http.StatusUnauthorized, "Authorization header required")
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := jwt.ValidateToken(tokenString)
        if err != nil {
            response.Error(c, http.StatusUnauthorized, "Invalid token")
            c.Abort()
            return
        }

        // Check if token is blacklisted (for logout)
        if isTokenBlacklisted(claims.TokenID) {
            response.Error(c, http.StatusUnauthorized, "Token has been revoked")
            c.Abort()
            return
        }

        c.Set("user_id", claims.UserID)
        c.Set("email", claims.Email)
        c.Set("role", claims.Role)
        c.Set("token_id", claims.TokenID)
        
        c.Next()
    }
}

func RequireRole(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole, exists := c.Get("role")
        if !exists || userRole != role {
            response.Error(c, http.StatusForbidden, "Insufficient permissions")
            c.Abort()
            return
        }
        c.Next()
    }
}
```

---

## 🛡️ API SECURITY

### Input Validation & Sanitization
```go
// internal/validator/validator.go
package validator

import (
    "regexp"
    "strings"
    "unicode/utf8"
    "github.com/go-playground/validator/v10"
)

type CustomValidator struct {
    validator *validator.Validate
}

func New() *CustomValidator {
    v := validator.New()
    
    // Register custom validation rules
    v.RegisterValidation("nohtml", validateNoHTML)
    v.RegisterValidation("strongpassword", validateStrongPassword)
    v.RegisterValidation("ethaddress", validateEthAddress)
    
    return &CustomValidator{validator: v}
}

func (cv *CustomValidator) ValidateStruct(s interface{}) error {
    return cv.validator.Struct(s)
}

func validateNoHTML(fl validator.FieldLevel) bool {
    field := fl.Field().String()
    htmlTagPattern := `<[a-zA-Z/][^>]*>`
    matched, _ := regexp.MatchString(htmlTagPattern, field)
    return !matched
}

func validateStrongPassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    
    if len(password) < 8 {
        return false
    }
    
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
    hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
    
    return hasUpper && hasLower && hasNumber && hasSpecial
}

func validateEthAddress(fl validator.FieldLevel) bool {
    address := fl.Field().String()
    pattern := `^0x[a-fA-F0-9]{40}$`
    matched, _ := regexp.MatchString(pattern, address)
    return matched
}
```

### Rate Limiting
```go
// internal/middleware/rate_limit_middleware.go
package middleware

import (
    "net/http"
    "sync"
    "time"
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(rps int, burst int) *RateLimiter {
    return &RateLimiter{
        limiters: make(map[string]*rate.Limiter),
        rate:     rate.Limit(rps),
        burst:    burst,
    }
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    limiter, exists := rl.limiters[ip]
    if !exists {
        limiter = rate.NewLimiter(rl.rate, rl.burst)
        rl.limiters[ip] = limiter
    }
    
    return limiter
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()
        limiter := rl.getLimiter(ip)
        
        if !limiter.Allow() {
            response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded")
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### CORS Security
```go
// internal/middleware/cors_middleware.go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
)

func CORSMiddleware() gin.HandlerFunc {
    config := cors.Config{
        AllowOrigins:     []string{"https://yourdomain.com", "https://app.yourdomain.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }
    
    return cors.New(config)
}
```

---

## 🔒 WEB3 SECURITY

### Secure Wallet Connection
```js
// src/composables/useWallet.js
import { ref, computed } from 'vue'
import { useAppKit, useAppKitAccount, useAppKitProvider } from '@reown/appkit/vue'
import { ethers } from 'ethers'

export function useWallet() {
  const { open } = useAppKit()
  const { address, isConnected, chainId } = useAppKitAccount()
  const { walletProvider } = useAppKitProvider()
  
  const isConnecting = ref(false)
  const error = ref(null)
  
  // Validate network
  const isValidNetwork = computed(() => {
    return chainId.value === 168 || chainId.value === 4102 // AIOZ mainnet/testnet
  })
  
  async function connectWallet() {
    try {
      isConnecting.value = true
      error.value = null
      
      await open()
      
      if (!isValidNetwork.value) {
        throw new Error('Please connect to AIOZ Network')
      }
      
      // Verify wallet ownership with signature
      await verifyWalletOwnership()
      
    } catch (err) {
      error.value = err.message
      console.error('Wallet connection failed:', err)
    } finally {
      isConnecting.value = false
    }
  }
  
  async function verifyWalletOwnership() {
    if (!address.value || !walletProvider.value) return
    
    const message = `Sign this message to verify your wallet ownership. Nonce: ${Date.now()}`
    const signature = await walletProvider.value.request({
      method: 'personal_sign',
      params: [message, address.value]
    })
    
    // Verify signature on backend
    const response = await fetch('/api/auth/verify-signature', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        message,
        signature,
        address: address.value
      })
    })
    
    if (!response.ok) {
      throw new Error('Signature verification failed')
    }
  }
  
  return {
    address,
    isConnected,
    chainId,
    isValidNetwork,
    isConnecting,
    error,
    connectWallet
  }
}
```

### Smart Contract Security
```go
// pkg/contract/contract.go
package contract

import (
    "context"
    "math/big"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
)

type ContractManager struct {
    client     *ethclient.Client
    privateKey *ecdsa.PrivateKey
    chainID    *big.Int
}

func NewContractManager(rpcURL, privateKeyHex string) (*ContractManager, error) {
    client, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, err
    }
    
    privateKey, err := crypto.HexToECDSA(privateKeyHex)
    if err != nil {
        return nil, err
    }
    
    chainID, err := client.ChainID(context.Background())
    if err != nil {
        return nil, err
    }
    
    return &ContractManager{
        client:     client,
        privateKey: privateKey,
        chainID:    chainID,
    }, nil
}

func (cm *ContractManager) CreateTransactor() (*bind.TransactOpts, error) {
    nonce, err := cm.client.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(cm.privateKey.PublicKey))
    if err != nil {
        return nil, err
    }
    
    gasPrice, err := cm.client.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    auth, err := bind.NewKeyedTransactorWithChainID(cm.privateKey, cm.chainID)
    if err != nil {
        return nil, err
    }
    
    auth.Nonce = big.NewInt(int64(nonce))
    auth.Value = big.NewInt(0)      // 0 ETH for function calls
    auth.GasLimit = uint64(300000)  // Set appropriate gas limit
    auth.GasPrice = gasPrice
    
    return auth, nil
}
```

---

## 🚨 SECURITY MONITORING

### Security Headers Middleware
```go
// internal/middleware/security_middleware.go
package middleware

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func SecurityHeadersMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
        
        c.Next()
    }
}
```

### Security Event Logging
```go
// pkg/security/audit.go
package security

import (
    "time"
    "github.com/sirupsen/logrus"
)

type SecurityEvent struct {
    Timestamp time.Time `json:"timestamp"`
    Level     string    `json:"level"`
    Event     string    `json:"event"`
    UserID    *uint     `json:"user_id,omitempty"`
    IP        string    `json:"ip"`
    UserAgent string    `json:"user_agent"`
    Details   string    `json:"details,omitempty"`
}

func LogSecurityEvent(level, event, ip, userAgent string, userID *uint, details string) {
    securityEvent := SecurityEvent{
        Timestamp: time.Now(),
        Level:     level,
        Event:     event,
        UserID:    userID,
        IP:        ip,
        UserAgent: userAgent,
        Details:   details,
    }
    
    logger.WithFields(logrus.Fields{
        "security_event": securityEvent,
    }).Log(logrus.WarnLevel, event)
}

func LogFailedLogin(ip, userAgent, email string) {
    LogSecurityEvent("warning", "failed_login", ip, userAgent, nil, "Email: "+email)
}

func LogSuspiciousActivity(ip, userAgent, userID string, details string) {
    uid := parseUserID(userID)
    LogSecurityEvent("critical", "suspicious_activity", ip, userAgent, uid, details)
}
```

---

## 🔍 VULNERABILITY SCANNING

### Dependency Security Scanning
```yaml
# .github/workflows/security.yml
name: Security Scan

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 2 * * 1'  # Weekly on Monday

jobs:
  security-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Run Gosec Security Scanner
        uses: securecodewarrior/github-action-gosec@master
        with:
          args: './...'
          
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'
          
      - name: Upload Trivy scan results
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'
          
      - name: Run npm audit (frontend)
        working-directory: ./frontend
        run: npm audit --audit-level moderate
```

---

## 📋 SECURITY CHECKLIST

### Pre-Deployment Security Checklist
```
[ ] Environment variables are properly configured
[ ] Secrets are not hardcoded in the application
[ ] JWT secrets are strong and regularly rotated
[ ] Rate limiting is implemented on all endpoints
[ ] Input validation is implemented for all user inputs
[ ] SQL injection protection is in place
[ ] XSS protection headers are set
[ ] CORS is properly configured
[ ] HTTPS is enforced in production
[ ] Security headers are implemented
[ ] Logging and monitoring are enabled
[ ] Dependency vulnerability scan is clean
[ ] Authentication flow is secure
[ ] Authorization checks are in place
[ ] Error messages don't leak sensitive information
[ ] Database connections use SSL
[ ] File upload restrictions are in place
[ ] Session management is secure
[ ] Password policies are enforced
[ ] Multi-factor authentication is considered
```

---

## DO / DON'T

✅ **DO**
- Implement defense-in-depth security
- Validate all inputs on both client and server
- Use parameterized queries to prevent SQL injection
- Implement proper error handling without information leakage
- Use HTTPS everywhere in production
- Regularly rotate secrets and API keys
- Implement rate limiting and throttling
- Log security events for monitoring
- Keep dependencies updated
- Use security headers

❌ **DON'T**
- NEVER hardcode secrets in code
- NEVER trust client-side validation only
- NEVER expose sensitive data in error messages
- NEVER use weak encryption algorithms
- NEVER skip input sanitization
- NEVER ignore security warnings
- NEVER store passwords in plain text
- NEVER use default credentials
- NEVER disable security features for convenience
- NEVER ignore security logs and alerts
