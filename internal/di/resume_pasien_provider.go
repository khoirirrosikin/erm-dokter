package di

import (
	"database/sql"

	"erm-dokter/internal/config"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/resumepasien"
)

func provideResumePasien(
	db *sql.DB,
	cfg *config.Config,
	log *logger.Logger,
	rawatJalanSvc rawatjalan.Service,
	rawatInapSvc rawatinap.Service,
) *resumepasien.Handler {
	repo := resumepasien.NewRepository(db)
	svc := resumepasien.NewService(repo, rawatJalanSvc, rawatInapSvc, cfg.MaxEditRekamMedisJam, log)
	return resumepasien.NewHandler(svc, cfg.EncryptionKey)
}
