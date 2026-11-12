package di_test

import (
	"database/sql"
	"testing"

	"erm-dokter/internal/config"
	"erm-dokter/internal/di"
	"erm-dokter/internal/pkg/logger"
)

func TestProvideRouteConfig_WiresAllHandlersCleanly(t *testing.T) {
	var db *sql.DB

	cfg := &config.Config{
		AppEnv:                      "test",
		JWTSecret:                   "0123456789abcdef0123456789abcdef",
		EncryptionKey:               "0123456789abcdef0123456789abcdef",
		ServiceAPIKey:               "test-service-key",
		MaxEditRekamMedisJam:        48,
		URLBerkasDigital:            "http://simrs.local/webapps/berkasrawat/",
		URLAntreanPoli:              "http://simrs.local/antrean",
		LoginRateLimitRate:          5,
		LoginRateLimitWindowMinutes: 1,
		LoginRateLimitEnabled:       true,
	}

	rc := di.ProvideRouteConfig(db, cfg, logger.New())
	if rc == nil {
		t.Fatal("expected non-nil RouteConfig")
	}
	if rc.MasterHandler == nil || rc.DiagnosaHandler == nil || rc.PenilaianMedisHandler == nil || rc.LaboratoriumHandler == nil {
		t.Fatalf("expected handlers to be wired, got Master=%v Diagnosa=%v PenilaianMedis=%v Lab=%v",
			rc.MasterHandler, rc.DiagnosaHandler, rc.PenilaianMedisHandler, rc.LaboratoriumHandler)
	}
}
