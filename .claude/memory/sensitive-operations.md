---
name: handle-sensitive-data-best-practices
description: Go best practices for handling sensitive data, configuration, and security in API development
metadata:
  type: feedback | project
---

## 🔐 Sensitive Operations & Security Best Practices for Your BookNest Backend

### 📌 Core Principles

1. **Never hardcode secrets** - All credentials must come from environment variables or secure secret managers (Render Secrets, Docker Secret Files)
2. **Use minimal permissions** - Services should only access the data they need via JWT scopes/roles
3. **Validate all inputs** - Never trust client input; sanitize strings and escape HTML before rendering
4. **Fail securely by default** - Return generic error messages that don't leak system info

---

## 🔑 JWT Configuration & Security (MUST HAVE)

### ✅ Correct Implementation Pattern

```go
// domain/domain.go for key IDs
const (
    CurrentKeyID = "current" // v1 secret: render env var or docker secret file path
    PrevKeyID     = "prev"   // v0 legacy fallback
)

type JWTClaims struct {
    UserID  string `json:"user_id"`
    Email   string `json:"email,omitempty"`
    UserRole int    `json:"user_role"`
}

// middleware/jwt_auth.go for auth middleware (DO NOT MODIFY WITHOUT ENV SETUP)
func LoadJWTConfigFromEnv() (JWTConfig, error) {
    keys := map[string][]byte{
        PrevKeyID:   []byte(os.Getenv("JWT_SECRET_V0")), // legacy support
        CurrentKeyID: []byte(os.Getenv("JWT_SECRET_V1")), // current secret required
    }

    cfg := JWTConfig{}
    for kid, key := range keys {
        if len(key) > 0 {
            cfg.Keys[kid] = key
        }
    }

    return cfg, nil // error if neither env var is set
}

// Test pattern from jwt_auth_test.go:
func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": user.ID.String(),
        "email":   user.Email,
        // optionally add other claims like roles, permissions via custom struct
    })

    if err != nil {
        t.Fatalf("failed to sign token: %v", err)
    }
    
    // Always include kid header for proper key lookup in middleware
}
```

### ❌ NEVER Do This (Antipatterns)

- Hardcoded keys anywhere
- Missing `kid` (key ID) claims that can't be matched to secrets at runtime
- Ignoring token expiration/validity checks  
- Logging full tokens or sensitive data to stdout/stderr

---

## 📁 Environment Variables & Configuration Management

### ✅ Secure Pattern for Production:
```yaml
# Dockerfile - mount secret file instead of using env var directly
COPY docker/secrets/jwt_secrets.conf /run/dropins/certs/secret.jwt # render secrets dir or .env.local
CMD ["./app"] # renders mounts JWT_SECRET_V1 from the conf to os.Getenv("JWT_SECRET_V0")
```

### ❌ Security Risks (Do NOT Do This):

- Committing `.env` files with real credentials anywhere except `/.env.example`, `/docker/secrets/` and local dev-only paths like `/dev/docker/dev.env`:
  ```bash
  # NEVER commit: .env, config.yaml (with secrets) 
  DO commit: /.env.example /docs/config.local.go.sample docker/secrets/*
  ```

- Using default/insecure credentials anywhere on any branch without marking them `local` or requiring user override in `.render.yaml`.  
  **Never** use hardcoded test keys that match prod key IDs like `"current"` and `"prev"`. Always require env var population.

---

## 🧪 Testing Best Practices with JWT Tokens (DO NOT DELETE)

### ✅ Recommended Test Approach:
```go
func mockJWT(signingKey string, kid string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id":   "00000000-0000-0000-0000-000000000001", // test-only ID 
        "email":     "test@example.com" + t.Name(),             // unique per-test
    })
    token.Header["kid"] = kid // required for middleware to find correct key

    return token.SignedString([]byte(signingKey))
}

// TestAIChat uses mock AIs with injected user_id via simple test server
```

### ❌ Don't Remove These:
- The `mockJWT` helper functions in `/internal/http/controller/*_test.go`
- JWT mocking in middleware tests (`jwt_auth_test.go`) — these validate the security layer works correctly under all conditions.

---

## ⚖️ Error Handling & Validation Rules (DO NOT MODIFY)

### ✅ Required Response Formats:

```go
// Chat handler must return sanitized, structured responses from service.Chat():

// 1a. Valid response format for /ai/chat POST success
type AIChatResponse struct { Message string } // sanitize with ai_service.Sanitize() before sending

// Error formats (DO NOT CHAANGE):
- {"error":"validation error"}           // generic validation failure 
- {"error":ai_chat_response_error}       // service-layer business logic errors  
```  

### ❌ Never Accept These Patterns:
- Responses that leak PII like full email/usernames unless explicitly allowed by a scope in claims  
- Returning stack traces or internal IDs (`traceID`, `serverIP`) to clients

---

## 🛠️ AI Service Security Considerations (DO NOT MODIFY WITHOUT AUTH)

```go
// ai_service.go - DO NOT CHAANGE without middleware review:

func Chat(ctx context.Context, input domain.AIChatRequest, userID string) (*domain.AIChatResponse, error) {
    if err := ValidateInput(input); err != nil { return nil, err }     // sanitize before processing
    
    embedding, err := embedService.CreateUserEmbedding(// never store raw user content in embeddings without hashing/obfuscating )
    
    response := ai_service.ChatWithAIProvider(userRole, userID)          // pass role from JWT claims
    if strings.Contains(response.Message, systemInstruction.AIResponseInjection) { return nil, ErrSafetyViolation }
```

**Key Rules:**
1. Validate all user inputs against a whitelist pattern before embedding  
2. Never expose raw AI responses without sanitization (`Sanitize()` filter injection/HTML escape)  
3. Rate-limit requests to prevent abuse of your API endpoints  

---

## 📜 Summary Checklist (DO NOT MODIFY WITHOUT REVIEWING EACH ITEM):
- [ ] JWT_SECRET_V0 env var set in `/.env.local` and Docker secrets file for test runs, but only commit `.env.example`, `/docs/config.*.go.sample`, `docker/secrets/*`  
- [ ] CurrentKeyID `"current"` mapped to v1 secret; PrevKeyID `"prev"` fallback if available (commit these config mappings in domain files)  
- [ ] All sensitive handlers use middleware with proper validation, not direct client access  
- [ ] Error messages are generic (`"validation error"`, `ai_chat_response_error`) — never leak stack traces or internal IDs.  
- [ ] `.env.local` and any local-only configs excluded from git; only example/commit-safe paths included in `.gitignore`.