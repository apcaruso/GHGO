# ghgo

ghgo is a local-first backend engine for greenhouse gas (GHG) activity data entry, emissions calculation, and early-stage reporting.

The project is written in Go and uses SQLite for the current local storage implementation. The desktop interface has been removed; the backend now exposes frontend-agnostic app services plus a local HTTP API entry point.

## What The Program Does

ghgo helps collect activity data for an organization and convert that data into emissions results in `kgCO2e` and `tCO2e`.

Main backend capabilities currently present in the codebase:

- Local SQLite database storage with embedded schema migrations.
- Organization, facility, and reporting period setup.
- Reporting period settings, including mobile combustion method selection.
- Spreadsheet-style pasted text parsing and validation.
- Commit workflows for pasted activity data, including server-side re-parsing and replacement of matching active records.
- Editable/replaceable saved activity data by saving replacement input.
- Calculation runs that convert active activity records into emissions results.
- Report table builders for completed calculation runs.
- Local seeding of embedded, versioned emission factor packs during backend initialization.

Supported activity data entry flows:

- Electricity: monthly facility consumption in `kWh`.
- Natural gas: monthly facility consumption in `Smc`.
- Mobile combustion by fuel consumed: fuel type and litres.
- Mobile combustion by distance travelled: vehicle details, fuel type, and `km`.
- Refrigerants: facility gas/substance and `kg` leaked.

Current backend workflow supported by the internal packages:

1. Initialize storage, run migrations, and seed the default factor set.
2. Create an organization, facilities, and a reporting period through backend code.
3. Configure reporting period settings, especially the mobile combustion method if mobile data will be entered.
4. Parse activity data from spreadsheet-style text.
5. Validate parsed rows and fix any blocking errors.
6. Commit the raw pasted data. The backend re-parses it and supersedes matching active records.
7. Run a calculation for the reporting period and factor set.
8. Build report tables for a completed calculation run.

## Conversion Factors

The default factor set visible in the codebase is `DEFRA/DESNZ 2025`.

The application seeds default factors on startup from the normalized JSON factor pack at `factorpacks/defra-2025/factor-pack.json`. The calculation engine then looks up factors stored in SQLite and applies them to active activity records to create calculation results.

Factor packs are checked-in data files, not Go source. Each pack is versioned by:

- `source`: factor authority or publisher, for example `DEFRA`.
- `year`: reporting or publication year.
- `version`: exact pack version used for the database factor set.
- `metadata`: pack-level provenance and operational notes.
- `normalized_rows`: rows already mapped into ghgo's normalized factor schema.

Rows in `normalized_rows` contain the stable emission factor ID, source fields, normalized lookup fields such as `activity_type`, `fuel_type`, `vehicle_type`, `vehicle_size_class`, `substance`, `input_unit`, and the final `factor_value`. The loader validates required fields, compacts metadata JSON, creates the matching factor set when needed, and inserts only missing rows so startup seeding is idempotent.

The seeded factor coverage currently includes:

- UK purchased electricity, using `kgCO2e/kWh`.
- Natural gas, using `kgCO2e/Smc`.
- Mobile fuel combustion for diesel, petrol, LPG, and CNG, using `kgCO2e/L`.
- Vehicle distance factors for cars, vans, and motorbikes by supported size/fuel combinations, using `kgCO2e/km`.
- Refrigerants R134a, R410A, R407C, R404A, R32, and R22, using `kgCO2e/kg`.

The original DEFRA workbook is not checked in. DEFRA import/mapping code remains under `internal/factors` for development workflows and tests, guarded by the `ghgo_devtools` build tag. There is no normal end-user import command in the repository yet.

## Current Limitations

This project is not finished yet.

Known limitations:

- The HTTP API is intentionally minimal and currently covers setup, pasted input parsing/commit, calculations, reports, and factor-set lookup.
- Report tables exist, but polished report generation, printing, and export workflows are not complete.
- Chart dataset tables exist, but graph rendering, printing, and exporting are not implemented.
- Supported activity categories and conversion factors are intentionally limited.
- The current storage implementation is SQLite-only.
- Data entry is based on pasted spreadsheet text rather than full file import workflows or API payloads.
- The architecture is still in transition from the removed desktop frontend toward backend services.

## Planned Work / TODO

- Expand API endpoint coverage and harden API documentation.
- Add report generation/export workflows.
- Add graph dataset/export workflows.
- Add support for more conversion factors.
- Expand data-driven emission factor packs beyond the default DEFRA/DESNZ 2025 pack.

## Usage Instructions

Prerequisites:

- Go compatible with the module declaration in `go.mod` (`go 1.25.0`).
- Network access the first time Go resolves module dependencies.

Install dependencies:

```sh
go mod download
```

Run the HTTP API:

```sh
go run ./cmd/ghgo-api
```

The API command opens the configured database, runs migrations, seeds the default factor set, and serves HTTP until stopped. A no-build browser UI is served from the same origin at:

```text
http://127.0.0.1:8080/ui/
```

By default the API command looks for the UI in `../frontend` or `frontend`, relative to the process working directory. Override this with `GHGO_UI_DIR=/path/to/frontend` when running from another layout.

By default, ghgo stores its local SQLite database at:

```text
./data/ghgo.sqlite
```

Override the database path:

```sh
GHGO_DB_PATH=/custom/path/ghgo.sqlite go run ./cmd/ghgo-api
```

Override the API listen address:

```sh
GHGO_HTTP_ADDR=127.0.0.1:9090 go run ./cmd/ghgo-api
```

HTTP API defaults:

- `GHGO_DB_PATH`: `data/ghgo.sqlite`
- `GHGO_HTTP_ADDR`: `127.0.0.1:8080`

API responses use JSON envelopes:

```json
{ "data": {} }
```

```json
{ "error": { "code": "not_found", "message": "..." } }
```

Minimal API endpoints:

```text
GET  /healthz
GET  /organizations
POST /organizations
GET  /organizations/{id}
GET  /organizations/{organizationID}/facilities
POST /organizations/{organizationID}/facilities
GET  /organizations/{organizationID}/reporting-periods
POST /organizations/{organizationID}/reporting-periods
GET  /reporting-periods/{id}
GET  /reporting-periods/{id}/settings
PUT  /reporting-periods/{id}/settings
GET  /reporting-periods/{id}/facilities/{facilityID}/electricity-settings
PUT  /reporting-periods/{id}/facilities/{facilityID}/electricity-settings
POST /inputs/parse
POST /inputs/commit
POST /reporting-periods/{id}/calculations
GET  /reporting-periods/{id}/calculation-runs
GET  /calculation-runs/{id}
GET  /calculation-runs/{id}/report-tables
GET  /factor-sets
GET  /factor-sets/default
GET  /factor-sets/{id}
```

Example organization create request:

```sh
curl -sS -X POST http://127.0.0.1:8080/organizations \
  -H 'Content-Type: application/json' \
  -d '{"name":"Acme Ltd"}'
```

Run tests:

```sh
go test ./...
```

Build a local binary:

```sh
go build ./cmd/ghgo-api
```

## Repository Structure

- `cmd/ghgo-api`: HTTP API entry point; opens SQLite, runs migrations, seeds default factors, and serves the API.
- `internal/app`: API-facing use-case services and backend bootstrap wiring.
- `internal/httpapi`: standard-library HTTP adapter for `internal/app` services.
- `internal/input`: parsers, validators, normalization, hashing, and commit logic for pasted activity data.
- `internal/calc`: calculation engine that converts active activity records into calculation results.
- `internal/report`: deterministic report table builders for completed calculation runs.
- `internal/factors`: factor-pack loading, DEFRA/DESNZ lookup, parsing, and development import/mapping code.
- `internal/store`: SQLite adapter, queries, migrations runner, and compatibility persistence methods.
- `internal/domain`: core domain types and enums.
- `internal/vocab`: normalized vocabulary for units, months, fuels, vehicles, refrigerants, inputs, and methods.
- `migrations`: embedded SQL migrations used to create and evolve the SQLite schema.
- `factorpacks`: embedded normalized factor-pack JSON data.
- `docs`: project documentation; currently minimal and partially outdated.
- `data`: default local database location created when the app runs.
- `testdata`: currently empty placeholder for future test fixtures and supporting test data.

## Development Notes

- Run `gofmt` on edited Go files before committing changes.
- Run `go test ./...` after changes that affect parsing, storage, calculation, reporting, or backend helpers.
- Run `go test -tags ghgo_devtools ./internal/factors` after changing DEFRA import or mapping code.
- Add schema changes as new files under `migrations`; migrations are embedded through `migrations/embed.go`.
- Keep HTTP handlers transport-only: decode requests, fill path IDs, call `internal/app`, and encode JSON responses.
- Implement report generation/export work in or near `internal/report`.
- Implement graph data and export work near `internal/report`, reusing existing chart dataset tables where practical.
- Add or update default emission factors by editing checked-in factor-pack JSON under `factorpacks`; update `internal/factors`, `internal/vocab`, `internal/store`, and calculation code only when the normalized schema or lookup behavior needs to change.
- Add or update input formats in `internal/input` and expose them through backend services.
- Keep generated SQLite databases, local binaries, generated reports, generated graphs, and exported files out of version control.
