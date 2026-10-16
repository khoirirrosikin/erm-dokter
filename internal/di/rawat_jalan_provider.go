package di

import (
	"database/sql"

	"erm-dokter/internal/config"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatjalan"
)

func provideRawatJalan(db *sql.DB, cfg *config.Config, log *logger.Logger) *rawatjalan.Handler {
	repo := rawatjalan.NewRepository(db)
	svc := rawatjalan.NewService(repo, log, cfg.URLAntreanPoli)
	return rawatjalan.NewHandler(svc, cfg.EncryptionKey)
}
