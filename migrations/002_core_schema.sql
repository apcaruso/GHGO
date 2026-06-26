CREATE TABLE IF NOT EXISTS organizations_next (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

INSERT OR IGNORE INTO organizations_next (id, name, created_at, updated_at)
SELECT id, name, created_at, created_at FROM organizations;

DROP TABLE organizations;

ALTER TABLE organizations_next RENAME TO organizations;

CREATE TABLE IF NOT EXISTS facilities (
  id TEXT PRIMARY KEY,
  organization_id TEXT NOT NULL,
  name TEXT NOT NULL,
  country_code TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_facilities_organization_id ON facilities (organization_id);
CREATE INDEX IF NOT EXISTS idx_facilities_organization_name ON facilities (organization_id, name);

CREATE TABLE IF NOT EXISTS reporting_periods (
  id TEXT PRIMARY KEY,
  organization_id TEXT NOT NULL,
  year INTEGER NOT NULL,
  starts_on TEXT NOT NULL,
  ends_on TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (organization_id, year),
  FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_reporting_periods_organization_id ON reporting_periods (organization_id);
CREATE INDEX IF NOT EXISTS idx_reporting_periods_organization_year ON reporting_periods (organization_id, year);

CREATE TABLE IF NOT EXISTS reporting_period_settings (
  id TEXT PRIMARY KEY,
  organization_id TEXT NOT NULL,
  reporting_period_id TEXT NOT NULL,
  mobile_method TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (reporting_period_id),
  FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT,
  FOREIGN KEY (reporting_period_id) REFERENCES reporting_periods(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS electricity_settings (
  id TEXT PRIMARY KEY,
  organization_id TEXT NOT NULL,
  reporting_period_id TEXT NOT NULL,
  facility_id TEXT NOT NULL,
  has_guarantees_of_origin INTEGER NOT NULL,
  go_coverage TEXT NOT NULL,
  go_reference TEXT NULL,
  go_market TEXT NULL,
  go_cancelled_at TEXT NULL,
  evidence_file_id TEXT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (reporting_period_id, facility_id),
  FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT,
  FOREIGN KEY (reporting_period_id) REFERENCES reporting_periods(id) ON DELETE RESTRICT,
  FOREIGN KEY (facility_id) REFERENCES facilities(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS activity_records (
  id TEXT PRIMARY KEY,
  organization_id TEXT NOT NULL,
  facility_id TEXT NULL,
  reporting_period_id TEXT NOT NULL,
  source_kind TEXT NOT NULL,
  scope INTEGER NOT NULL,
  vector TEXT NOT NULL,
  category TEXT NOT NULL,
  method TEXT NOT NULL,
  activity_type TEXT NOT NULL,
  period_start TEXT NOT NULL,
  period_end TEXT NOT NULL,
  amount REAL NOT NULL,
  unit TEXT NOT NULL,
  fuel_type TEXT NULL,
  vehicle_name TEXT NULL,
  plate TEXT NULL,
  vehicle_type TEXT NULL,
  vehicle_size_class TEXT NULL,
  substance TEXT NULL,
  status TEXT NOT NULL,
  source_hash TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT,
  FOREIGN KEY (facility_id) REFERENCES facilities(id) ON DELETE RESTRICT,
  FOREIGN KEY (reporting_period_id) REFERENCES reporting_periods(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_activity_records_organization_id ON activity_records (organization_id);
CREATE INDEX IF NOT EXISTS idx_activity_records_reporting_period_id ON activity_records (reporting_period_id);
CREATE INDEX IF NOT EXISTS idx_activity_records_period_vector ON activity_records (reporting_period_id, vector);
CREATE INDEX IF NOT EXISTS idx_activity_records_period_source_kind ON activity_records (reporting_period_id, source_kind);
CREATE INDEX IF NOT EXISTS idx_activity_records_period_status ON activity_records (reporting_period_id, status);
CREATE INDEX IF NOT EXISTS idx_activity_records_source_hash ON activity_records (source_hash);

CREATE TABLE IF NOT EXISTS paste_batches (
  id TEXT PRIMARY KEY,
  organization_id TEXT NOT NULL,
  reporting_period_id TEXT NOT NULL,
  input_kind TEXT NOT NULL,
  context_json TEXT NOT NULL,
  raw_text TEXT NOT NULL,
  raw_hash TEXT NOT NULL,
  rows_total INTEGER NOT NULL,
  rows_valid INTEGER NOT NULL,
  rows_error INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  committed_at TEXT NULL,
  FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT,
  FOREIGN KEY (reporting_period_id) REFERENCES reporting_periods(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_paste_batches_reporting_period_id ON paste_batches (reporting_period_id);
CREATE INDEX IF NOT EXISTS idx_paste_batches_raw_hash ON paste_batches (raw_hash);

CREATE TABLE IF NOT EXISTS paste_rows (
  id TEXT PRIMARY KEY,
  paste_batch_id TEXT NOT NULL,
  row_number INTEGER NOT NULL,
  raw_json TEXT NOT NULL,
  normalized_json TEXT NOT NULL,
  errors_json TEXT NOT NULL,
  warnings_json TEXT NOT NULL,
  activity_record_id TEXT NULL,
  FOREIGN KEY (paste_batch_id) REFERENCES paste_batches(id) ON DELETE CASCADE,
  FOREIGN KEY (activity_record_id) REFERENCES activity_records(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_paste_rows_paste_batch_id ON paste_rows (paste_batch_id);
CREATE INDEX IF NOT EXISTS idx_paste_rows_batch_row_number ON paste_rows (paste_batch_id, row_number);

CREATE TABLE IF NOT EXISTS factor_sets (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  source TEXT NOT NULL,
  year INTEGER NOT NULL,
  version TEXT NOT NULL,
  imported_at TEXT NOT NULL,
  metadata_json TEXT NOT NULL,
  UNIQUE (source, year, version)
);

CREATE TABLE IF NOT EXISTS emission_factors (
  id TEXT PRIMARY KEY,
  factor_set_id TEXT NOT NULL,
  source TEXT NOT NULL,
  scope INTEGER NOT NULL,
  level_1 TEXT NULL,
  level_2 TEXT NULL,
  level_3 TEXT NULL,
  level_4 TEXT NULL,
  column_text TEXT NULL,
  activity_type TEXT NULL,
  fuel_type TEXT NULL,
  vehicle_type TEXT NULL,
  vehicle_size_class TEXT NULL,
  substance TEXT NULL,
  input_unit TEXT NOT NULL,
  factor_unit TEXT NOT NULL,
  ghg TEXT NOT NULL,
  factor_value REAL NOT NULL,
  metadata_json TEXT NOT NULL,
  FOREIGN KEY (factor_set_id) REFERENCES factor_sets(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_emission_factors_factor_set_id ON emission_factors (factor_set_id);
CREATE INDEX IF NOT EXISTS idx_emission_factors_set_activity_type ON emission_factors (factor_set_id, activity_type);
CREATE INDEX IF NOT EXISTS idx_emission_factors_set_fuel_type ON emission_factors (factor_set_id, fuel_type);
CREATE INDEX IF NOT EXISTS idx_emission_factors_set_vehicle ON emission_factors (factor_set_id, vehicle_type, vehicle_size_class, fuel_type);
CREATE INDEX IF NOT EXISTS idx_emission_factors_set_substance ON emission_factors (factor_set_id, substance);
CREATE INDEX IF NOT EXISTS idx_emission_factors_set_input_unit ON emission_factors (factor_set_id, input_unit);

CREATE TABLE IF NOT EXISTS calculation_runs (
  id TEXT PRIMARY KEY,
  organization_id TEXT NOT NULL,
  reporting_period_id TEXT NOT NULL,
  factor_set_id TEXT NOT NULL,
  started_at TEXT NOT NULL,
  completed_at TEXT NULL,
  settings_snapshot_json TEXT NOT NULL,
  FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT,
  FOREIGN KEY (reporting_period_id) REFERENCES reporting_periods(id) ON DELETE RESTRICT,
  FOREIGN KEY (factor_set_id) REFERENCES factor_sets(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_calculation_runs_reporting_period_id ON calculation_runs (reporting_period_id);
CREATE INDEX IF NOT EXISTS idx_calculation_runs_factor_set_id ON calculation_runs (factor_set_id);
CREATE TABLE IF NOT EXISTS calculation_results (
  id TEXT PRIMARY KEY,
  calculation_run_id TEXT NOT NULL,
  activity_record_id TEXT NOT NULL,
  scope INTEGER NOT NULL,
  vector TEXT NOT NULL,
  method TEXT NOT NULL,
  activity_amount REAL NOT NULL,
  activity_unit TEXT NOT NULL,
  factor_id TEXT NULL,
  factor_value REAL NOT NULL,
  factor_unit TEXT NOT NULL,
  factor_source TEXT NOT NULL,
  emissions_kgco2e REAL NOT NULL,
  is_primary INTEGER NOT NULL,
  notes_json TEXT NOT NULL,
  FOREIGN KEY (calculation_run_id) REFERENCES calculation_runs(id) ON DELETE RESTRICT,
  FOREIGN KEY (activity_record_id) REFERENCES activity_records(id) ON DELETE RESTRICT,
  FOREIGN KEY (factor_id) REFERENCES emission_factors(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_calculation_results_run_id ON calculation_results (calculation_run_id);
CREATE INDEX IF NOT EXISTS idx_calculation_results_run_vector ON calculation_results (calculation_run_id, vector);
CREATE INDEX IF NOT EXISTS idx_calculation_results_run_scope ON calculation_results (calculation_run_id, scope);
CREATE INDEX IF NOT EXISTS idx_calculation_results_run_method ON calculation_results (calculation_run_id, method);
CREATE INDEX IF NOT EXISTS idx_calculation_results_run_is_primary ON calculation_results (calculation_run_id, is_primary);

