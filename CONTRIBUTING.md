Contributing to Prometheus Klipper Exporter
==========================================

Thank you for your interest in contributing to the Prometheus Klipper Exporter 
project! This guide will help you understand how to contribute effectively.

Getting Started
---------------

### Setting Up Your Development Environment

1. Fork the repository on GitHub
2. Clone your fork locally:

   ```sh
   git clone https://github.com/your-username/prometheus-klipper-exporter.git
   cd prometheus-klipper-exporter
   ```

3. Install dependencies:

   ```sh
   go mod download
   ```

4. Build the project:

   ```sh
   make build
   ```

How to Contribute
-----------------

### Reporting Issues

Before reporting an issue, please:

1. Check existing issues to avoid duplicates
2. Provide clear steps to reproduce the issue
3. Include relevant logs and error messages
4. Specify your environment (OS, Go version, Klipper version)

### Submitting Pull Requests

1. Create a feature branch from `main`:

   ```sh
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following the project's coding standards
3. Test your changes thoroughly
4. Commit your changes with clear, descriptive messages
5. Push to your fork and submit a pull request

### Code Review Process

- All contributions will be reviewed by maintainers
- Be prepared to make requested changes
- Keep pull requests focused on a single feature or bug fix
- Large changes should be discussed in an issue first

Development Guidelines
----------------------

### Architecture Overview

The project follows these key architectural patterns:

- **Main Entry**: `main.go` handles HTTP server and routing
- **Collector Pattern**: `collector/collector.go` implements Prometheus Collector interface
- **Module System**: Each collector file handles specific Klipper API endpoints

### Coding Standards

#### Go Code Style

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Keep functions focused and small
- Add comments for complex logic
- Follow the existing code patterns and conventions

#### Markdown Style

- Follow `.markdownlint.json` rules
- Use setext headers for main sections (underlined with `=` or `-`)
- Use ATX headers for subsections (prefixed with `#`)
- Keep line length reasonable (MD013 rule)

### Key Patterns to Follow

1. **Metric Name Sanitization**: Use `getValidLabelName()` from collector.go for Prometheus metric names
2. **Boolean Conversion**: Use `boolToFloat64()` for converting booleans to 0/1 values
3. **Module Registration**: Check `slices.Contains(c.modules, "module_name")` before collecting metrics
4. **API Response Handling**: Implement `fetchMoonraker*` functions for HTTP requests and JSON parsing
5. **Error Handling**: Return early on errors but log them first using `log.Error(err)`

### Naming Conventions

- **Prometheus Metrics**: Use `klipper_*` prefix with snake_case naming

## Testing

### Running Tests

```sh
# Run all tests
make test

# Run specific test file
go test ./tests/label_sanitization_test.go -v
```

### Writing Tests

- Add tests for new any utility functions
- Follow existing test patterns

### Validation

Before submitting your changes:

1. Run all tests to ensure nothing is broken
2. Verify the code builds without warnings
3. Check formatting with `make fmt`
4. Test manually if applicable

Documentation
-------------

### Updating Documentation

- Keep README.md up to date with new features
- Update module documentation when adding new metrics
- Add examples for new configuration options
- Follow markdown linting rules

Community Guidelines
--------------------

### Code of Conduct

- Be respectful and professional
- Provide constructive feedback
- Be open to different perspectives
- Follow GitHub's Community Guidelines

### Communication

- Use GitHub issues for bug reports and feature requests
- Use pull requests for code contributions
- Be responsive to feedback and questions

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Klipper Documentation](https://www.klipper3d.org/)
- [Moonraker API Documentation](https://moonraker.readthedocs.io/)

Thank you for contributing to the Prometheus Klipper Exporter project!
