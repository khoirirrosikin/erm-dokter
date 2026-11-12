package resumepasien

import (
	"context"
	"database/sql"
)

type Repository interface {
	DetailResumePasienRalan(ctx context.Context, noRawat string) (*ResumePasienRalan, error)
	RiwayatResumePasienRalanByNoRM(ctx context.Context, noRM string) ([]ResumePasienRalan, error)
	CekResumePasienRalanAda(ctx context.Context, noRawat string) (bool, error)
	SimpanResumePasienRalan(ctx context.Context, noRawat, kodeDokter string, req SimpanResumePasienRalanRequest) (*ResumePasienRalan, error)
	UpdateResumePasienRalan(ctx context.Context, noRawat string, req UpdateResumePasienRalanRequest) (*ResumePasienRalan, error)
	HapusResumePasienRalan(ctx context.Context, noRawat string) error

	DetailResumePasienRanap(ctx context.Context, noRawat string) (*ResumePasienRanap, error)
	RiwayatResumePasienRanapByNoRM(ctx context.Context, noRM string) ([]ResumePasienRanap, error)
	GetInfoKamarInap(ctx context.Context, noRawat string) (*InfoKamarInap, error)
	CekResumePasienRanapAda(ctx context.Context, noRawat string) (bool, error)
	SimpanResumePasienRanap(ctx context.Context, noRawat, kodeDokter string, req SimpanResumePasienRanapRequest) (*ResumePasienRanap, error)
	UpdateResumePasienRanap(ctx context.Context, noRawat string, req UpdateResumePasienRanapRequest) (*ResumePasienRanap, error)
	HapusResumePasienRanap(ctx context.Context, noRawat string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}
