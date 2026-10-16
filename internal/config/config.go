package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                      string
	AppPort                     string
	AppName                     string
	DBHost                      string
	DBPort                      string
	DBUser                      string
	DBPass                      string
	DBName                      string
	JWTSecret                   string
	UserKey                     string
	PasswordKey                 string
	EncryptionKey               string
	CORSOrigin                  string
	MaxEditRekamMedisJam        int
	URLBerkasDigital            string
	ServiceAPIKey               string
	KodeBerkasLabPK             []string
	KodeBerkasLabPA             []string
	KodeBerkasLabMB             []string
	LogFormat                   string
	LogLevel                    string
	LogFilePath                 string
	LoginRateLimitEnabled       bool
	LoginRateLimitRate          int
	LoginRateLimitWindowMinutes int
	EKLAIMBaseURL               string
	EKLAIMEncryptionKey         string
	EKLAIMKodeRS                string
	EKLAIMKodeTarif             string
	EKLAIMDefaultCoderNIK       string
	URLAntreanPoli              string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO]\tFile .env tidak ditemukan, membaca dari environment variable sistem")
	}

	maxEditJam, _ := strconv.Atoi(getEnvOrDefault("MAX_EDIT_REKAM_MEDIS_JAM", "48"))
	if maxEditJam <= 0 {
		maxEditJam = 48
	}

	loginRateLimitEnabled := true
	if envVal := os.Getenv("LOGIN_RATE_LIMIT_ENABLED"); envVal != "" {
		if parsed, err := strconv.ParseBool(envVal); err == nil {
			loginRateLimitEnabled = parsed
		}
	}

	loginRateLimitRate, _ := strconv.Atoi(getEnvOrDefault("LOGIN_RATE_LIMIT_RATE", "10"))
	if loginRateLimitRate <= 0 {
		loginRateLimitRate = 10
	}

	loginRateLimitWindow, _ := strconv.Atoi(getEnvOrDefault("LOGIN_RATE_LIMIT_WINDOW_MINUTES", "1"))
	if loginRateLimitWindow <= 0 {
		loginRateLimitWindow = 1
	}

	parseKodeSlice := func(envKey string) []string {
		raw := strings.TrimSpace(os.Getenv(envKey))
		var result []string
		if raw != "" {
			for _, k := range strings.Split(raw, ",") {
				if kTrim := strings.TrimSpace(k); kTrim != "" {
					result = append(result, kTrim)
				}
			}
		}
		return result
	}

	cfg := &Config{
		AppEnv:                      getEnvOrDefault("APP_ENV", "development"),
		AppPort:                     getEnvOrDefault("APP_PORT", "8080"),
		AppName:                     getEnvOrDefault("APP_NAME", "erm-dokter"),
		DBHost:                      os.Getenv("DB_HOST"),
		DBPort:                      getEnvOrDefault("DB_PORT", "3306"),
		DBUser:                      os.Getenv("DB_USER"),
		DBPass:                      os.Getenv("DB_PASSWORD"),
		DBName:                      os.Getenv("DB_NAME"),
		JWTSecret:                   os.Getenv("JWT_SECRET"),
		UserKey:                     os.Getenv("USER_KEY"),
		PasswordKey:                 os.Getenv("PASSWORD_KEY"),
		EncryptionKey:               os.Getenv("ENCRYPTION_KEY"),
		CORSOrigin:                  getEnvOrDefault("CORS_ORIGIN", "*"),
		MaxEditRekamMedisJam:        maxEditJam,
		URLBerkasDigital:            os.Getenv("URL_BERKAS_DIGITAL"),
		ServiceAPIKey:               strings.TrimSpace(os.Getenv("SERVICE_API_KEY")),
		KodeBerkasLabPK:             parseKodeSlice("KODE_BERKAS_LAB_PK"),
		KodeBerkasLabPA:             parseKodeSlice("KODE_BERKAS_LAB_PA"),
		KodeBerkasLabMB:             parseKodeSlice("KODE_BERKAS_LAB_MB"),
		LogFormat:                   getEnvOrDefault("LOG_FORMAT", "json"),
		LogLevel:                    getEnvOrDefault("LOG_LEVEL", "info"),
		LogFilePath:                 getEnvOrDefault("LOG_FILE_PATH", "logs/app.log"),
		LoginRateLimitEnabled:       loginRateLimitEnabled,
		LoginRateLimitRate:          loginRateLimitRate,
		LoginRateLimitWindowMinutes: loginRateLimitWindow,
		EKLAIMBaseURL:               strings.TrimSpace(os.Getenv("EKLAIM_BASE_URL")),
		EKLAIMEncryptionKey:         strings.TrimSpace(os.Getenv("EKLAIM_ENCRYPTION_KEY")),
		EKLAIMKodeRS:                strings.TrimSpace(os.Getenv("EKLAIM_KODE_RS")),
		EKLAIMKodeTarif:             getEnvOrDefault("EKLAIM_KODE_TARIF", "BP"),
		EKLAIMDefaultCoderNIK:       getEnvOrDefault("EKLAIM_NIK_CODER", "1234567890123456"),
		URLAntreanPoli:              strings.TrimSpace(os.Getenv("URL_ANTREAN_POLI")),
	}
	return cfg
}

func (c *Config) Validate() error {
	required := map[string]string{
		"DB_HOST":            c.DBHost,
		"DB_USER":            c.DBUser,
		"DB_NAME":            c.DBName,
		"JWT_SECRET":         c.JWTSecret,
		"ENCRYPTION_KEY":     c.EncryptionKey,
		"URL_BERKAS_DIGITAL": c.URLBerkasDigital,
		"KODE_BERKAS_LAB_PA": os.Getenv("KODE_BERKAS_LAB_PA"),
	}
	for key, val := range required {
		if val == "" {
			return fmt.Errorf("environment variable %s wajib diisi", key)
		}
	}
	return nil
}

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
