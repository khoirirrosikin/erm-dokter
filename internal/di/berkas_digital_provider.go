package di

import (
	"erm-dokter/internal/berkasdigital"
	"erm-dokter/internal/config"
)

func provideBerkasDigital(svc berkasdigital.Service, cfg *config.Config) *berkasdigital.Handler {
	return berkasdigital.NewHandler(svc, cfg.EncryptionKey)
}
