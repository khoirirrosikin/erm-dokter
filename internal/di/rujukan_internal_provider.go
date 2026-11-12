package di

import (
	"database/sql"

	"erm-dokter/internal/config"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/rujukaninternal"
)

func provideRujukanInternal(
	db *sql.DB,
	cfg *config.Config,
	log *logger.Logger,
	rawatJalanSvc rawatjalan.Service,
) *rujukaninternal.Handler {
	repo := rujukaninternal.NewRepository(db)
	svc := rujukaninternal.NewService(repo, rawatJalanSvc, cfg.MaxEditRekamMedisJam, log)
	return rujukaninternal.NewHandler(svc, cfg.EncryptionKey)
}
