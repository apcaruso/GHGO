# Architecture

ghgo is being reshaped into a frontend-agnostic backend engine for greenhouse gas reporting.

The current implementation uses Go and SQLite, with embedded migrations and a normalized DEFRA/DESNZ 2025 JSON factor pack. The desktop UI has been removed.

API-facing use-case services live in `internal/app`. Storage dependencies are expressed through context-aware interfaces in `internal/ports`; `internal/store` is the current SQLite adapter and owns SQLite-specific opening and migrations. Calculation, input commit, report building, and factor lookup depend on ports instead of the concrete SQLite store.

The next architectural targets are HTTP API adapters, alternate `ports.Store` implementations, and broader data-driven emission factor pack coverage.
