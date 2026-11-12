package di

import (
	"erm-dokter/internal/config"
	"erm-dokter/internal/rawatjalan"
)

func provideRawatJalan(svc rawatjalan.Service, cfg *config.Config) *rawatjalan.Handler {
	return rawatjalan.NewHandler(svc, cfg.EncryptionKey)
}
