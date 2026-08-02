package config

import "testing"

func TestParseYAMLAppliesStructuredFields(t *testing.T) {
	cfg, err := ParseYAML(`name: api
domain: api.example.com
port: 3000
replicas: 3
env:
  APP_ENV: production
  LOG_LEVEL: info
rollback:
  enabled: true
  failureText: [panic:, fatal]
`)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}

	if cfg.Name != "api" || cfg.Namespace != "api" || cfg.Port != 3000 || cfg.Replicas != 3 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.Env["APP_ENV"] != "production" || cfg.Rollback.FailureText[1] != "fatal" {
		t.Fatalf("nested values were not parsed: %+v", cfg)
	}
}

func TestValidateRequiresDomain(t *testing.T) {
	cfg := Config{Name: "api"}
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected missing domain validation error")
	}
}
