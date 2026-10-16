package routes

import (
	"net/http"
	"strings"
	"time"

	"erm-dokter/internal/auth"
	"erm-dokter/internal/berkasdigital"
	"erm-dokter/internal/diagnosa"
	"erm-dokter/internal/docs"
	"erm-dokter/internal/health"
	"erm-dokter/internal/laboratorium"
	"erm-dokter/internal/master"
	"erm-dokter/internal/middleware"
	"erm-dokter/internal/obat"
	"erm-dokter/internal/pasien"
	"erm-dokter/internal/pemeriksaan"
	"erm-dokter/internal/penilaianmedis"
	"erm-dokter/internal/radiologi"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/resep"
	"erm-dokter/internal/resumepasien"
	"erm-dokter/internal/rujukaninternal"
	"erm-dokter/internal/tindakan"
)

type RouteConfig struct {
	Mux                      *http.ServeMux
	HealthHandler            *health.Handler
	DocsHandler              *docs.Handler
	AuthHandler              *auth.Handler
	MasterHandler            *master.Handler
	ObatHandler              *obat.Handler
	RawatJalanHandler        *rawatjalan.Handler
	RawatInapHandler         *rawatinap.Handler
	PemeriksaanHandler       *pemeriksaan.Handler
	ResepHandler             *resep.Handler
	RujukanInternalHandler   *rujukaninternal.Handler
	PenilaianMedisHandler    *penilaianmedis.Handler
	TindakanHandler          *tindakan.Handler
	LaboratoriumHandler      *laboratorium.Handler
	RadiologiHandler         *radiologi.Handler
	BerkasDigitalHandler     *berkasdigital.Handler
	ResumePasienHandler      *resumepasien.Handler
	PasienHandler            *pasien.Handler
	DiagnosaHandler          *diagnosa.Handler
	AuthMiddleware           func(http.HandlerFunc) http.HandlerFunc
	ServiceAuthMiddleware    func(http.HandlerFunc) http.HandlerFunc
	TimeoutMiddleware        func(http.HandlerFunc) http.HandlerFunc
	LoginRateLimitMiddleware func(http.HandlerFunc) http.HandlerFunc
	AppEnv                   string
	AllowedIPs              []string
}

func (c *RouteConfig) IsProduction() bool {
	return strings.ToLower(strings.TrimSpace(c.AppEnv)) == "production"
}

func NewRouteConfig(
	healthHandler *health.Handler,
	docsHandler *docs.Handler,
	authHandler *auth.Handler,
	masterHandler *master.Handler,
	obatHandler *obat.Handler,
	rawatJalanHandler *rawatjalan.Handler,
	rawatInapHandler *rawatinap.Handler,
	pemeriksaanHandler *pemeriksaan.Handler,
	resepHandler *resep.Handler,
	rujukanInternalHandler *rujukaninternal.Handler,
	penilaianMedisHandler *penilaianmedis.Handler,
	tindakanHandler *tindakan.Handler,
	laboratoriumHandler *laboratorium.Handler,
	radiologiHandler *radiologi.Handler,
	berkasDigitalHandler *berkasdigital.Handler,
	resumePasienHandler *resumepasien.Handler,
	pasienHandler *pasien.Handler,
	diagnosaHandler *diagnosa.Handler,
	jwtSecret string,
	serviceAPIKey string,
	appEnv string,
	allowedIPs []string,
) *RouteConfig {
	return &RouteConfig{
		Mux:                      http.NewServeMux(),
		HealthHandler:            healthHandler,
		DocsHandler:              docsHandler,
		AuthHandler:              authHandler,
		MasterHandler:            masterHandler,
		ObatHandler:              obatHandler,
		RawatJalanHandler:        rawatJalanHandler,
		RawatInapHandler:         rawatInapHandler,
		PemeriksaanHandler:       pemeriksaanHandler,
		ResepHandler:             resepHandler,
		RujukanInternalHandler:   rujukanInternalHandler,
		PenilaianMedisHandler:    penilaianMedisHandler,
		TindakanHandler:          tindakanHandler,
		LaboratoriumHandler:      laboratoriumHandler,
		RadiologiHandler:         radiologiHandler,
		BerkasDigitalHandler:     berkasDigitalHandler,
		ResumePasienHandler:      resumePasienHandler,
		PasienHandler:            pasienHandler,
		DiagnosaHandler:          diagnosaHandler,
		AuthMiddleware:           middleware.JWTMiddleware(jwtSecret),
		ServiceAuthMiddleware:    middleware.ServiceOrJWTMiddleware(jwtSecret, serviceAPIKey),
		TimeoutMiddleware:        middleware.TimeoutMiddleware(30 * time.Second),
		LoginRateLimitMiddleware: middleware.LoginRateLimitMiddleware(10, 1*time.Minute, true),
		AppEnv:                   appEnv,
		AllowedIPs:              allowedIPs,
	}
}

func (c *RouteConfig) Setup() {
	loginRateLimit := c.LoginRateLimitMiddleware
	if loginRateLimit == nil {
		loginRateLimit = middleware.LoginRateLimitMiddleware(10, 1*time.Minute, true)
	}

	c.HealthHandler.RegisterRoutes(c.Mux)
	if !c.IsProduction() {
		c.DocsHandler.RegisterRoutes(c.Mux)
	}
	c.AuthHandler.RegisterRoutes(c.Mux, loginRateLimit, c.AuthMiddleware, c.TimeoutMiddleware)
	c.MasterHandler.RegisterRoutes(c.Mux, c.AuthMiddleware, c.TimeoutMiddleware)
	c.ObatHandler.RegisterRoutes(c.Mux, c.AuthMiddleware, c.TimeoutMiddleware)
	c.RawatJalanHandler.RegisterRoutes(c.Mux, c.AuthMiddleware, c.ServiceAuthMiddleware, c.TimeoutMiddleware)
	c.RawatInapHandler.RegisterRoutes(c.Mux, c.AuthMiddleware, c.ServiceAuthMiddleware, c.TimeoutMiddleware)
	c.PemeriksaanHandler.RegisterRoutes(c.Mux, c.AuthMiddleware, c.TimeoutMiddleware)
	c.ResepHandler.RegisterRoutes(c.Mux, c.AuthMiddleware, c.TimeoutMiddleware)
	c.RujukanInternalHandler.RegisterRoutes(c.Mux, c.AuthMiddleware, c.TimeoutMiddleware)
	c.PenilaianMedisHandler.RegisterRoutes(c.Mux, c.AuthMiddleware, c.TimeoutMiddleware)
	c.TindakanHandler.RegisterRoutes(c.Mux, c.AuthMiddleware, c.TimeoutMiddleware)
	c.LaboratoriumHandler.RegisterRoutes(c.Mux, c.AuthMiddleware, c.TimeoutMiddleware)
	c.RadiologiHandler.RegisterRoutes(c.Mux, c.AuthMiddleware, c.TimeoutMiddleware)
	c.BerkasDigitalHandler.RegisterRoutes(c.Mux, c.AuthMiddleware, c.TimeoutMiddleware)
	c.ResumePasienHandler.RegisterRoutes(c.Mux, c.AuthMiddleware, c.TimeoutMiddleware)
	c.PasienHandler.RegisterRoutes(c.Mux, c.AuthMiddleware, c.TimeoutMiddleware)
	c.DiagnosaHandler.RegisterRoutes(c.Mux, c.AuthMiddleware, c.TimeoutMiddleware)
}

func (c *RouteConfig) BuildHandler(corsOrigin string) http.Handler {
	var h http.Handler = c.Mux
	if c.IsProduction() {
		h = middleware.IPWhitelistMiddleware(c.AllowedIPs)(h)
	}
	h = middleware.LoggingMiddleware(h)
	h = middleware.RequestIDMiddleware(h)
	h = middleware.CORSMiddleware(corsOrigin)(h)
	return h
}
