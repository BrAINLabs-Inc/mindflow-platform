package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoad_ReadsDotEnvWithoutOverriding covers dotenv.config() parity:
// values come from ./.env, but a variable already set in the real
// environment wins.
func TestLoad_ReadsDotEnvWithoutOverriding(t *testing.T) {
	dir := t.TempDir()
	env := "SUPABASE_URL=https://from-dotenv.supabase.co\nDATABASE_URL=postgres://from-dotenv\nPORT=4000\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(env), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	for _, k := range []string{"SUPABASE_URL", "DATABASE_URL", "R2_PUBLIC_URL", "PUBLIC_URL_BASE"} {
		t.Setenv(k, "") // registers cleanup of anything .env sets
		os.Unsetenv(k)
	}
	t.Setenv("PORT", "5000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.SupabaseURL != "https://from-dotenv.supabase.co" || cfg.DatabaseURL != "postgres://from-dotenv" {
		t.Errorf(".env values not loaded: %+v", cfg)
	}
	if cfg.Port != "5000" {
		t.Errorf("Port = %q, want real env 5000 to win over .env", cfg.Port)
	}
}

func TestLoad_MissingDotEnvIsFine(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("SUPABASE_URL", "https://x.supabase.co")
	t.Setenv("DATABASE_URL", "postgres://x")
	if _, err := Load(); err != nil {
		t.Fatalf("Load without .env: %v", err)
	}
}

func TestLoad_PublicURLBaseFallback(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("SUPABASE_URL", "https://x.supabase.co")
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("R2_PUBLIC_URL", "")
	t.Setenv("PUBLIC_URL_BASE", "https://legacy.example.com")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.R2PublicURLBase != "https://legacy.example.com" {
		t.Errorf("R2PublicURLBase = %q", cfg.R2PublicURLBase)
	}
}
