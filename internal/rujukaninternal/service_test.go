package rujukaninternal

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"

	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/shared/apperror"
)

type mockRepository struct {
	opsiList         []OpsiPoliDokter
	opsiErr          error
	rujukanList      []RujukanInternal
	rujukanErr       error
	opsiAktifSet     bool
	opsiAktif        bool
	opsiAktifErr     error
	adaResult        bool
	adaErr           error
	poliAdaSet       bool
	poliAdaResult    bool
	poliAdaErr       error
	aktivitasKonsul  bool
	aktivitasErr     error
	simpanErr        error
	hapusErr         error
	simpanCalledWith struct{ noRawat, kdPoli, kdDokter string }
	hapusCalledWith  struct{ noRawat, kdPoli, kdDokter string }
}

func (m *mockRepository) DaftarOpsiPoliDokter(ctx context.Context, kodeDokterLogin, keyword string) ([]OpsiPoliDokter, error) {
	return m.opsiList, m.opsiErr
}

func (m *mockRepository) DaftarRujukanInternalByNoRawat(ctx context.Context, noRawat string) ([]RujukanInternal, error) {
	return m.rujukanList, m.rujukanErr
}

func (m *mockRepository) CekOpsiPoliDokterAktif(ctx context.Context, kdPoli, kdDokter string) (bool, error) {
	if m.opsiAktifErr != nil {
		return false, m.opsiAktifErr
	}
	if !m.opsiAktifSet {
		return true, nil
	}
	return m.opsiAktif, nil
}

func (m *mockRepository) CekRujukanInternalAda(ctx context.Context, noRawat, kdDokter string) (bool, error) {
	return m.adaResult, m.adaErr
}

func (m *mockRepository) CekRujukanInternalPoliAda(ctx context.Context, noRawat, kdPoli, kdDokter string) (bool, error) {
	if m.poliAdaErr != nil {
		return false, m.poliAdaErr
	}
	if !m.poliAdaSet {
		return true, nil
	}
	return m.poliAdaResult, nil
}

func (m *mockRepository) CekAktivitasDokterKonsulen(ctx context.Context, noRawat, kdDokter string) (bool, error) {
	return m.aktivitasKonsul, m.aktivitasErr
}

func (m *mockRepository) SimpanRujukanInternal(ctx context.Context, noRawat, kdPoli, kdDokter string) error {
	m.simpanCalledWith.noRawat = noRawat
	m.simpanCalledWith.kdPoli = kdPoli
	m.simpanCalledWith.kdDokter = kdDokter
	return m.simpanErr
}

func (m *mockRepository) HapusRujukanInternal(ctx context.Context, noRawat, kdPoli, kdDokter string) error {
	m.hapusCalledWith.noRawat = noRawat
	m.hapusCalledWith.kdPoli = kdPoli
	m.hapusCalledWith.kdDokter = kdDokter
	return m.hapusErr
}

type mockRawatJalanService struct {
	rawatjalan.Service
	tglReg, jamReg    string
	exists            bool
	err               error
	kdPenjamin        string
	statusBayar       string
	dpjpUtama         string
	statusPemeriksaan string
	statusLanjut      string
	cekAksesFn        func(ctx context.Context, noRawat, kodeDokter string) (bool, bool, error)
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
	dpjp := m.dpjpUtama
	if dpjp == "" {
		dpjp = "DR001"
	}
	stts := m.statusPemeriksaan
	if stts == "" {
		stts = "Belum"
	}
	lanjut := m.statusLanjut
	if lanjut == "" {
		lanjut = "Ralan"
	}
	return &rawatjalan.InfoRegistrasiPasien{
		TanggalRegistrasi: m.tglReg,
		JamRegistrasi:     m.jamReg,
		KodePenjamin:      kdPenjamin,
		StatusBayar:       sttsBayar,
		KodeDokter:        dpjp,
		StatusPemeriksaan: stts,
		StatusLanjut:      lanjut,
	}, nil
}

func (m *mockRawatJalanService) CekAksesKunjungan(ctx context.Context, noRawat string, kodeDokter string) (bool, bool, error) {
	if m.cekAksesFn != nil {
		return m.cekAksesFn(ctx, noRawat, kodeDokter)
	}
	if !m.exists {
		return false, false, nil
	}
	dpjp := m.dpjpUtama
	if dpjp == "" {
		dpjp = "DR001"
	}
	return true, kodeDokter == dpjp, nil
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
	if len(result) != 1 || result[0].KodePoli != "INT" || result[0].KodeDokter != "DR002" {
		t.Fatalf("unexpected result: %+v", result)
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
	if len(result) != 1 || result[0].NoRawat != "2026/04/22/000001" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestSimpanRujukanInternal_CannotReferToSelf(t *testing.T) {
	log := logger.New()
	svc := NewService(&mockRepository{}, &mockRawatJalanService{}, 48, log)

	_, err := svc.SimpanRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "INT", "DR001")
	var valErr apperror.ValidationError
	if !errors.As(err, &valErr) || valErr["id_tujuan"] == "" {
		t.Fatalf("expected validation error on id_tujuan, got %v", err)
	}
}

func TestSimpanRujukanInternal_AuthorizationAndRules(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")

	t.Run("External doctor without access is rejected with 403 Forbidden", func(t *testing.T) {
		rjSvc := &mockRawatJalanService{
			tglReg: today, jamReg: "08:00:00", exists: true, dpjpUtama: "DR001",
			cekAksesFn: func(_ context.Context, _, _ string) (bool, bool, error) {
				return true, false, nil
			},
		}
		svc := NewService(&mockRepository{}, rjSvc, 48, log)

		_, err := svc.SimpanRujukanInternal(context.Background(), "DR999", "2026/04/22/000001", "INT", "DR002")
		var forbidden *apperror.ForbiddenError
		if !errors.As(err, &forbidden) {
			t.Fatalf("expected ForbiddenError for external doctor, got %v", err)
		}
	})

	t.Run("Consulting doctor can refer to another specialist, even after admin switches visit to Ranap within 48h", func(t *testing.T) {
		repo := &mockRepository{
			rujukanList: []RujukanInternal{
				{NoRawat: "2026/04/22/000001", InfoPoliDokter: InfoPoliDokter{KodePoli: "JTG", KodeDokter: "DR003"}},
			},
		}
		rjSvc := &mockRawatJalanService{
			tglReg: today, jamReg: "08:00:00", exists: true, dpjpUtama: "DR001", statusLanjut: "Ranap",
			cekAksesFn: func(_ context.Context, _, kd string) (bool, bool, error) {
				return true, kd == "DR002", nil
			},
		}
		svc := NewService(repo, rjSvc, 48, log)

		got, err := svc.SimpanRujukanInternal(context.Background(), "DR002", "2026/04/22/000001", "JTG", "DR003")
		if err != nil || got == nil || got.KodeDokter != "DR003" {
			t.Fatalf("expected consulting doctor referral to succeed within 48h even if Ranap, got=%+v err=%v", got, err)
		}
	})

	t.Run("Consulting doctor cannot refer back to main DPJP", func(t *testing.T) {
		rjSvc := &mockRawatJalanService{
			tglReg: today, jamReg: "08:00:00", exists: true, dpjpUtama: "DR001",
			cekAksesFn: func(_ context.Context, _, kd string) (bool, bool, error) {
				return true, kd == "DR002", nil
			},
		}
		svc := NewService(&mockRepository{}, rjSvc, 48, log)

		_, err := svc.SimpanRujukanInternal(context.Background(), "DR002", "2026/04/22/000001", "UMU", "DR001")
		var valErr apperror.ValidationError
		if !errors.As(err, &valErr) || valErr["id_tujuan"] == "" {
			t.Fatalf("expected validation error when referring back to main DPJP, got %v", err)
		}
	})

	t.Run("Inactive target poli or doctor is rejected with 400 ValidationError", func(t *testing.T) {
		repo := &mockRepository{opsiAktifSet: true, opsiAktif: false}
		rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true, dpjpUtama: "DR001"}
		svc := NewService(repo, rjSvc, 48, log)

		_, err := svc.SimpanRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "NONAKTIF", "DR002")
		var valErr apperror.ValidationError
		if !errors.As(err, &valErr) || valErr["id_tujuan"] == "" {
			t.Fatalf("expected validation error for inactive poli/dokter, got %v", err)
		}
	})

	t.Run("MySQL 1062 duplicate race is mapped to 400 ValidationError instead of 500", func(t *testing.T) {
		repo := &mockRepository{
			simpanErr: &mysqlDriver.MySQLError{Number: 1062, Message: "Duplicate entry"},
		}
		rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true, dpjpUtama: "DR001"}
		svc := NewService(repo, rjSvc, 48, log)

		_, err := svc.SimpanRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "INT", "DR002")
		var valErr apperror.ValidationError
		if !errors.As(err, &valErr) || valErr["id_tujuan"] != "Tujuan rujukan sudah dipilih" {
			t.Fatalf("expected 1062 race mapped to ValidationError, got %v", err)
		}
	})

	t.Run("Cancelled visit (stts = Batal) is rejected with 400 BusinessError", func(t *testing.T) {
		rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true, statusPemeriksaan: "Batal"}
		svc := NewService(&mockRepository{}, rjSvc, 48, log)

		_, err := svc.SimpanRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "INT", "DR002")
		var bErr *apperror.BusinessError
		if !errors.As(err, &bErr) {
			t.Fatalf("expected BusinessError for Batal visit, got %v", err)
		}
	})
}

func TestSimpanRujukanInternal_DuplicateAndWindowAndBPJS(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")

	t.Run("Duplicate referral", func(t *testing.T) {
		repo := &mockRepository{adaResult: true}
		rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true}
		svc := NewService(repo, rjSvc, 48, log)

		_, err := svc.SimpanRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "INT", "DR002")
		var valErr apperror.ValidationError
		if !errors.As(err, &valErr) || valErr["id_tujuan"] == "" {
			t.Fatalf("expected validation error, got %v", err)
		}
	})

	t.Run("Exceeds 48 hours", func(t *testing.T) {
		pastDate := time.Now().Add(-72 * time.Hour).Format("2006-01-02")
		rjSvc := &mockRawatJalanService{tglReg: pastDate, jamReg: "08:00:00", exists: true}
		svc := NewService(&mockRepository{}, rjSvc, 48, log)

		_, err := svc.SimpanRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "INT", "DR002")
		var forbidden *apperror.ForbiddenError
		if !errors.As(err, &forbidden) || forbidden.Message != "Rujukan internal tidak dapat disimpan, melebihi batas 48 jam" {
			t.Fatalf("expected 48h ForbiddenError, got %v", err)
		}
	})

	t.Run("BPJS Sudah Bayar", func(t *testing.T) {
		rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true, kdPenjamin: "BPJ", statusBayar: "Sudah Bayar"}
		svc := NewService(&mockRepository{}, rjSvc, 48, log)

		_, err := svc.SimpanRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "INT", "DR002")
		var bErr *apperror.BusinessError
		if !errors.As(err, &bErr) {
			t.Fatalf("expected BusinessError, got %v", err)
		}
	})
}

func TestHapusRujukanInternal_AuthorizationAndClinicalLock(t *testing.T) {
	log := logger.New()
	today := time.Now().Format("2006-01-02")

	t.Run("DPJP utama deletes unused referral with matching KodePoli", func(t *testing.T) {
		repo := &mockRepository{}
		rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true, dpjpUtama: "DR001"}
		svc := NewService(repo, rjSvc, 48, log)

		if err := svc.HapusRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "INT", "DR002"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.hapusCalledWith.noRawat != "2026/04/22/000001" || repo.hapusCalledWith.kdPoli != "INT" || repo.hapusCalledWith.kdDokter != "DR002" {
			t.Fatalf("unexpected delete arguments: %+v", repo.hapusCalledWith)
		}
	})

	t.Run("Consulting doctor can delete another unacted referral, but CANNOT delete their own referral", func(t *testing.T) {
		repo := &mockRepository{}
		rjSvc := &mockRawatJalanService{
			tglReg: today, jamReg: "08:00:00", exists: true, dpjpUtama: "DR001",
			cekAksesFn: func(_ context.Context, _, kd string) (bool, bool, error) {
				return true, kd == "DR002", nil
			},
		}
		svc := NewService(repo, rjSvc, 48, log)

		errSelf := svc.HapusRujukanInternal(context.Background(), "DR002", "2026/04/22/000001", "INT", "DR002")
		var forbidden *apperror.ForbiddenError
		if !errors.As(errSelf, &forbidden) {
			t.Fatalf("expected ForbiddenError when consulting doctor deletes own referral, got %v", errSelf)
		}

		if errOther := svc.HapusRujukanInternal(context.Background(), "DR002", "2026/04/22/000001", "JTG", "DR003"); errOther != nil {
			t.Fatalf("expected consulting doctor to be able to cancel unacted referral DR003, got %v", errOther)
		}
	})

	t.Run("External doctor is rejected with 403 Forbidden", func(t *testing.T) {
		rjSvc := &mockRawatJalanService{
			tglReg: today, jamReg: "08:00:00", exists: true, dpjpUtama: "DR001",
			cekAksesFn: func(_ context.Context, _, _ string) (bool, bool, error) {
				return true, false, nil
			},
		}
		svc := NewService(&mockRepository{}, rjSvc, 48, log)

		err := svc.HapusRujukanInternal(context.Background(), "DR999", "2026/04/22/000001", "INT", "DR002")
		var forbidden *apperror.ForbiddenError
		if !errors.As(err, &forbidden) {
			t.Fatalf("expected ForbiddenError for external doctor, got %v", err)
		}
	})

	t.Run("Blocked when target consulting doctor already filled SOAPIE, Penilaian Medis, or Resep", func(t *testing.T) {
		repo := &mockRepository{aktivitasKonsul: true}
		rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true, dpjpUtama: "DR001"}
		svc := NewService(repo, rjSvc, 48, log)

		err := svc.HapusRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "INT", "DR002")
		var forbidden *apperror.ForbiddenError
		if !errors.As(err, &forbidden) {
			t.Fatalf("expected ForbiddenError when target doctor already filled clinical records, got %v", err)
		}
		if repo.hapusCalledWith.noRawat != "" {
			t.Fatal("expected HapusRujukanInternal repo NOT to be called when clinical records exist")
		}
	})

	t.Run("Mismatched KodePoli or missing referral returns 404 NotFoundError before clinical check", func(t *testing.T) {
		repo := &mockRepository{poliAdaSet: true, poliAdaResult: false, aktivitasKonsul: true}
		rjSvc := &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true, dpjpUtama: "DR001"}
		svc := NewService(repo, rjSvc, 48, log)

		err := svc.HapusRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "SALAH_POLI", "DR002")
		var notFound *apperror.NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("expected NotFoundError when (no_rawat, kd_poli, kd_dokter) does not exist, got %v", err)
		}

		repoRace := &mockRepository{hapusErr: sql.ErrNoRows}
		svcRace := NewService(repoRace, rjSvc, 48, log)
		if errRace := svcRace.HapusRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "INT", "DR002"); !errors.As(errRace, &notFound) {
			t.Fatalf("expected NotFoundError when DELETE affects 0 rows concurrently, got %v", errRace)
		}
	})

	t.Run("Exceeds 48h, BPJS Sudah Bayar, and Batal visit rejected on delete", func(t *testing.T) {
		pastDate := time.Now().Add(-72 * time.Hour).Format("2006-01-02")
		svc48 := NewService(&mockRepository{}, &mockRawatJalanService{tglReg: pastDate, jamReg: "08:00:00", exists: true}, 48, log)
		var forbidden *apperror.ForbiddenError
		if err := svc48.HapusRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "INT", "DR002"); !errors.As(err, &forbidden) {
			t.Fatalf("expected 48h ForbiddenError, got %v", err)
		}

		svcBPJS := NewService(&mockRepository{}, &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true, kdPenjamin: "BPJ", statusBayar: "Sudah Bayar"}, 48, log)
		var bErr *apperror.BusinessError
		if err := svcBPJS.HapusRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "INT", "DR002"); !errors.As(err, &bErr) {
			t.Fatalf("expected BPJS BusinessError, got %v", err)
		}

		svcBatal := NewService(&mockRepository{}, &mockRawatJalanService{tglReg: today, jamReg: "08:00:00", exists: true, statusPemeriksaan: "Batal"}, 48, log)
		if err := svcBatal.HapusRujukanInternal(context.Background(), "DR001", "2026/04/22/000001", "INT", "DR002"); !errors.As(err, &bErr) {
			t.Fatalf("expected Batal BusinessError, got %v", err)
		}
	})
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
