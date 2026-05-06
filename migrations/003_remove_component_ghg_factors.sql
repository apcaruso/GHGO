DELETE FROM emission_factors
WHERE factor_set_id = 'factor_set_defra_2025'
  AND (
    lower(metadata_json) LIKE '%"ghg_unit":"kg co2e of %'
    OR lower(metadata_json) LIKE '%"ghg_unit": "kg co2e of %'
  );
