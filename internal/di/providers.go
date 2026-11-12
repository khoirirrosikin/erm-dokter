package di

import (
	"database/sql"
	"time"

	"erm-dokter/internal/berkasdigital"
	"erm-dokter/internal/config"
	"erm-dokter/internal/master"
	"erm-dokter/internal/middleware"
	"erm-dokter/internal/obat"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/routes"
	"erm-dokter/internal/shared/apperror"
	"erm-dokter/internal/tindakan"
)

func ProvideRouteConfig(db *sql.DB, cfg *config.Config, log *logger.Logger) *routes.RouteConfig {
	apperror.SetLogger(log)

	masterSvc := master.NewService(master.NewRepository(db), log)
	obatSvc := obat.NewService(obat.NewRepository(db), log)
	rawatJalanSvc := rawatjalan.NewService(rawatjalan.NewRepository(db), log, cfg.URLAntreanPoli)
	rawatInapSvc := rawatinap.NewService(rawatinap.NewRepository(db), log)
	tindakanSvc := tindakan.NewService(tindakan.NewRepository(db), log)
	berkasSvc := berkasdigital.NewService(berkasdigital.NewRepository(db), cfg.URLBerkasDigital, log)

	rc := routes.NewRouteConfig(
		provideHealth(),
		provideDocs(),
		provideAuth(db, cfg, log),
		provideMaster(masterSvc),
		provideObat(obatSvc, cfg),
		provideRawatJalan(rawatJalanSvc, cfg),
		provideRawatInapHandler(rawatInapSvc, cfg),
		providePemeriksaan(db, cfg, log, rawatJalanSvc, rawatInapSvc),
		provideResep(db, cfg, log, rawatJalanSvc, rawatInapSvc, obatSvc),
		provideRujukanInternal(db, cfg, log, rawatJalanSvc),
		providePenilaianMedis(db, cfg, log, rawatJalanSvc, rawatInapSvc),
		provideTindakan(tindakanSvc, cfg),
		provideLaboratorium(db, cfg, log, berkasSvc, rawatJalanSvc, rawatInapSvc, tindakanSvc),
		provideRadiologi(db, cfg, log, rawatJalanSvc, rawatInapSvc, tindakanSvc),
		provideBerkasDigital(berkasSvc, cfg),
		provideResumePasien(db, cfg, log, rawatJalanSvc, rawatInapSvc),
		providePasien(db, cfg, log),
		provideDiagnosa(db, cfg, log, rawatJalanSvc, rawatInapSvc, masterSvc),
		cfg.JWTSecret,
		cfg.ServiceAPIKey,
		cfg.AppEnv,
		cfg.AllowedIPs,
	)

	rc.LoginRateLimitMiddleware = middleware.LoginRateLimitMiddleware(
		cfg.LoginRateLimitRate,
		time.Duration(cfg.LoginRateLimitWindowMinutes)*time.Minute,
		cfg.LoginRateLimitEnabled,
	)

	return rc
}
