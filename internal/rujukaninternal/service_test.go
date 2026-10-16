package rujukaninternal

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

type mockRepository struct {
	opsiList         []OpsiPoliDokter
	opsiErr          error
	rujukanList      []RujukanInternal
	rujukanErr       error
	adaResult        bool
	adaErr           error
	simpanErr        error
	hapusErr         error
	simpanCalledWith struct{ noRawat, kdPoli, kdDokter string }
	hapusCalledWith  struct{ noRawat, kdDokter string }
}

func (m *mockRepository) DaftarOpsiPoliDokter(ctx context.Context, kodeDokterLogin, keyword string) ([]OpsiPoliDokter, error) {
	return m.opsiList, m.opsiErr
}

func (m *mockRepository) DaftarRujukanInternalByNoRawat(ctx context.Context, noRawat string) ([]RujukanInternal, error) {
	return m.rujukanList, m.rujukanErr
}

func (m *mockRepository) CekRujukanInternalAda(ctx context.Context, noRawat, kdDokter string) (bool, error) {
	return m.adaResult, m.adaErr
}

func (m *mockRepository) SimpanRujukanInternal(ctx context.Context, noRawat, kdPoli, kdDokter string) error {
	m.simpanCalledWith.noRawat = noRawat
	m.simpanCalledWith.kdPoli = kdPoli
	m.simpanCalledWith.kdDokter = kdDokter
	return m.simpanErr
}

func (m *mockRepository) HapusRujukanInternal(ctx context.Context, noRawat, kdDokter string) error {
	m.hapusCalledWith.noRawat = noRawat
	m.hapusCalledWith.kdDokter = kdDokter
	return m.hapusErr
}

type mockRawatJalanService struct {
	rawatjalan.Service
	tglReg, jamReg string
	exists         bool
	err            error
	kdPenjamin     string
	statusBayar    string
}

func (m *mockRawatJalanService) DaftarAntreanDokter(ctx context.Context, kodeDokter string, filter rawatjalan.FilterAntreanDokter) ([]rawatjalan.KunjunganRawatJalan, shared.PaginationMeta, error) {
	return nil, shared.PaginationMeta{}, nil
}
func (m *mockRawatJalanService) DetailKunjungan(ctx context.Context, noRawat string, kodeDokter string) (*rawatjalan.KunjunganRawatJalan, error) {
	return nil, nil
}
func (m *mockRawatJalanService) RiwayatKunjunganPasien(ctx context.Context, noRM string) ([]rawatjalan.KunjunganRawatJalan, error) {
	return nil, nil
}
func (m *mockRawatJalanService) GetWaktuRegistrasi(ctx context.Context, noRawat string) (string, string, bool, error) {
	return m.tglReg, m.jamReg, m.exists, m.err
}
func (m *mockRawatJalanService) GetInfoRegistrasi(ctx context.Context, noRawat string) (*rawatjalan.InfoRegistrasiPasien, error) {
	if m.err != nil {
		return nil, m.err
	}
	if !m.exists {
		return nil, nil
	}
	kdPenjamin := m.kdPenjamin
	if kdPenjamin == "" {
		kdPenjamin = "UMU"
	}
	sttsBayar := m.statusBayar
	if sttsBayar == "" {
		sttsBayar = "Belum Bayar"
	}
	return &rawatjalan.InfoRegistrasiPasien{
		TanggalRegistrasi: m.tglReg,
		JamRegistrasi:     m.jamReg,
		KodePenjamin:      kdPenjamin,
		StatusBayar:       sttsBayar,
	}, nil
}
func (m *mockRawatJalanService) DaftarStatusPemeriksaan(ctx context.Context) []rawatjalan.OpsiReferensi {
	return nil
}
func (m *mockRawatJalanService) DaftarStatusLanjut(ctx context.Context) []rawatjalan.OpsiReferensi {
	return nil
}
func (m *mockRawatJalanService) DaftarStatusBayar(ctx context.Context) []rawatjalan.OpsiReferensi {
	return nil
}
func (m *mockRawatJalanService) DaftarJenisAntrean(ctx context.Context) []rawatjalan.OpsiReferensi {
	return nil
}
func (m *mockRawatJalanService) PanggilAntrean(ctx context.Context, noRawat string, kdPoli string, kodeDokter string) (*rawatjalan.HasilPanggilAntrean, error) {
	return nil, nil
}

func TestDaftarOpsiPoliDokter_Success(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{
		opsiList: []OpsiPoliDokter{
			{InfoPoliDokter: InfoPoliDokter{KodePoli: "INT", NamaPoli: "Penyakit Dalam", KodeDokter: "DR002", NamaDokter: "dr. Sp.PD"}},
		},
	}
	svc := NewService(repo, &mockRawatJalanService{}, 48, log)

	result, err := svc.DaftarOpsiPoliDokter(context.Background(), "DR001", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}
	if result[0].KodePoli != "INT" || result[0].KodeDokter != "DR002" {
		t.Fatalf("unexpected item content: %+v", result[0])
	}
}

func TestDaftarOpsiPoliDokter_WithKeyword(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{
		opsiList: []OpsiPoliDokter{
			{InfoPoliDokter: InfoPoliDokter{KodePoli: "INT", NamaPoli: "Penyakit Dalam", KodeDokter: "DR002", NamaDokter: "dr. Sp.PD"}},
		},
	}
	svc := NewService(repo, &mockRawatJalanService{}, 48, log)

	result, err := svc.DaftarOpsiPoliDokter(context.Background(), "DR001", "Penyakit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}
}

func TestDaftarRujukanInternal_Success(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{
		rujukanList: []RujukanInternal{
			{NoRawat: "2026/04/22/000001", InfoPoliDokter: InfoPoliDokter{KodePoli: "INT", NamaPoli: "Penyakit Dalam", KodeDokter: "DR002", NamaDokter: "dr. Sp.PD"}},
		},
	}
	svc := NewService(repo, &mockRawatJalanService{}, 48, log)

	result, err := svc.DaftarRujukanInternal(context.Background(), "2026/04/22/000001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}
	if result[0].NoRawat != "2026/04/22/000001" {
		t.Errorf("expected no_rawat 2026/04/22/000001, got %s", result[0].NoRawat)
	}
}

func TestSimpanRujukanInternal_CannotReferToSelf(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{}
	svc := NewService(repo, &mockRawatJalanService{}, 48, log)

	_, err := svc.SimpanRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "INT", "DR001")
	if err == nil {
		t.Fatal("expected error when referring to self, got nil")
	}
	var valErr apperror.ValidationError
	if !errors.As(err, &valErr) || valErr["id_tujuan"] == "" {
		t.Fatalf("expected validation error on id_tujuan, got %v", err)
	}
}

func TestSimpanRujukanInternal_DuplicateReferral(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{
		adaResult: true,
	}
	today := time.Now().Format("2006-01-02")
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, 48, log)

	_, err := svc.SimpanRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "INT", "DR002")
	if err == nil {
		t.Fatal("expected error for duplicate referral, got nil")
	}
	var valErr apperror.ValidationError
	if !errors.As(err, &valErr) || valErr["id_tujuan"] == "" {
		t.Fatalf("expected validation error on duplicate referral, got %v", err)
	}
}

func TestSimpanRujukanInternal_Exceeds48Hours(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{adaResult: false}
	pastDate := time.Now().Add(-72 * time.Hour).Format("2006-01-02")
	rjSvc := &mockRawatJalanService{tglReg: pastDate, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, 48, log)

	_, err := svc.SimpanRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "INT", "DR002")
	if err == nil {
		t.Fatal("expected 48 hours forbidden error, got nil")
	}
	var forbidden *apperror.ForbiddenError
	if !errors.As(err, &forbidden) {
		t.Fatalf("expected ForbiddenError, got %v", err)
	}
	if forbidden.Message != "Rujukan internal tidak dapat disimpan, melebihi batas 48 jam" {
		t.Errorf("unexpected error message: %s", forbidden.Message)
	}
}

func TestSimpanRujukanInternal_Success(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{
		adaResult: false,
		rujukanList: []RujukanInternal{
			{NoRawat: "2026/04/22/000001", InfoPoliDokter: InfoPoliDokter{KodePoli: "INT", NamaPoli: "Penyakit Dalam", KodeDokter: "DR002", NamaDokter: "dr. Sp.PD"}},
		},
	}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, 48, log)

	result, err := svc.SimpanRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "INT", "DR002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if repo.simpanCalledWith.kdDokter != "DR002" || repo.simpanCalledWith.kdPoli != "INT" {
		t.Fatalf("unexpected save arguments: %+v", repo.simpanCalledWith)
	}
}

func TestHapusRujukanInternal_Success(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, 48, log)

	err := svc.HapusRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "DR002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.hapusCalledWith.noRawat != "2026/04/22/000001" || repo.hapusCalledWith.kdDokter != "DR002" {
		t.Fatalf("unexpected delete arguments: %+v", repo.hapusCalledWith)
	}
}

func TestHapusRujukanInternal_NotFound(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{hapusErr: sql.ErrNoRows}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, 48, log)

	err := svc.HapusRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "DR002")
	if err == nil {
		t.Fatal("expected not found error, got nil")
	}
	var notFound *apperror.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestHapusRujukanInternal_Exceeds48Hours(t *testing.T) {
	log := logger.New()
	repo := &mockRepository{}
	pastDate := time.Now().Add(-72 * time.Hour).Format("2006-01-02")
	rjSvc := &mockRawatJalanService{tglReg: pastDate, jamReg: "08:00:00", exists: true}
	svc := NewService(repo, rjSvc, 48, log)

	err := svc.HapusRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "DR002")
	if err == nil {
		t.Fatal("expected 48 hours forbidden error, got nil")
	}
	var forbidden *apperror.ForbiddenError
	if !errors.As(err, &forbidden) {
		t.Fatalf("expected ForbiddenError, got %v", err)
	}
	if forbidden.Message != "Rujukan internal tidak dapat dihapus, melebihi batas 48 jam" {
		t.Errorf("unexpected error message: %s", forbidden.Message)
	}
}

func TestSimpanRujukanInternal_BPJSSudahBayar(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{adaResult: false}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true, kdPenjamin: "BPJ", statusBayar: "Sudah Bayar"}
	svc := NewService(repo, rjSvc, 48, log)

	_, err := svc.SimpanRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "INT", "DR002")
	if err == nil {
		t.Fatal("expected error for BPJS sudah bayar on simpan, got nil")
	}
	var bErr *apperror.BusinessError
	if !errors.As(err, &bErr) {
		t.Fatalf("expected BusinessError, got %v", err)
	}
	if bErr.Message != "Pasien BPJS sudah bayar, rujukan tidak dapat disimpan" {
		t.Errorf("unexpected error message: %s", bErr.Message)
	}
}

func TestHapusRujukanInternal_BPJSSudahBayar(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")
	repo := &mockRepository{}
	rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true, kdPenjamin: "BPJ", statusBayar: "Sudah Bayar"}
	svc := NewService(repo, rjSvc, 48, log)

	err := svc.HapusRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "DR002")
	if err == nil {
		t.Fatal("expected error for BPJS sudah bayar on hapus, got nil")
	}
	var bErr *apperror.BusinessError
	if !errors.As(err, &bErr) {
		t.Fatalf("expected BusinessError, got %v", err)
	}
	if bErr.Message != "Pasien BPJS sudah bayar, rujukan tidak dapat dihapus" {
		t.Errorf("unexpected error message: %s", bErr.Message)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
