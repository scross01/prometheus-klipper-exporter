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

For detailed information about the project structure, build commands, test
environment, and implementation patterns, see the [Developers Guide](docs/developers/).

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
---------------------

### Architecture Overview

The project follows these key architectural patterns:

- **Main Entry**: `main.go` handles HTTP server and routing
- **Collector Pattern**: `collector/collector.go` implements Prometheus Collector interface
- **Module System**: Each collector file handles specific Klipper API endpoints
- **Documentation Site**: `docs/` is a [VitePress](https://vitepress.dev/) site published to GitHub Pages

For detailed architecture and implementation guidance, refer to the
[Developers Guide](docs/developers/).

### Coding Standards

#### Go Code Style

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Keep functions focused and small
- Follow the existing code patterns and conventions

#### Markdown Style

- Follow `.markdownlint.json` rules
- Use setext headers for main sections (underlined with `=` or `-`)
- Use ATX headers for subsections (prefixed with `#`)
- Keep line length reasonable (MD013 rule)

### Naming Conventions

- **Prometheus Metrics**: Use `klipper_*` prefix with snake_case naming

For full details on metric naming, shared utilities, and module registration
patterns, see the [Developers Guide](docs/developers/#collector-implementation-guide).

### Validation

Before submitting your changes:

1. Run all tests to ensure nothing is broken: `make test`
2. Verify the code builds without warnings: `make build`
3. Check formatting: `make fmt`
4. If docs were changed, build the site to verify: `cd docs && npm run build`
5. Test manually using the [virtual printer environment](docs/developers/#virtual-printer-test-environment)

Documentation
-------------

### Documentation Site

The project uses [VitePress](https://vitepress.dev/) for its documentation site,
published to GitHub Pages at `https://scross01.github.io/prometheus-klipper-exporter`.

The docs live in the `docs/` directory.

#### Building and previewing locally

```sh
cd docs
npm install
npm run dev     # Start the dev server (hot-reload)
npm run build   # Build the static site
npm run preview # Preview the built site
```

#### Project documentation

- Keep README.md up to date with new features
- Update module documentation when adding new metrics
- Create or update metric reference pages in `docs/metrics/`
- Register any new metric pages in the sidebar at `docs/.vitepress/config.js`
- Add examples for new configuration options
- Follow `.markdownlint.json` rules (`setext_with_atx` headers, reasonable line length)

For documentation contribution details, see the [Developers Guide](docs/developers/#documentation-site).

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
