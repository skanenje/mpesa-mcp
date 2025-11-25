# Code Restructuring Migration Guide

## New Project Structure

The codebase has been restructured to follow Go best practices with clear separation of concerns:

```
mpesa-mcp/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── mpesa/
│   │   ├── client.go            # M-Pesa API client (core)
│   │   ├── auth.go              # OAuth authentication
│   │   ├── stk_push.go          # STK Push operations
│   │   ├── qr_code.go           # QR code generation
│   │   └── types.go             # Shared data types
│   ├── mcp/
│   │   ├── server.go            # MCP server initialization
│   │   ├── tools.go             # MCP tool handlers
│   │   └── prompts.go           # MCP prompt handlers
│   └── utils/
│       └── phone.go             # Utility functions
├── go.mod
├── go.sum
├── .env
├── .env.example
├── .gitignore
└── README.md
```

## Key Improvements

### 1. **Separation of Concerns**
- **Config Layer**: Environment variable loading and validation
- **M-Pesa Layer**: All M-Pesa API interactions isolated
- **MCP Layer**: MCP server setup and handlers
- **Utils Layer**: Reusable utility functions

### 2. **Thread Safety**
- Access token management now uses `sync.RWMutex` for concurrent access
- Safe token refresh in background goroutine

### 3. **Better Testability**
- Each package can be tested independently
- Dependencies are injected (e.g., `mpesa.Client` into `mcp.Server`)

### 4. **Standard Go Layout**
- `cmd/` for application entry points
- `internal/` for private application code
- Clear package boundaries

## Running the Application

### Old way:
```bash
go run *.go
```

### New way:
```bash
go run cmd/server/main.go
```

Or build and run:
```bash
go build -o mpesa-mcp cmd/server/main.go
./mpesa-mcp
```

## What Changed

### Configuration
- Moved from `AppContext` in main.go to dedicated `config.Config`
- Validation logic separated

### M-Pesa Client
- All API operations now methods on `mpesa.Client`
- Token management encapsulated with thread-safe access
- HTTP client reused across requests

### MCP Server
- Server setup separated from tool/prompt registration
- Clean initialization in `mcp.NewServer()`

## Old Files (Can be deleted)
- `main.go` → `cmd/server/main.go`
- `auth.go` → `internal/mpesa/auth.go`
- `stk_push.go` → `internal/mpesa/stk_push.go`
- `qr_code.go` → `internal/mpesa/qr_code.go`
- `tools.go` → `internal/mcp/tools.go`
- `prompts.go` → `internal/mcp/prompts.go`
