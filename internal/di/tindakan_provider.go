package di

import (
	"erm-dokter/internal/config"
	"erm-dokter/internal/tindakan"
)

func provideTindakan(svc tindakan.Service, cfg *config.Config) *tindakan.Handler {
	return tindakan.NewHandler(svc, cfg.EncryptionKey)
}
