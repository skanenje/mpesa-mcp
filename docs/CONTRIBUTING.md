# Contributing to M-Pesa MCP Server

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing to the project.

## 🎯 Ways to Contribute

### 1. Code Contributions
- Add new M-Pesa API operations
- Improve error handling
- Add tests
- Optimize performance
- Fix bugs

### 2. Documentation
- Improve README and guides
- Add integration examples
- Create tutorials and videos
- Translate documentation

### 3. Testing
- Test with different agent frameworks
- Report bugs and issues
- Suggest improvements
- Test in production scenarios

### 4. Community
- Answer questions in issues
- Share your use cases
- Write blog posts
- Give talks about the project

## 🚀 Getting Started

### Prerequisites
- Go 1.22+
- Git
- Daraja API sandbox credentials
- Basic understanding of Go and MCP protocol

### Setup Development Environment

```bash
# 1. Fork the repository on GitHub

# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/mpesa-mcp.git
cd mpesa-mcp

# 3. Add upstream remote
git remote add upstream https://github.com/ORIGINAL_OWNER/mpesa-mcp.git

# 4. Create .env file
cp .env.example .env
# Edit .env with your Daraja credentials

# 5. Install dependencies
go mod download

# 6. Run the server
go run cmd/server/main.go

# 7. Test it works
curl http://localhost:8080/health
```

## 📝 Development Workflow

### 1. Create a Branch

```bash
# Update your main branch
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/your-feature-name
# or
git checkout -b fix/bug-description
```

### 2. Make Your Changes

Follow these guidelines:

#### Code Style
- Follow standard Go conventions
- Run `go fmt` before committing
- Use meaningful variable names
- Add comments for complex logic
- Keep functions small and focused

#### Project Structure
```
mpesa-mcp/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── mpesa/          # M-Pesa API client
│   │   ├── client.go
│   │   ├── callback.go     # Callback handling
│   │   └── ...
│   ├── mcp/            # MCP server and handlers
│   ├── transport/      # Transport layer (SSE)
│   └── utils/          # Utility functions
├── examples/           # Integration examples
└── scripts/            # Helper scripts
```

#### Adding a New M-Pesa Operation

1. **Add types** in `internal/mpesa/types.go`:
```go
type NewOperationRequest struct {
    Field1 string `json:"Field1"`
    Field2 int    `json:"Field2"`
}

type NewOperationResponse struct {
    ResponseCode string `json:"ResponseCode"`
    Message      string `json:"Message"`
}
```

2. **Implement method** in `internal/mpesa/` (create new file if needed):
```go
func (c *Client) NewOperation(ctx context.Context, param1 string) (*NewOperationResponse, error) {
    // Implementation
}
```

3. **Add MCP tool** in `internal/mcp/tools.go`:
```go
type NewOperationInput struct {
    Param1 string `json:"param1" jsonschema:"required,description=Description"`
}

// In registerTools():
mcpsdk.AddTool(
    s.mcp,
    &mcpsdk.Tool{
        Name:        "new_operation",
        Description: "Description of what this does",
    },
    func(ctx context.Context, req *mcpsdk.CallToolRequest, input NewOperationInput) (*mcpsdk.CallToolResult, map[string]interface{}, error) {
        // Handler implementation
    },
)
```

4. **Update documentation** in README.md

5. **Add example** in `examples/`

### 3. Test Your Changes

```bash
# Run the server
go run cmd/server/main.go

# Test with curl
curl http://localhost:8080/health

# Test with Python example
cd examples
python python_client.py

# If you added tests (recommended):
go test ./...
```

### 4. Commit Your Changes

```bash
# Stage your changes
git add .

# Commit with a descriptive message
git commit -m "feat: add new M-Pesa operation for X"
# or
git commit -m "fix: resolve issue with token refresh"
# or
git commit -m "docs: improve integration examples"
```

**Commit Message Format:**
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `refactor:` Code refactoring
- `test:` Adding tests
- `chore:` Maintenance tasks

### 5. Push and Create Pull Request

```bash
# Push to your fork
git push origin feature/your-feature-name

# Go to GitHub and create a Pull Request
```

**Pull Request Guidelines:**
- Provide a clear description of changes
- Reference any related issues
- Include screenshots/examples if applicable
- Ensure all checks pass
- Be responsive to feedback

## 🧪 Testing Guidelines

### Manual Testing
1. Start the server
2. Test all affected endpoints
3. Verify with Daraja sandbox
4. Test with example integrations

### Automated Testing (Future)
```bash
# Unit tests
go test ./internal/...

# Integration tests
go test ./tests/integration/...

# Coverage
go test -cover ./...
```

## 📋 Code Review Process

1. **Automated Checks**: CI/CD runs tests and linters
2. **Maintainer Review**: A maintainer reviews your code
3. **Feedback**: Address any requested changes
4. **Approval**: Once approved, your PR will be merged
5. **Celebration**: Your contribution is now part of the project! 🎉

## 🐛 Reporting Bugs

### Before Reporting
- Check if the issue already exists
- Test with the latest version
- Verify it's not a configuration issue

### Bug Report Template
```markdown
**Describe the bug**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce:
1. Start server with '...'
2. Call endpoint '...'
3. See error

**Expected behavior**
What you expected to happen.

**Actual behavior**
What actually happened.

**Environment:**
- OS: [e.g., Ubuntu 22.04]
- Go version: [e.g., 1.22.1]
- Server version: [e.g., v1.0.0]

**Logs**
```
Paste relevant logs here
```

**Additional context**
Any other relevant information.
```

## 💡 Suggesting Features

### Feature Request Template
```markdown
**Is your feature request related to a problem?**
A clear description of the problem.

**Describe the solution you'd like**
What you want to happen.

**Describe alternatives you've considered**
Other solutions you've thought about.

**Use case**
How would this feature be used?

**Additional context**
Any other relevant information.
```

## 📚 Documentation Guidelines

### README Updates
- Keep it concise and scannable
- Use examples liberally
- Update table of contents
- Test all code examples

### Code Comments
```go
// Good: Explains WHY
// RefreshToken gets a new access token because tokens expire every 60 minutes
func (c *Client) RefreshToken(ctx context.Context) error {

// Bad: Explains WHAT (obvious from code)
// RefreshToken refreshes the token
func (c *Client) RefreshToken(ctx context.Context) error {
```

### Example Code
- Must be complete and runnable
- Include error handling
- Add comments explaining key steps
- Test before committing

## 🎨 Style Guide

### Go Code
```go
// Good
func (c *Client) InitiatePayment(ctx context.Context, amount int, phone string) (*Response, error) {
    if amount <= 0 {
        return nil, fmt.Errorf("amount must be positive")
    }
    
    // Format phone number
    formattedPhone := utils.FormatPhoneNumber(phone)
    
    // Build request
    req := buildRequest(amount, formattedPhone)
    
    // Send to API
    return c.sendRequest(ctx, req)
}

// Bad
func (c *Client) InitiatePayment(ctx context.Context, a int, p string) (*Response, error) {
    if a <= 0 { return nil, fmt.Errorf("amount must be positive") }
    fp := utils.FormatPhoneNumber(p)
    req := buildRequest(a, fp)
    return c.sendRequest(ctx, req)
}
```

### Error Messages
```go
// Good: Specific and actionable
return fmt.Errorf("failed to parse phone number %q: must start with 254 or 0", phone)

// Bad: Vague
return fmt.Errorf("invalid input")
```

## 🔒 Security Guidelines

- **Never commit credentials** to `.env` files
- **Validate all inputs** before processing
- **Use HTTPS** in production
- **Sanitize error messages** (don't leak sensitive info)
- **Report security issues privately** to maintainers

## 📄 License

By contributing, you agree that your contributions will be licensed under the MIT License.

## 🤝 Code of Conduct

### Our Pledge
We are committed to providing a welcoming and inclusive environment for all contributors.

### Our Standards
- Be respectful and considerate
- Accept constructive criticism gracefully
- Focus on what's best for the community
- Show empathy towards others

### Unacceptable Behavior
- Harassment or discrimination
- Trolling or insulting comments
- Personal or political attacks
- Publishing others' private information

## 📞 Getting Help

- **Questions**: Open a GitHub issue with the "question" label
- **Discussions**: Use GitHub Discussions for general topics
- **Chat**: Join our community chat (if available)

## 🎉 Recognition

Contributors will be:
- Listed in CONTRIBUTORS.md
- Mentioned in release notes
- Credited in documentation

Thank you for contributing to M-Pesa MCP Server! 🚀
