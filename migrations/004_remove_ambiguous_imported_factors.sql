DELETE FROM emission_factors
WHERE factor_set_id = 'factor_set_defra_2025'
  AND activity_type = 'diesel_mobile'
  AND fuel_type = 'diesel'
  AND NOT (
    level_1 = 'Fuels'
    AND level_2 = 'Liquid fuels'
    AND level_3 = 'Diesel (average biofuel blend)'
  );

DELETE FROM emission_factors
WHERE factor_set_id = 'factor_set_defra_2025'
  AND activity_type = 'petrol_mobile'
  AND fuel_type = 'petrol'
  AND NOT (
    level_1 = 'Fuels'
    AND level_2 = 'Liquid fuels'
    AND level_3 = 'Petrol (average biofuel blend)'
  );

DELETE FROM emission_factors
WHERE factor_set_id = 'factor_set_defra_2025'
  AND activity_type = 'vehicle_distance'
  AND vehicle_type = 'car'
  AND lower(COALESCE(level_2, '')) LIKE '%market segment%';

DELETE FROM emission_factors
WHERE factor_set_id = 'factor_set_defra_2025'
  AND activity_type = 'refrigerant_leakage'
  AND lower(COALESCE(column_text, '')) LIKE 'emissions including only %';
