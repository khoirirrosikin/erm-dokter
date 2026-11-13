package di

import (
	"database/sql"

	"erm-dokter/internal/berkasdigital"
	"erm-dokter/internal/config"
	"erm-dokter/internal/laboratorium"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/tindakan"
)

func provideLaboratorium(
	db *sql.DB,
	cfg *config.Config,
	log *logger.Logger,
	berkasSvc berkasdigital.Service,
	rawatJalanSvc rawatjalan.Service,
	rawatInapSvc rawatinap.Service,
	tindakanSvc tindakan.Service,
) *laboratorium.Handler {
	repo := laboratorium.NewRepository(db)

	svc := laboratorium.NewService(
		repo,
		berkasSvc,
		rawatJalanSvc,
		rawatInapSvc,
		tindakanSvc,
		cfg.MaxEditRekamMedisJam,
		cfg.URLBerkasDigital,
		cfg.KodeBerkasLabPK,
		cfg.KodeBerkasLabPA,
		cfg.KodeBerkasLabMB,
		log,
		cfg.KodePenjaminKunciBayar...,
	)
	return laboratorium.NewHandler(svc, cfg.EncryptionKey)
}
