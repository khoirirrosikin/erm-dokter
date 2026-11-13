package di

import (
	"database/sql"

	"erm-dokter/internal/config"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/radiologi"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/tindakan"
)

func provideRadiologi(
	db *sql.DB,
	cfg *config.Config,
	log *logger.Logger,
	rawatJalanSvc rawatjalan.Service,
	rawatInapSvc rawatinap.Service,
	tindakanSvc tindakan.Service,
) *radiologi.Handler {
	repo := radiologi.NewRepository(db)

	svc := radiologi.NewService(
		repo,
		rawatJalanSvc,
		rawatInapSvc,
		tindakanSvc,
		cfg.MaxEditRekamMedisJam,
		log,
		cfg.KodePenjaminKunciBayar...,
	)
	return radiologi.NewHandler(svc, cfg.EncryptionKey)
}
