# Architecture Decision Records

This document captures notable architecture decisions made in the Synnergy Network.

## ADR-0001: Go Monorepo Structure

**Status**: Accepted 2024-06-07

**Context**: Synnergy spans core node services, CLI tools and multiple GUIs. Maintaining them in a single repository simplifies coordination but can complicate dependency management.

**Decision**: Use a monorepo rooted at `synnergy-network` for all Go packages, GUIs and supporting scripts. Shared utilities live under `internal/` and documentation under `docs/`.

**Consequences**:

- Easier cross-module refactoring and unified issue tracking
- Requires clear contribution guidelines and staged workflow
- Builds and tests must scope to affected packages to maintain performance

## ADR-0002: Command-Oriented CLI

**Status**: Accepted 2024-06-07

**Context**: The CLI exposes numerous features: ledger operations, networking, AI integration and more. A consistent structure is needed for discoverability.

**Decision**: Organize CLI commands into modular groups under `cmd/` with each file registering a command set. Shared logic lives in `internal/` packages.

**Consequences**:

- Users can discover functionality via `synnergy help`
- Command groups remain loosely coupled and easier to extend
- Requires thorough documentation and examples for each group


## ADR-0003: Markdown-First Documentation

**Status**: Accepted 2024-06-07

**Context**: The project spans many components and requires consistent documentation for developers and end users. A lightweight approach is needed that works across platforms and integrates with version control.

**Decision**: Author all documentation in Markdown under the `docs/` directory and version it alongside source code. Generate any rendered sites from these files during release builds.

**Consequences**:

- Documentation changes follow the same review process as code
- Allows tooling such as static site generators or linters to operate on docs
- Contributors must keep Markdown files formatted and up to date

## ADR-0004: Versioned Documentation Releases

**Status**: Accepted 2024-06-07

**Context**: As the codebase matures, documentation must remain synchronized with releases and use a consistent style. Stakeholders require archived copies of docs for each version.

**Decision**: Version the `docs/` directory alongside code tags. Each release publishes a snapshot of the Markdown files. Continuous integration runs Markdown linters to enforce the [Google developer documentation style guide](https://developers.google.com/style).

**Consequences**:

- Documentation can be browsed per release tag.
- Style consistency improves readability and translation.
- CI gains additional checks and dependencies.
