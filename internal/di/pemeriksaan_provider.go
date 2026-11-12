package di

import (
	"database/sql"

	"erm-dokter/internal/config"
	"erm-dokter/internal/pemeriksaan"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
)

func providePemeriksaan(
	db *sql.DB,
	cfg *config.Config,
	log *logger.Logger,
	rawatJalanSvc rawatjalan.Service,
	rawatInapSvc rawatinap.Service,
) *pemeriksaan.Handler {
	repo := pemeriksaan.NewRepository(db)
	svc := pemeriksaan.NewService(repo, rawatJalanSvc, rawatInapSvc, cfg.MaxEditRekamMedisJam, log)
	return pemeriksaan.NewHandler(svc, cfg.EncryptionKey)
}
