package di

import (
	"erm-dokter/internal/config"
	"erm-dokter/internal/obat"
)

func provideObat(svc obat.Service, cfg *config.Config) *obat.Handler {
	return obat.NewHandler(svc, cfg.EncryptionKey)
}
