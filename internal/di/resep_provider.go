package di

import (
	"database/sql"

	"erm-dokter/internal/config"
	"erm-dokter/internal/obat"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/resep"
)

func provideResep(
	db *sql.DB,
	cfg *config.Config,
	log *logger.Logger,
	rawatJalanSvc rawatjalan.Service,
	rawatInapSvc rawatinap.Service,
	obatSvc obat.Service,
) *resep.Handler {
	repo := resep.NewRepository(db)
	svc := resep.NewService(repo, rawatJalanSvc, rawatInapSvc, obatSvc, cfg.MaxEditRekamMedisJam, log)
	return resep.NewHandler(svc, cfg.EncryptionKey)
}
