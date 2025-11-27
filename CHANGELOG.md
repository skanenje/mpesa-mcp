# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Fixed
- **Critical**: Fixed "Wrong credentials" error (500.001.1001) in STK Push
  - Updated `.env` to use raw passkey instead of base64-encoded passkey
  - The M-Pesa API requires: `Base64(Shortcode + RawPasskey + Timestamp)`
  - Previously was double-encoding by using already-encoded passkey
  - Sandbox passkey now correctly set to: `bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919`

### Added
- Optional `account_reference` and `transaction_desc` parameters to `stk_push` tool
- Field length validation (account_reference max 12 chars, transaction_desc max 13 chars)
- `InitiateSTKPushWithOptions` method for more flexible STK Push requests
- Comprehensive passkey configuration guide (`docs/PASSKEY_GUIDE.md`)
- Test script for verifying fixed credentials (`scripts/test-stk-fixed.sh`)

### Changed
- Updated `.env.example` with correct passkey format and documentation
- Enhanced troubleshooting guide with passkey error section
- Updated README with passkey configuration warning

### Documentation
- Added `docs/PASSKEY_GUIDE.md` - Complete guide to passkey configuration
- Updated `docs/TROUBLESHOOTING.md` - Added "Wrong credentials" error section
- Updated `README.md` - Added passkey configuration reference

## [1.0.0] - Initial Release

### Added
- M-Pesa MCP Server implementation in Go
- STK Push payment initiation
- QR Code generation
- OAuth token management
- Callback handling for payment notifications
- SSE and HTTP transport support
- Comprehensive documentation
- Docker support
- Example integrations (Python, Google ADK)
