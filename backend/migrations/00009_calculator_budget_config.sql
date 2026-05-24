-- +goose Up
CREATE TABLE calculator_budget_config (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    config JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO calculator_budget_config (id, config) VALUES (1, '{
  "dev_rate_rub": 4000,
  "audit_base": 100000,
  "training": 40000,
  "om_base": 8000,
  "capex_min": 150000,
  "integration_mult_simple": 0.85,
  "integration_mult_standard": 1,
  "integration_mult_complex": 1.35,
  "integration_cost_2_roles": 40000,
  "integration_cost_3_roles": 80000,
  "integration_cost_4plus_roles": 150000,
  "om_units_500": 4000,
  "om_units_2000": 7000,
  "om_steps_3": 3000,
  "om_steps_5": 5000,
  "om_hm_150": 3000,
  "dev_hours_base": 20,
  "dev_hours_per_step": 6,
  "dev_hours_units_500": 15,
  "dev_hours_units_1000": 30,
  "dev_hours_units_2000": 40,
  "dev_hours_automation_high": 15,
  "dev_hours_hm_80": 10,
  "dev_hours_hm_200": 20,
  "automation_pct_threshold": 70,
  "capex_round_step": 50000,
  "om_round_step": 1000,
  "tier_simple_max": 3,
  "tier_standard_max": 6,
  "units_threshold_500": 500,
  "units_threshold_1000": 1000,
  "units_threshold_2000": 2000,
  "hm_threshold_50": 50,
  "hm_threshold_80": 80,
  "hm_threshold_150": 150,
  "hm_threshold_200": 200
}'::jsonb);

-- +goose Down
DROP TABLE IF EXISTS calculator_budget_config;
