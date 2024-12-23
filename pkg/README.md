# EC2 Manager Packages

This directory contains the core packages for the EC2 Manager tool. Each package is designed with a specific responsibility to maintain clean separation of concerns.

## Package Structure

### ami/
Core functionality for AMI-related operations including:
- Instance migration
- AMI creation and management
- Volume snapshot handling
- Tag management

### client/
AWS client handling and configuration:
- EC2 client initialization
- AWS credential management
- Mock client support for testing
- Client interface definitions

### config/
Configuration management:
- AWS configuration loading
- Environment variable handling
- Default settings

### logger/
Logging utilities:
- Structured logging setup
- Log level management
- Context-aware logging

### testutil/
Testing utilities:
- Mock client setup
- Test helper functions
- Common test fixtures

### types/
Common types and interfaces:
- EC2 client interfaces
- Shared data structures
- Type definitions

## Design Principles

1. **Separation of Concerns**
   - Each package has a single, well-defined responsibility
   - Minimal dependencies between packages
   - Clear interfaces between components

2. **Testability**
   - All packages are designed with testing in mind
   - Mock interfaces for external dependencies
   - Helper functions to simplify test setup

3. **Error Handling**
   - Consistent error handling patterns
   - Detailed error messages
   - Error wrapping for context

4. **Configuration**
   - Configuration is centralized
   - Environment-based configuration
   - Sensible defaults

5. **Logging**
   - Structured logging throughout
   - Consistent log levels
   - Context-aware logging

## Package Dependencies

```
cmd/
 └─► ami/
     └─► client/
         └─► types/
             └─► config/
                 └─► logger/
```

## Adding New Packages

When adding a new package:
1. Create a clear package description in comments
2. Define public interfaces
3. Add comprehensive tests
4. Update this README
5. Consider impact on existing packages
