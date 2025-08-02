# Synnergy Documentation

This directory contains Markdown guides and references for developing, operating and using the Synnergy Network.

## Guides

- [Developer Guide](developer-guide.md) – environment setup, coding standards and contribution workflow.
- [CLI User Manual](cli-user-guide.md) – command reference and usage examples.
- [GUI User Manual](gui-user-guide.md) – running and troubleshooting the web interface.
- [API Reference](api-reference.md) – Go package overview with code snippets.
- [Architecture Decision Records](architecture-decisions.md) – catalog of notable technical choices.

## Documentation Process

- All documentation lives in this repository and changes undergo the standard pull request review.
- Run `go fmt` and `go vet` on Go snippets before committing to ensure they compile.
- Release builds generate any rendered sites directly from these Markdown files to keep sources authoritative.
