package di

import (
	"erm-dokter/internal/config"
	"erm-dokter/internal/rawatinap"
)

func provideRawatInapHandler(svc rawatinap.Service, cfg *config.Config) *rawatinap.Handler {
	return rawatinap.NewHandler(svc, cfg.EncryptionKey)
}
