package config_test

import (
	"os"
	"testing"

	"erm-dokter/internal/config"
)

func TestConfigLoad(t *testing.T) {
	os.Setenv("APP_ENV", "test")
	os.Setenv("APP_PORT", "9999")
	os.Setenv("MAX_EDIT_REKAM_MEDIS_JAM", "72")
	os.Setenv("KODE_BERKAS_LAB_PK", "005, 006")

	defer func() {
		os.Unsetenv("APP_ENV")
		os.Unsetenv("APP_PORT")
		os.Unsetenv("MAX_EDIT_REKAM_MEDIS_JAM")
		os.Unsetenv("KODE_BERKAS_LAB_PK")
	}()

	cfg := config.Load()

	if cfg.AppEnv != "test" {
		t.Errorf("Expected AppEnv 'test', got '%s'", cfg.AppEnv)
	}
	if cfg.AppPort != "9999" {
		t.Errorf("Expected AppPort '9999', got '%s'", cfg.AppPort)
	}
	if cfg.MaxEditRekamMedisJam != 72 {
		t.Errorf("Expected MaxEditRekamMedisJam 72, got %d", cfg.MaxEditRekamMedisJam)
	}
	if len(cfg.KodeBerkasLabPK) != 2 || cfg.KodeBerkasLabPK[0] != "005" || cfg.KodeBerkasLabPK[1] != "006" {
		t.Errorf("Unexpected KodeBerkasLabPK: %v", cfg.KodeBerkasLabPK)
	}
	if cfg.LogFormat != "json" {
		t.Errorf("Expected default LogFormat 'json', got '%s'", cfg.LogFormat)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("Expected default LogLevel 'info', got '%s'", cfg.LogLevel)
	}
	if cfg.LogFilePath != "logs/app.log" {
		t.Errorf("Expected default LogFilePath 'logs/app.log', got '%s'", cfg.LogFilePath)
	}
	if !cfg.LoginRateLimitEnabled {
		t.Errorf("Expected default LoginRateLimitEnabled true, got %v", cfg.LoginRateLimitEnabled)
	}
	if cfg.LoginRateLimitRate != 10 {
		t.Errorf("Expected default LoginRateLimitRate 10, got %d", cfg.LoginRateLimitRate)
	}
	if cfg.LoginRateLimitWindowMinutes != 1 {
		t.Errorf("Expected default LoginRateLimitWindowMinutes 1, got %d", cfg.LoginRateLimitWindowMinutes)
	}
}

func TestConfigLoad_CustomRateLimit(t *testing.T) {
	os.Setenv("LOGIN_RATE_LIMIT_ENABLED", "false")
	os.Setenv("LOGIN_RATE_LIMIT_RATE", "5")
	os.Setenv("LOGIN_RATE_LIMIT_WINDOW_MINUTES", "2")

	defer func() {
		os.Unsetenv("LOGIN_RATE_LIMIT_ENABLED")
		os.Unsetenv("LOGIN_RATE_LIMIT_RATE")
		os.Unsetenv("LOGIN_RATE_LIMIT_WINDOW_MINUTES")
	}()

	cfg := config.Load()

	if cfg.LoginRateLimitEnabled {
		t.Errorf("Expected LoginRateLimitEnabled false, got %v", cfg.LoginRateLimitEnabled)
	}
	if cfg.LoginRateLimitRate != 5 {
		t.Errorf("Expected LoginRateLimitRate 5, got %d", cfg.LoginRateLimitRate)
	}
	if cfg.LoginRateLimitWindowMinutes != 2 {
		t.Errorf("Expected LoginRateLimitWindowMinutes 2, got %d", cfg.LoginRateLimitWindowMinutes)
	}
}

func TestConfigLoad_EKLAIM(t *testing.T) {
	os.Setenv("EKLAIM_BASE_URL", "http://eklaim.example.com/e-klaim/ws.php")
	os.Setenv("EKLAIM_ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	os.Setenv("EKLAIM_KODE_RS", "1234567")
	os.Setenv("EKLAIM_KODE_TARIF", "DP")
	os.Setenv("EKLAIM_NIK_CODER", "9876543210123456")

	defer func() {
		os.Unsetenv("EKLAIM_BASE_URL")
		os.Unsetenv("EKLAIM_ENCRYPTION_KEY")
		os.Unsetenv("EKLAIM_KODE_RS")
		os.Unsetenv("EKLAIM_KODE_TARIF")
		os.Unsetenv("EKLAIM_NIK_CODER")
	}()

	cfg := config.Load()

	if cfg.EKLAIMBaseURL != "http://eklaim.example.com/e-klaim/ws.php" {
		t.Errorf("Expected EKLAIMBaseURL match, got '%s'", cfg.EKLAIMBaseURL)
	}
	if cfg.EKLAIMEncryptionKey != "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" {
		t.Errorf("Expected EKLAIMEncryptionKey match, got '%s'", cfg.EKLAIMEncryptionKey)
	}
	if cfg.EKLAIMKodeRS != "1234567" {
		t.Errorf("Expected EKLAIMKodeRS match, got '%s'", cfg.EKLAIMKodeRS)
	}
	if cfg.EKLAIMKodeTarif != "DP" {
		t.Errorf("Expected EKLAIMKodeTarif match, got '%s'", cfg.EKLAIMKodeTarif)
	}
	if cfg.EKLAIMDefaultCoderNIK != "9876543210123456" {
		t.Errorf("Expected EKLAIMDefaultCoderNIK match, got '%s'", cfg.EKLAIMDefaultCoderNIK)
	}
}

func TestConfigLoad_URLAntreanPoli(t *testing.T) {
	testURL := "http://antrean.example.com/antrianpoli/api.php?action=tambah"
	os.Setenv("URL_ANTREAN_POLI", testURL)
	defer os.Unsetenv("URL_ANTREAN_POLI")

	cfg := config.Load()
	if cfg.URLAntreanPoli != testURL {
		t.Errorf("Expected URLAntreanPoli '%s', got '%s'", testURL, cfg.URLAntreanPoli)
	}
}

func TestConfigLoad_AllowedIPs_NoFallback(t *testing.T) {
	// Development with no ALLOWED_IPS: must be empty slice (no fallback)
	os.Setenv("APP_ENV", "development")
	os.Unsetenv("ALLOWED_IPS")

	cfg := config.Load()
	if cfg.IsProduction() {
		t.Errorf("Expected IsProduction false for development")
	}
	if len(cfg.AllowedIPs) != 0 {
		t.Errorf("Expected AllowedIPs to be empty when not set, got %v", cfg.AllowedIPs)
	}

	// Production with custom ALLOWED_IPS
	os.Setenv("APP_ENV", "production")
	os.Setenv("ALLOWED_IPS", "10.0.0.1, 10.0.0.2")
	defer func() {
		os.Unsetenv("APP_ENV")
		os.Unsetenv("ALLOWED_IPS")
	}()

	cfgProd := config.Load()
	if !cfgProd.IsProduction() {
		t.Errorf("Expected IsProduction true for production")
	}
	if len(cfgProd.AllowedIPs) != 2 || cfgProd.AllowedIPs[0] != "10.0.0.1" || cfgProd.AllowedIPs[1] != "10.0.0.2" {
		t.Errorf("Expected custom AllowedIPs, got %v", cfgProd.AllowedIPs)
	}
}

func TestConfig_Validate_Production_AllowedIPsRequired(t *testing.T) {
	cfg := &config.Config{
		AppEnv:             "production",
		DBHost:             "localhost",
		DBUser:             "root",
		DBName:             "test_db",
		JWTSecret:          "secret",
		EncryptionKey:      "01234567890123456789012345678901",
		URLBerkasDigital:   "http://localhost/berkas",
		AllowedIPs:         nil, // Empty in production
	}
	os.Setenv("KODE_BERKAS_LAB_PA", "001")
	defer os.Unsetenv("KODE_BERKAS_LAB_PA")

	if err := cfg.Validate(); err == nil {
		t.Errorf("Expected validation error in production when AllowedIPs is empty")
	}

	// Now provide AllowedIPs
	cfg.AllowedIPs = []string{"10.0.0.1"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Expected validation to pass when AllowedIPs is provided, got: %v", err)
	}
}

