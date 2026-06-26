# Architecture

ghgo is being reshaped into a frontend-agnostic backend engine for greenhouse gas reporting.

The current implementation uses Go and SQLite, with embedded migrations and a normalized DEFRA/DESNZ 2025 JSON factor pack. The desktop UI has been removed.

API-facing use-case services live in `internal/app`. `internal/store` is the SQLite adapter and owns opening, migrations, and queries. Calculation, input commit, report building, and factor lookup use that store directly.

The next architectural targets are broader HTTP coverage and data-driven emission factor pack coverage.
