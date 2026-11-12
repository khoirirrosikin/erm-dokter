package rujukaninternal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"erm-dokter/internal/pkg/database"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

type Service interface {
	DaftarOpsiPoliDokter(ctx context.Context, kodeDokterLogin, keyword string) ([]OpsiPoliDokter, error)
	DaftarRujukanInternal(ctx context.Context, noRawat string) ([]RujukanInternal, error)
	SimpanRujukanInternal(ctx context.Context, kodeDokterLogin, noRawat, targetKodePoli, targetKodeDokter string) (*RujukanInternal, error)
	HapusRujukanInternal(ctx context.Context, kodeDokterLogin, noRawat, targetKodePoli, targetKodeDokter string) error
}

type service struct {
	repo              Repository
	rawatJalanService rawatjalan.Service
	maxEditJam        int
	log               *logger.Logger
}

func NewService(repo Repository, rawatJalanService rawatjalan.Service, maxEditJam int, log *logger.Logger) Service {
	if maxEditJam <= 0 {
		maxEditJam = 48
	}
	return &service{
		repo:              repo,
		rawatJalanService: rawatJalanService,
		maxEditJam:        maxEditJam,
		log:               log,
	}
}

func (s *service) DaftarOpsiPoliDokter(ctx context.Context, kodeDokterLogin, keyword string) ([]OpsiPoliDokter, error) {
	cleanKeyword := strings.TrimSpace(keyword)
	list, err := s.repo.DaftarOpsiPoliDokter(ctx, strings.TrimSpace(kodeDokterLogin), cleanKeyword)
	if err != nil {
		s.log.Error("Gagal mengambil opsi poli dokter untuk dokter %s (keyword: %s): %v", kodeDokterLogin, cleanKeyword, err)
		return nil, err
	}
	return list, nil
}

func (s *service) DaftarRujukanInternal(ctx context.Context, noRawat string) ([]RujukanInternal, error) {
	list, err := s.repo.DaftarRujukanInternalByNoRawat(ctx, strings.TrimSpace(noRawat))
	if err != nil {
		s.log.Error("Gagal mengambil daftar rujukan internal no_rawat %s: %v", noRawat, err)
		return nil, err
	}
	return list, nil
}

func (s *service) SimpanRujukanInternal(ctx context.Context, kodeDokterLogin, noRawat, targetKodePoli, targetKodeDokter string) (*RujukanInternal, error) {
	kodeDokterLogin = strings.TrimSpace(kodeDokterLogin)
	noRawat = strings.TrimSpace(noRawat)
	targetKodePoli = strings.TrimSpace(targetKodePoli)
	targetKodeDokter = strings.TrimSpace(targetKodeDokter)

	if targetKodeDokter == kodeDokterLogin {
		s.log.Warn("Dokter %s mencoba membuat rujukan internal ke dokter pemeriksa pada no_rawat %s", kodeDokterLogin, noRawat)
		return nil, apperror.ValidationError{"id_tujuan": "Tujuan rujukan tidak dapat dipilih"}
	}

	infoRegistrasi, err := s.validasiRegistrasiDanStatus(ctx, noRawat, "disimpan")
	if err != nil {
		return nil, err
	}

	if err := s.validasiAksesDokter(ctx, noRawat, kodeDokterLogin, infoRegistrasi.KodeDokter, "merujuk pasien pada kunjungan ini"); err != nil {
		return nil, err
	}

	if infoRegistrasi.KodeDokter != "" && targetKodeDokter == infoRegistrasi.KodeDokter {
		s.log.Warn("Dokter konsulen %s mencoba merujuk balik ke DPJP utama %s pada no_rawat %s", kodeDokterLogin, infoRegistrasi.KodeDokter, noRawat)
		return nil, apperror.ValidationError{"id_tujuan": "Tidak dapat merujuk kembali ke dokter utama kunjungan"}
	}

	opsiAktif, err := s.repo.CekOpsiPoliDokterAktif(ctx, targetKodePoli, targetKodeDokter)
	if err != nil {
		s.log.Error("Gagal memeriksa keaktifan opsi poli '%s' dokter '%s': %v", targetKodePoli, targetKodeDokter, err)
		return nil, err
	}
	if !opsiAktif {
		return nil, apperror.ValidationError{"id_tujuan": "Tujuan poliklinik atau dokter rujukan tidak ditemukan atau tidak aktif"}
	}

	ada, err := s.repo.CekRujukanInternalAda(ctx, noRawat, targetKodeDokter)
	if err != nil {
		s.log.Error("Gagal memeriksa duplikasi rujukan internal no_rawat %s dokter %s: %v", noRawat, targetKodeDokter, err)
		return nil, err
	}
	if ada {
		s.log.Warn("Percobaan duplikasi rujukan internal no_rawat %s ke dokter %s", noRawat, targetKodeDokter)
		return nil, apperror.ValidationError{"id_tujuan": "Tujuan rujukan sudah dipilih"}
	}

	if err := s.repo.SimpanRujukanInternal(ctx, noRawat, targetKodePoli, targetKodeDokter); err != nil {
		if database.IsDuplicateEntry(err) {
			return nil, apperror.ValidationError{"id_tujuan": "Tujuan rujukan sudah dipilih"}
		}
		s.log.Error("Gagal menyimpan data rujukan internal no_rawat '%s' poli '%s' dokter '%s': %v", noRawat, targetKodePoli, targetKodeDokter, err)
		return nil, err
	}

	s.log.Info("Berhasil menyimpan data rujukan internal no_rawat '%s' ke poli '%s' dokter '%s' oleh dokter '%s'", noRawat, targetKodePoli, targetKodeDokter, kodeDokterLogin)

	list, err := s.repo.DaftarRujukanInternalByNoRawat(ctx, noRawat)
	if err != nil {
		return nil, nil
	}

	for i := range list {
		if list[i].KodeDokter == targetKodeDokter && list[i].KodePoli == targetKodePoli {
			return &list[i], nil
		}
	}

	return nil, nil
}

func (s *service) HapusRujukanInternal(ctx context.Context, kodeDokterLogin, noRawat, targetKodePoli, targetKodeDokter string) error {
	kodeDokterLogin = strings.TrimSpace(kodeDokterLogin)
	noRawat = strings.TrimSpace(noRawat)
	targetKodePoli = strings.TrimSpace(targetKodePoli)
	targetKodeDokter = strings.TrimSpace(targetKodeDokter)

	infoRegistrasi, err := s.validasiRegistrasiDanStatus(ctx, noRawat, "dihapus")
	if err != nil {
		return err
	}

	if err := s.validasiAksesDokter(ctx, noRawat, kodeDokterLogin, infoRegistrasi.KodeDokter, "menghapus rujukan internal pada kunjungan ini"); err != nil {
		return err
	}

	if targetKodeDokter == kodeDokterLogin {
		return apperror.NewForbiddenError(
			"Anda tidak dapat menghapus rujukan internal yang ditujukan kepada Anda sendiri",
			fmt.Sprintf("Dokter konsulen '%s' mencoba menghapus rujukan dirinya sendiri pada no_rawat '%s'", kodeDokterLogin, noRawat),
		)
	}

	ada, err := s.repo.CekRujukanInternalPoliAda(ctx, noRawat, targetKodePoli, targetKodeDokter)
	if err != nil {
		s.log.Error("Gagal memeriksa keberadaan rujukan internal no_rawat '%s' poli '%s' dokter '%s': %v", noRawat, targetKodePoli, targetKodeDokter, err)
		return err
	}
	if !ada {
		return apperror.NewNotFoundError(
			"Data rujukan internal tidak ditemukan",
			fmt.Sprintf("Data rujukan internal tidak ditemukan untuk no_rawat '%s' poli '%s' dokter '%s'", noRawat, targetKodePoli, targetKodeDokter),
		)
	}

	sudahMengisi, err := s.repo.CekAktivitasDokterKonsulen(ctx, noRawat, targetKodeDokter)
	if err != nil {
		s.log.Error("Gagal memeriksa aktivitas klinis dokter konsulen '%s' pada no_rawat '%s': %v", targetKodeDokter, noRawat, err)
		return err
	}
	if sudahMengisi {
		return apperror.NewForbiddenError(
			"Rujukan internal tidak dapat dihapus karena dokter tujuan sudah mengisi pemeriksaan, penilaian medis, atau resep",
			fmt.Sprintf("Penghapusan rujukan internal no_rawat '%s' dokter '%s' ditolak karena dokter konsulen sudah memiliki rekam medis/resep", noRawat, targetKodeDokter),
		)
	}

	if err := s.repo.HapusRujukanInternal(ctx, noRawat, targetKodePoli, targetKodeDokter); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError(
				"Data rujukan internal tidak ditemukan",
				fmt.Sprintf("Data rujukan internal tidak ditemukan untuk no_rawat '%s' poli '%s' dokter '%s'", noRawat, targetKodePoli, targetKodeDokter),
			)
		}
		s.log.Error("Gagal menghapus data rujukan internal no_rawat '%s' poli '%s' dokter '%s' oleh dokter '%s': %v", noRawat, targetKodePoli, targetKodeDokter, kodeDokterLogin, err)
		return err
	}

	s.log.Info("Berhasil menghapus data rujukan internal no_rawat '%s' poli '%s' dokter '%s' oleh dokter '%s'", noRawat, targetKodePoli, targetKodeDokter, kodeDokterLogin)
	return nil
}

func (s *service) validasiAksesDokter(ctx context.Context, noRawat, kodeDokterLogin, dpjpUtama, deskripsiAksi string) error {
	if dpjpUtama != "" && kodeDokterLogin == strings.TrimSpace(dpjpUtama) {
		return nil
	}

	exists, authorized, err := s.rawatJalanService.CekAksesKunjungan(ctx, noRawat, kodeDokterLogin)
	if err != nil {
		s.log.Error("Gagal memeriksa akses kunjungan no_rawat '%s' untuk dokter '%s': %v", noRawat, kodeDokterLogin, err)
		return err
	}
	if !exists {
		return apperror.NewNotFoundError(
			"Data registrasi kunjungan pasien tidak ditemukan",
			fmt.Sprintf("Data registrasi kunjungan pasien tidak ditemukan untuk no_rawat '%s'", noRawat),
		)
	}
	if !authorized {
		return apperror.NewForbiddenError(
			fmt.Sprintf("Anda tidak memiliki akses untuk %s", deskripsiAksi),
			fmt.Sprintf("Dokter '%s' ditolak %s pada no_rawat '%s'", kodeDokterLogin, deskripsiAksi, noRawat),
		)
	}
	return nil
}

func (s *service) validasiRegistrasiDanStatus(ctx context.Context, noRawat, aksi string) (*rawatjalan.InfoRegistrasiPasien, error) {
	infoRegistrasi, err := s.rawatJalanService.GetInfoRegistrasi(ctx, noRawat)
	if err != nil {
		return nil, err
	}
	if infoRegistrasi == nil {
		return nil, apperror.NewNotFoundError(
			"Data registrasi kunjungan pasien tidak ditemukan",
			fmt.Sprintf("Data registrasi kunjungan pasien tidak ditemukan untuk no_rawat '%s'", noRawat),
		)
	}

	if strings.EqualFold(strings.TrimSpace(infoRegistrasi.StatusPemeriksaan), "Batal") {
		return nil, apperror.NewBusinessError(fmt.Sprintf("Kunjungan pasien berstatus batal, rujukan tidak dapat %s", aksi))
	}

	if infoRegistrasi.StatusBayar == "Sudah Bayar" && infoRegistrasi.KodePenjamin == "BPJ" {
		return nil, apperror.NewBusinessError(fmt.Sprintf("Pasien BPJS sudah bayar, rujukan tidak dapat %s", aksi))
	}

	tanggalRegistrasi := infoRegistrasi.TanggalRegistrasi
	jamRegistrasi := infoRegistrasi.JamRegistrasi

	waktuRegistrasi, err := shared.ParseWaktu(tanggalRegistrasi, jamRegistrasi)
	if err != nil {
		s.log.Error("Gagal parse waktu registrasi no_rawat '%s' (%s %s): %v", noRawat, tanggalRegistrasi, jamRegistrasi, err)
		return nil, err
	}

	batasWaktu := waktuRegistrasi.Add(time.Duration(s.maxEditJam) * time.Hour)
	if time.Now().After(batasWaktu) {
		return nil, apperror.NewForbiddenError(
			fmt.Sprintf("Rujukan internal tidak dapat %s, melebihi batas %d jam", aksi, s.maxEditJam),
			fmt.Sprintf("Rujukan internal no_rawat '%s' ditolak untuk %s karena melewati batas %d jam dari registrasi (%s)", noRawat, aksi, s.maxEditJam, waktuRegistrasi.Format("2006-01-02 15:04:05")),
		)
	}

	return infoRegistrasi, nil
}
