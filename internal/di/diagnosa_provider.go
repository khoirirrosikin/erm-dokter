package di

import (
	"database/sql"
	"time"

	"erm-dokter/internal/config"
	"erm-dokter/internal/diagnosa"
	"erm-dokter/internal/master"
	"erm-dokter/internal/pkg/eklaim"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
)

func provideDiagnosa(
	db *sql.DB,
	cfg *config.Config,
	log *logger.Logger,
	rawatJalanSvc rawatjalan.Service,
	rawatInapSvc rawatinap.Service,
	masterSvc master.Service,
) *diagnosa.Handler {
	repo := diagnosa.NewRepository(db)
	eklaimClient := eklaim.NewClient(
		cfg.EKLAIMBaseURL,
		cfg.EKLAIMEncryptionKey,
		cfg.EKLAIMKodeRS,
		cfg.EKLAIMKodeTarif,
		cfg.EKLAIMDefaultCoderNIK,
		15*time.Second,
	)
	svc := diagnosa.NewService(repo, rawatJalanSvc, rawatInapSvc, masterSvc, eklaimClient, cfg.MaxEditRekamMedisJam, log)
	return diagnosa.NewHandler(svc, cfg.EncryptionKey)
}
