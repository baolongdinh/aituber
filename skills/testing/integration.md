# Integration & E2E Testing Skill

## TRIGGER
Read this file when implementing integration tests, E2E tests, or comprehensive testing strategies beyond unit tests.

---
--- 

## ⚡ RAPID TESTING DEVELOPMENT (PRO-LEVEL) 

Use the automation script `scripts/e2e-scaffold.js` to quickly generate a Playwright E2E test file for a new domain. 

**Command:** 
```bash 
node scripts/e2e-scaffold.js --domain=<name> 
``` 


## IDENTITY
You are a Senior QA Engineer specializing in comprehensive testing strategies.
You ensure applications work correctly across all layers and in real-world scenarios.
You never skip integration testing - unit tests alone are insufficient for production readiness.

---

## 🧪 INTEGRATION TESTING (Backend)

### Integration Test Setup
```go
// tests/integration/integration_test.go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    
    "github.com/stretchr/testify/suite"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
    "github.com/ory/dockertest"
    
    "your-project/internal/database"
    "your-project/internal/router"
    "your-project/pkg/config"
)

type IntegrationTestSuite struct {
    suite.Suite
    router       *gin.Engine
    dbContainer testcontainers.Container
    dbConfig    *database.Config
}

func (suite *IntegrationTestSuite) SetupSuite() {
    // Setup test database container
    suite.setupTestDatabase()
    
    // Setup application
    suite.setupApplication()
}

func (suite *IntegrationTestSuite) TearDownSuite() {
    if suite.dbContainer != nil {
        suite.dbContainer.Terminate(context.Background())
    }
}

func (suite *IntegrationTestSuite) setupTestDatabase() {
    // Use Docker test container for isolated database
    pool, err := dockertest.NewPool("")
    suite.Require().NoError(err)
    
    // Pull postgres image
    resource, err := pool.Run("postgres", "15-alpine", []string{
        "POSTGRES_PASSWORD=test",
        "POSTGRES_DB=test_db",
        "POSTGRES_USER=test_user",
    })
    suite.Require().NoError(err)
    
    // Wait for database to be ready
    err = pool.Retry(func() error {
        db, err := sql.Open("postgres", "postgres://test_user:test@localhost/test_db?sslmode=disable")
        if err != nil {
            return err
        }
        return db.Ping()
    })
    suite.Require().NoError(err)
    
    suite.dbContainer = &testcontainers.Container{}
    suite.dbConfig = &database.Config{
        Host:     "localhost",
        Port:     resource.GetPort("5432/tcp"),
        User:     "test_user",
        Password: "test",
        DBName:   "test_db",
        SSLMode:  "disable",
    }
}

func (suite *IntegrationTestSuite) setupApplication() {
    // Load test configuration
    cfg := &config.Config{
        Database: *suite.dbConfig,
        JWT: config.JWTConfig{
            Secret:     "test-secret-key",
            ExpiresIn:  "24h",
        },
        Server: config.ServerConfig{
            Port: "8080",
            Mode: "test",
        },
    }
    
    // Initialize database
    db, err := database.NewConnection(cfg.Database)
    suite.Require().NoError(err)
    
    // Run migrations
    err = database.RunMigrations(db, "file://../migrations")
    suite.Require().NoError(err)
    
    // Setup router
    suite.router = router.Setup(cfg, db)
}

func TestIntegrationSuite(t *testing.T) {
    suite.Run(t, new(IntegrationTestSuite))
}
```

### API Integration Tests
```go
// tests/integration/auth_test.go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

type AuthIntegrationTestSuite struct {
    IntegrationTestSuite
    authToken string
    userID    uint
}

func (suite *AuthIntegrationTestSuite) SetupTest() {
    // Clean up database before each test
    suite.cleanupDatabase()
}

func (suite *AuthIntegrationTestSuite) TestUserRegistration() {
    // Prepare request
    payload := map[string]string{
        "email":    "test@example.com",
        "password": "StrongPassword123!",
        "name":     "Test User",
    }
    
    body, _ := json.Marshal(payload)
    req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    // Execute request
    w := httptest.NewRecorder()
    suite.router.ServeHTTP(w, req)
    
    // Assertions
    assert.Equal(suite.T(), http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(suite.T(), err)
    
    data := response["data"].(map[string]interface{})
    suite.authToken = data["access_token"].(string)
    suite.userID = uint(data["user"].(map[string]interface{})["id"].(float64))
    
    assert.NotEmpty(suite.T(), suite.authToken)
    assert.Greater(suite.T(), suite.userID, uint(0))
}

func (suite *AuthIntegrationTestSuite) TestUserLogin() {
    // First register a user
    suite.TestUserRegistration()
    
    // Prepare login request
    payload := map[string]string{
        "email":    "test@example.com",
        "password": "StrongPassword123!",
    }
    
    body, _ := json.Marshal(payload)
    req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    // Execute request
    w := httptest.NewRecorder()
    suite.router.ServeHTTP(w, req)
    
    // Assertions
    assert.Equal(suite.T(), http.StatusOK, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(suite.T(), err)
    
    data := response["data"].(map[string]interface{})
    token := data["access_token"].(string)
    
    assert.NotEmpty(suite.T(), token)
}

func (suite *AuthIntegrationTestSuite) TestProtectedEndpoint() {
    // Setup authenticated user
    suite.TestUserRegistration()
    
    // Prepare request to protected endpoint
    req := httptest.NewRequest("GET", "/api/user/profile", nil)
    req.Header.Set("Authorization", "Bearer "+suite.authToken)
    
    // Execute request
    w := httptest.NewRecorder()
    suite.router.ServeHTTP(w, req)
    
    // Assertions
    assert.Equal(suite.T(), http.StatusOK, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(suite.T(), err)
    
    data := response["data"].(map[string]interface{})
    assert.Equal(suite.T(), "test@example.com", data["email"])
}

func (suite *AuthIntegrationTestSuite) cleanupDatabase() {
    // Clean up test data
    db, _ := database.NewConnection(*suite.dbConfig)
    defer db.Close()
    
    db.Exec("DELETE FROM users WHERE email LIKE '%@example.com'")
}
```

---

## 🎭 END-TO-END TESTING (Frontend)

### Cypress E2E Test Setup
```javascript
// cypress.config.js
const { defineConfig } = require('cypress')

module.exports = defineConfig({
  e2e: {
    baseUrl: 'http://localhost:3000',
    supportFile: 'cypress/support/e2e.js',
    specPattern: 'cypress/e2e/**/*.cy.{js,jsx,ts,tsx}',
    video: true,
    screenshotOnRunFailure: true,
    viewportWidth: 1280,
    viewportHeight: 720,
    env: {
      apiUrl: 'http://localhost:8080'
    },
    setupNodeEvents(on, config) {
      // Plugin for database seeding
      on('task', {
        async seedDatabase({ seedFile }) {
          const { execSync } = require('child_process')
          execSync(`npm run seed:test -- ${seedFile}`, { stdio: 'inherit' })
          return null
        },
        
        async clearDatabase() {
          const { execSync } = require('child_process')
          execSync('npm run db:clear:test', { stdio: 'inherit' })
          return null
        }
      })
    }
  }
})
```

### E2E Test Examples
```javascript
// cypress/e2e/auth.cy.js
describe('Authentication Flow', () => {
  beforeEach(() => {
    cy.task('clearDatabase')
    cy.visit('/')
  })

  it('should allow user to register', () => {
    cy.get('[data-cy=register-button]').click()
    
    // Fill registration form
    cy.get('[data-cy=email-input]').type('test@example.com')
    cy.get('[data-cy=password-input]').type('StrongPassword123!')
    cy.get('[data-cy=name-input]').type('Test User')
    
    // Submit form
    cy.get('[data-cy=submit-button]').click()
    
    // Verify successful registration
    cy.url().should('include', '/dashboard')
    cy.get('[data-cy=user-menu]').should('contain', 'Test User')
    
    // Verify API call was successful
    cy.window().then((win) => {
      expect(win.localStorage.getItem('access_token')).to.exist
    })
  })

  it('should allow user to login', () => {
    // First register a user
    cy.registerUser('test@example.com', 'StrongPassword123!', 'Test User')
    cy.clearCookies()
    
    // Now test login
    cy.get('[data-cy=login-button]').click()
    
    // Fill login form
    cy.get('[data-cy=email-input]').type('test@example.com')
    cy.get('[data-cy=password-input]').type('StrongPassword123!')
    
    // Submit form
    cy.get('[data-cy=submit-button]').click()
    
    // Verify successful login
    cy.url().should('include', '/dashboard')
    cy.get('[data-cy=user-menu]').should('contain', 'Test User')
  })

  it('should show error for invalid credentials', () => {
    cy.get('[data-cy=login-button]').click()
    
    // Fill with invalid credentials
    cy.get('[data-cy=email-input]').type('invalid@example.com')
    cy.get('[data-cy=password-input]').type('wrongpassword')
    
    // Submit form
    cy.get('[data-cy=submit-button]').click()
    
    // Verify error message
    cy.get('[data-cy=error-message]').should('be.visible')
    cy.get('[data-cy=error-message]').should('contain', 'Invalid credentials')
    
    // Verify still on login page
    cy.url().should('include', '/login')
  })
})

// Custom command for user registration
Cypress.Commands.add('registerUser', (email, password, name) => {
  cy.request({
    method: 'POST',
    url: `${Cypress.env('apiUrl')}/api/auth/register`,
    body: { email, password, name }
  })
})
```

### Web3 E2E Tests
```javascript
// cypress/e2e/wallet.cy.js
describe('Wallet Connection', () => {
  beforeEach(() => {
    cy.visit('/')
  })

  it('should connect wallet successfully', () => {
    // Mock wallet connection
    cy.window().then((win) => {
      win.ethereum = {
        request: cy.stub().resolves('0x1234567890123456789012345678901234567890'),
        on: cy.stub(),
        removeListener: cy.stub()
      }
    })

    // Click connect wallet button
    cy.get('[data-cy=connect-wallet-button]').click()
    
    // Verify wallet connection
    cy.get('[data-cy=wallet-address]').should('be.visible')
    cy.get('[data-cy=wallet-address]').should('contain', '0x1234...7890')
    
    // Verify network check
    cy.get('[data-cy=network-badge]').should('contain', 'AIOZ')
  })

  it('should show error for unsupported network', () => {
    // Mock unsupported network
    cy.window().then((win) => {
      win.ethereum = {
        request: cy.stub().resolves('1'), // Ethereum mainnet
        on: cy.stub(),
        removeListener: cy.stub()
      }
    })

    cy.get('[data-cy=connect-wallet-button]').click()
    
    // Verify network error
    cy.get('[data-cy=error-message]').should('be.visible')
    cy.get('[data-cy=error-message]').should('contain', 'Please connect to AIOZ Network')
  })
})
```

---

## 🔧 TEST AUTOMATION

### Test Database Seeding
```go
// tests/seeds/seeds.go
package seeds

import (
    "context"
    "database/sql"
    "encoding/json"
    "time"
    
    "your-project/internal/model"
    "golang.org/x/crypto/bcrypt"
)

type SeedData struct {
    Users []UserSeed `json:"users"`
    NFTs  []NFTSeed  `json:"nfts"`
}

type UserSeed struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    Name     string `json:"name"`
    Role     string `json:"role"`
}

type NFTSeed struct {
    Name        string  `json:"name"`
    Description string  `json:"description"`
    ImageURL    string  `json:"image_url"`
    Price       float64 `json:"price"`
    OwnerID     uint    `json:"owner_id"`
}

func SeedDatabase(db *sql.DB, seedFile string) error {
    // Read seed data
    data, err := ioutil.ReadFile(seedFile)
    if err != nil {
        return err
    }
    
    var seedData SeedData
    if err := json.Unmarshal(data, &seedData); err != nil {
        return err
    }
    
    // Seed users
    for _, userSeed := range seedData.Users {
        hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(userSeed.Password), bcrypt.DefaultCost)
        
        query := `INSERT INTO users (email, password, name, role, created_at, updated_at) 
                  VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
        
        err := db.QueryRow(query, userSeed.Email, string(hashedPassword), userSeed.Name, 
                          userSeed.Role, time.Now(), time.Now()).Scan(&userSeed.ID)
        if err != nil {
            return err
        }
    }
    
    // Seed NFTs
    for _, nftSeed := range seedData.NFTs {
        query := `INSERT INTO nfts (name, description, image_url, price, owner_id, created_at, updated_at) 
                  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
        
        err := db.QueryRow(query, nftSeed.Name, nftSeed.Description, nftSeed.ImageURL,
                          nftSeed.Price, nftSeed.OwnerID, time.Now(), time.Now()).Scan(&nftSeed.ID)
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

### Test Utilities
```javascript
// cypress/support/utils.js
export const testUtils = {
  // Generate test data
  generateUser() {
    return {
      email: `test${Date.now()}@example.com`,
      password: 'TestPassword123!',
      name: 'Test User'
    }
  },
  
  // Mock API responses
  mockApiResponse(endpoint, response, status = 200) {
    cy.intercept('GET', endpoint, {
      statusCode: status,
      body: response
    }).as(endpoint)
  },
  
  // Wait for loading
  waitForLoad(selector = '[data-cy=loading]', timeout = 5000) {
    cy.get(selector).should('not.exist')
  },
  
  // Login helper
  login(email, password) {
    cy.visit('/login')
    cy.get('[data-cy=email-input]').type(email)
    cy.get('[data-cy=password-input]').type(password)
    cy.get('[data-cy=submit-button]').click()
    cy.url().should('include', '/dashboard')
  },
  
  // Cleanup helper
  cleanup() {
    cy.clearCookies()
    cy.clearLocalStorage()
    cy.window().then((win) => {
      win.sessionStorage.clear()
    })
  }
}

// Make available globally
beforeEach(() => {
  cy.task('clearDatabase')
})
```

---

## 📊 TEST REPORTING

### Coverage Reporting
```yaml
# .github/workflows/test-coverage.yml
name: Test Coverage

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test-coverage:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_DB: test_db
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'
          
      - name: Run tests with coverage
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html
          
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
          flags: unittests
          
      - name: Run integration tests
        run: |
          go test -v -tags=integration ./tests/integration/...
          
      - name: Run E2E tests
        run: |
          npm run test:e2e
```

### Test Reports
```javascript
// cypress/plugins/report.js
const { defineConfig } = require('cypress')
const cucumber = require('cypress-cucumber-preprocessor').default

module.exports = defineConfig({
  e2e: {
    specPattern: '**/*.feature',
    setupNodeEvents(on, config) {
      on('file:preprocessor', cucumber({
        stepDefinitions: 'cypress/support/step_definitions/**/*.js'
      }))
      
      // Mochawesome reporting
      on('after:run', (results) => {
        const { merge } = require('mochawesome-merge')
        const { createReport } = require('mochawesome-report-generator')
        
        merge({
          files: results.config.reporterOptions.reportDir + '/mochawesome*.json'
        }).then((report) => {
          createReport({
            reportDir: 'mochawesome-report',
            reportJsonFile: 'mochawesome.json',
            reportHtmlFile: 'mochawesome.html',
            reportTitle: 'E2E Test Report'
          })
        })
      })
    },
    reporter: 'mochawesome',
    reporterOptions: {
      reportDir: 'mochawesome-report',
      overwrite: false,
      html: false,
      json: true
    }
  }
})
```

---

## 📋 TESTING CHECKLIST

### Pre-Release Testing Checklist
```
Unit Tests:
[ ] All service methods have unit tests
[ ] Edge cases are covered
[ ] Error scenarios are tested
[ ] Mock dependencies are properly isolated
[ ] Test coverage is >80%

Integration Tests:
[ ] API endpoints work with real database
[ ] Authentication flows are tested
[ ] External service integrations work
[ ] Database migrations work correctly
[ ] Error handling works end-to-end

E2E Tests:
[ ] Critical user journeys are tested
[ ] Web3 wallet connections work
[ ] Responsive design works on all devices
[ ] Error states are handled gracefully
[ ] Performance is acceptable

Security Tests:
[ ] Input validation prevents attacks
[ ] Authentication is secure
[ ] Authorization works correctly
[ ] Rate limiting is effective
[ ] Security headers are present

Performance Tests:
[ ] Load testing is performed
[ ] Database queries are optimized
[ ] Frontend bundle size is optimized
[ ] API response times are acceptable
[ ] Memory usage is within limits
```

---

## DO / DON'T

✅ **DO**
- Test at all levels: unit, integration, E2E
- Use test databases that mirror production
- Test both happy paths and error scenarios
- Automate test execution in CI/CD
- Monitor test coverage and quality
- Use realistic test data
- Test security vulnerabilities
- Test performance under load

❌ **DON'T**
- NEVER rely only on unit tests
- NEVER mock everything in integration tests
- NEVER skip testing error scenarios
- NEVER ignore flaky tests
- NEVER test against production data
- NEVER commit without tests passing
- NEVER ignore test coverage reports
- NEVER skip E2E tests for critical flows
