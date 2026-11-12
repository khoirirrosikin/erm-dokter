package di

import (
	"erm-dokter/internal/master"
)

func provideMaster(svc master.Service) *master.Handler {
	return master.NewHandler(svc)
}
