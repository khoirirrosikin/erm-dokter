package penilaianmedis

import (
	"context"
	"database/sql"
	"strings"
)

type Repository interface {
	DetailPenilaianMedisRalan(ctx context.Context, noRawat string) (*PenilaianMedisRalan, error)
	RiwayatPenilaianMedisRalanByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRalan, error)
	CekPenilaianMedisRalanAda(ctx context.Context, noRawat string) (bool, error)
	SimpanPenilaianMedisRalan(ctx context.Context, noRawat, kodeDokter string, req SimpanPenilaianMedisRalanRequest) error
	UpdatePenilaianMedisRalan(ctx context.Context, noRawat string, req UpdatePenilaianMedisRalanRequest) error
	HapusPenilaianMedisRalan(ctx context.Context, noRawat string) error

	DetailPenilaianMedisIGD(ctx context.Context, noRawat string) (*PenilaianMedisIGD, error)
	RiwayatPenilaianMedisIGDByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisIGD, error)
	CekPenilaianMedisIGDAda(ctx context.Context, noRawat string) (bool, error)
	SimpanPenilaianMedisIGD(ctx context.Context, noRawat, kodeDokter string, req SimpanPenilaianMedisIGDRequest) error
	UpdatePenilaianMedisIGD(ctx context.Context, noRawat string, req UpdatePenilaianMedisIGDRequest) error
	HapusPenilaianMedisIGD(ctx context.Context, noRawat string) error

	DetailPenilaianMedisRanap(ctx context.Context, noRawat string) (*PenilaianMedisRanap, error)
	RiwayatPenilaianMedisRanapByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRanap, error)
	CekPenilaianMedisRanapAda(ctx context.Context, noRawat string) (bool, error)
	SimpanPenilaianMedisRanap(ctx context.Context, noRawat, kodeDokter string, req SimpanPenilaianMedisRanapRequest) error
	UpdatePenilaianMedisRanap(ctx context.Context, noRawat string, req UpdatePenilaianMedisRanapRequest) error
	HapusPenilaianMedisRanap(ctx context.Context, noRawat string) error

	DetailPenilaianMedisRalanKandungan(ctx context.Context, noRawat string) (*PenilaianMedisRalanKandungan, error)
	RiwayatPenilaianMedisRalanKandunganByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRalanKandungan, error)
	CekPenilaianMedisRalanKandunganAda(ctx context.Context, noRawat string) (bool, error)
	SimpanPenilaianMedisRalanKandungan(ctx context.Context, noRawat, kodeDokter string, req SimpanPenilaianMedisRalanKandunganRequest) error
	UpdatePenilaianMedisRalanKandungan(ctx context.Context, noRawat string, req UpdatePenilaianMedisRalanKandunganRequest) error
	HapusPenilaianMedisRalanKandungan(ctx context.Context, noRawat string) error

	DetailPenilaianMedisRanapKandungan(ctx context.Context, noRawat string) (*PenilaianMedisRanapKandungan, error)
	RiwayatPenilaianMedisRanapKandunganByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRanapKandungan, error)
	CekPenilaianMedisRanapKandunganAda(ctx context.Context, noRawat string) (bool, error)
	SimpanPenilaianMedisRanapKandungan(ctx context.Context, noRawat, kodeDokter string, req SimpanPenilaianMedisRanapKandunganRequest) error
	UpdatePenilaianMedisRanapKandungan(ctx context.Context, noRawat string, req UpdatePenilaianMedisRanapKandunganRequest) error
	HapusPenilaianMedisRanapKandungan(ctx context.Context, noRawat string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func normalizeZeroDateTime(val string) string {
	trimmed := strings.TrimSpace(val)
	if strings.HasPrefix(trimmed, "0000-00-00") {
		return ""
	}
	return trimmed
}
