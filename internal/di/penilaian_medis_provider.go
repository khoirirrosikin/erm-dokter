package di

import (
	"database/sql"

	"erm-dokter/internal/config"
	"erm-dokter/internal/penilaianmedis"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
)

func providePenilaianMedis(
	db *sql.DB,
	cfg *config.Config,
	log *logger.Logger,
	rawatJalanSvc rawatjalan.Service,
	rawatInapSvc rawatinap.Service,
) *penilaianmedis.Handler {
	repo := penilaianmedis.NewRepository(db)
	svc := penilaianmedis.NewService(repo, rawatJalanSvc, rawatInapSvc, cfg.MaxEditRekamMedisJam, log)
	return penilaianmedis.NewHandler(svc, cfg.EncryptionKey)
}
