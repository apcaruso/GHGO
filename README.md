# ghgo

ghgo is a local-first desktop application for greenhouse gas (GHG) activity data entry, emissions calculation, and early-stage reporting.

The project is written in Go, uses SQLite for local storage, and uses Fyne for the desktop interface. It is not finished yet. The current implementation is functional enough to create local reporting data, run calculations, and inspect report tables, but several important product features and user-experience improvements are still planned.

## What The Program Does

ghgo helps a user collect activity data for an organization and convert that data into emissions results in `kgCO2e` and, where displayed by the UI, `tCO2e`.

Main features currently present in the codebase:

- Local SQLite database storage with embedded schema migrations.
- Desktop UI built with Fyne.
- Organization, facility, and reporting period setup.
- Reporting period settings, including mobile combustion method selection.
- Spreadsheet-style paste input flows with validation previews.
- Saved activity data views for the supported vectors.
- Editable/replaceable saved activity data by saving replacement input.
- Calculation runs that convert active activity records into emissions results.
- Report table builders and a report screen for completed calculation runs.
- Local seeding of a default DEFRA/DESNZ 2025 factor set at startup.

Supported activity data entry flows:

- Electricity: monthly facility consumption in `kWh`.
- Natural gas: monthly facility consumption in `Smc`.
- Mobile combustion by fuel consumed: fuel type and litres.
- Mobile combustion by distance travelled: vehicle details, fuel type, and `km`.
- Refrigerants: facility gas/substance and `kg` leaked.

Typical user workflow:

1. Start the application.
2. Create or select an organization.
3. Create facilities for the organization.
4. Create or select a reporting period.
5. Configure reporting period settings, especially the mobile combustion method if mobile data will be entered.
6. Paste activity data from a spreadsheet into the relevant screen.
7. Validate the pasted rows and fix any blocking errors.
8. Save the data. Saving replacement data supersedes matching active records.
9. Run a calculation from the Calculations screen.
10. Open the Reports screen to view report tables for a completed calculation run.

## Conversion Factors

The default factor set visible in the codebase is `DEFRA/DESNZ 2025`.

The application seeds default factors on startup from `internal/factors/seed_defra_2025.go`. The calculation engine then looks up factors stored in SQLite and applies them to active activity records to create calculation results.

The seeded factor coverage currently includes:

- UK purchased electricity, using `kgCO2e/kWh`.
- Natural gas, using `kgCO2e/Smc`.
- Mobile fuel combustion for diesel, petrol, LPG, and CNG, using `kgCO2e/L`.
- Vehicle distance factors for cars, vans, and motorbikes by supported size/fuel combinations, using `kgCO2e/km`.
- Refrigerants R134a, R410A, R407C, R404A, R32, and R22, using `kgCO2e/kg`.

The repository also contains `factorpacks/defra-2025/ghg-conversion-factors-2025-flat-format.xlsx` and DEFRA import/mapping code under `internal/factors`. The importer code is guarded by the `ghgo_devtools` build tag and there is no normal end-user import command in the repository yet.

## Current Limitations

This project is not finished yet.

Known limitations:

- The graphical interface is still basic and needs significant usability work.
- Report tables exist, but polished report generation, printing, and export workflows are not complete.
- Chart dataset tables exist, but graph rendering, printing, and exporting are not implemented.
- Supported activity categories and conversion factors are intentionally limited.
- The application is local-first only; there is no server, collaboration, or cloud sync flow.
- Data entry is based on pasted spreadsheet text rather than full file import workflows.
- The DEFRA factor import path exists in development code, but it is not exposed as a standard application command.
- The `docs/architecture.md` file is currently outdated compared with the implemented UI.

## Planned Work / TODO

- Add graph printing/exporting.
- Add report generation.
- Greatly improve the graphical user interface.
- Make the interface more usable and intuitive.
- Add support for more conversion factors.
- Improve the overall user experience.

## Usage Instructions

Prerequisites:

- Go compatible with the module declaration in `go.mod` (`go 1.25.0`).
- Network access the first time Go resolves module dependencies.
- Any OS-specific Fyne desktop prerequisites required by your platform.

Install dependencies:

```sh
go mod download
```

Run the desktop application:

```sh
go run ./cmd/ghgo
```

By default, ghgo stores its local SQLite database at:

```text
./data/ghgo.sqlite
```

Override the database path:

```sh
GHGO_DB_PATH=/custom/path/ghgo.sqlite go run ./cmd/ghgo
```

Run tests:

```sh
go test ./...
```

Build a local binary:

```sh
go build ./cmd/ghgo
```

## Repository Structure

- `cmd/ghgo`: application entry point; opens SQLite, runs migrations, seeds default factors, and starts the UI.
- `internal/ui`: Fyne desktop screens, navigation, widgets, forms, labels, and saved-data views.
- `internal/input`: parsers, validators, normalization, hashing, and commit logic for pasted activity data.
- `internal/calc`: calculation engine that converts active activity records into calculation results.
- `internal/report`: deterministic report table builders for completed calculation runs.
- `internal/factors`: DEFRA/DESNZ factor seeding, lookup, parsing, and development import/mapping code.
- `internal/store`: SQLite access layer, queries, migrations runner, and persistence methods.
- `internal/domain`: core domain types and enums.
- `internal/vocab`: normalized vocabulary for units, months, fuels, vehicles, refrigerants, inputs, and methods.
- `migrations`: embedded SQL migrations used to create and evolve the SQLite schema.
- `factorpacks`: source factor workbook files currently present in the repository.
- `docs`: project documentation; currently minimal and partially outdated.
- `data`: default local database location created when the app runs.
- `testdata`: currently empty placeholder for future test fixtures and supporting test data.

## Development Notes

- Run `gofmt` on edited Go files before committing changes.
- Run `go test ./...` after changes that affect parsing, storage, calculation, reporting, or UI helpers.
- Add schema changes as new files under `migrations`; migrations are embedded through `migrations/embed.go`.
- Implement UI and usability improvements in `internal/ui`.
- Implement report generation/export work in `internal/report` and expose it through `internal/ui/reports_screen.go`.
- Implement graph data, rendering, printing, and export work near `internal/report` and the reports UI, reusing existing chart dataset tables where practical.
- Add new conversion factor support in `internal/factors`, `internal/vocab`, `internal/store`, and the calculation code as needed.
- Add or update input formats in `internal/input` and connect them to screens in `internal/ui`.
- Keep generated SQLite databases, local binaries, generated reports, generated graphs, and exported files out of version control.
