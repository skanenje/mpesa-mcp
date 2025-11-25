# Architecture Overview

## Design Principles

This codebase follows these key principles:

1. **Separation of Concerns**: Each package has a single, well-defined responsibility
2. **Dependency Injection**: Dependencies are passed explicitly, making testing easier
3. **Thread Safety**: Shared state (access token) is protected with mutexes
4. **Standard Go Layout**: Follows community conventions for project structure

## Package Structure

### `cmd/server`
**Purpose**: Application entry point

- Initializes configuration
- Creates M-Pesa client
- Starts token refresh goroutine
- Initializes and runs MCP server

**Dependencies**: `config`, `mpesa`, `mcp`

### `internal/config`
**Purpose**: Configuration management

- Loads environment variables
- Validates required configuration
- Provides immutable config struct

**Dependencies**: None (only stdlib and godotenv)

### `internal/mpesa`
**Purpose**: M-Pesa API client and operations

**Files**:
- `client.go`: Core client with HTTP client and token management
- `auth.go`: OAuth token refresh logic
- `stk_push.go`: STK Push payment implementation
- `qr_code.go`: QR code generation
- `types.go`: Request/response data structures

**Key Features**:
- Thread-safe token access using `sync.RWMutex`
- Automatic token refresh every 50 minutes
- Reusable HTTP client with timeout
- Clean separation of API operations

**Dependencies**: `config`, `utils`

### `internal/mcp`
**Purpose**: MCP server setup and handlers

**Files**:
- `server.go`: MCP server initialization and lifecycle
- `tools.go`: Tool registration and handlers
- `prompts.go`: Prompt registration and handlers

**Key Features**:
- Wraps MCP SDK with M-Pesa functionality
- Clean handler functions for each tool
- Structured input validation via JSON schema

**Dependencies**: `mpesa`, MCP SDK

### `internal/utils`
**Purpose**: Shared utility functions

**Files**:
- `phone.go`: Phone number formatting

**Dependencies**: None (only stdlib)

## Data Flow

### Initialization Flow
```
main()
  ├─> config.Load()
  │     └─> Validate environment variables
  ├─> mpesa.NewClient(config)
  │     └─> Create HTTP client
  ├─> mpesaClient.StartTokenRefresh()
  │     ├─> Get initial token
  │     └─> Start refresh goroutine
  └─> mcp.NewServer(mpesaClient)
        ├─> Initialize MCP server
        ├─> Register tools
        ├─> Register prompts
        └─> Run server
```

### STK Push Flow
```
MCP Client Request
  └─> mcp.Server.stk_push handler
        └─> mpesa.Client.InitiateSTKPush()
              ├─> Format phone number (utils)
              ├─> Generate password & timestamp
              ├─> Create HTTP request
              ├─> Add Bearer token (thread-safe)
              ├─> Send to Daraja API
              └─> Parse & return response
```

### Token Refresh Flow
```
Background Goroutine (every 50 min)
  └─> mpesa.Client.RefreshToken()
        ├─> Create Basic Auth header
        ├─> Request new token from Daraja
        ├─> Parse response
        └─> Update token (thread-safe with mutex)
```

## Thread Safety

### Access Token Management
The access token is shared between:
- Token refresh goroutine (writer)
- API request handlers (readers)

**Solution**: `sync.RWMutex` in `mpesa.Client`
- `GetAccessToken()`: Acquires read lock
- `setAccessToken()`: Acquires write lock
- Multiple readers can access simultaneously
- Writers have exclusive access

## Testing Strategy

### Unit Tests (Recommended)
- `config`: Test validation logic
- `utils`: Test phone number formatting
- `mpesa`: Mock HTTP client, test request building
- `mcp`: Mock M-Pesa client, test handlers

### Integration Tests
- Test against Daraja sandbox API
- Verify token refresh mechanism
- Test full STK Push flow

## Extension Points

### Adding New M-Pesa Operations
1. Add types to `internal/mpesa/types.go`
2. Implement method on `mpesa.Client`
3. Add tool handler in `internal/mcp/tools.go`
4. Optionally add prompt in `internal/mcp/prompts.go`

### Adding Configuration
1. Add field to `config.Config`
2. Update `Load()` to read env var
3. Update `validate()` if required

## Performance Considerations

- **HTTP Client Reuse**: Single client with connection pooling
- **Token Caching**: Avoid unnecessary auth requests
- **Goroutine Management**: Single refresh goroutine, no leaks
- **Context Propagation**: Proper cancellation support

## Security Considerations

- **Credentials**: Never logged or exposed in errors
- **Token Storage**: In-memory only, not persisted
- **HTTPS**: All API calls use TLS
- **Input Validation**: Phone numbers sanitized
- **Error Messages**: Don't leak sensitive info
