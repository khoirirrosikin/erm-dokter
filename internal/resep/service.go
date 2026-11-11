package resep

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"erm-dokter/internal/obat"
	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/rawatinap"
	"erm-dokter/internal/rawatjalan"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

type Service interface {
	DaftarResep(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter FilterDaftarResep) ([]Resep, shared.PaginationMeta, error)
	DaftarResepByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterDaftarResep) ([]Resep, shared.PaginationMeta, error)
	DetailResep(ctx context.Context, noResep string) (*Resep, error)
	SimpanResep(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req SimpanResepRequest) (*Resep, error)
	HapusResep(ctx context.Context, kodeDokter, noRawat, noResep string, statusLanjut shared.StatusLanjut) error
	UpdateResep(ctx context.Context, kodeDokter, noRawat, noResep string, statusLanjut shared.StatusLanjut, req SimpanResepRequest) (*Resep, error)
	DaftarAturanPakai(ctx context.Context, keyword string) ([]AturanPakai, error)
	DaftarMetodeRacik(ctx context.Context) ([]MetodeRacik, error)
}

type service struct {
	repo              Repository
	rawatJalanService rawatjalan.Service
	rawatInapService  rawatinap.Service
	obatService       obat.Service
	maxEditJam        int
	log               *logger.Logger
}

func NewService(repo Repository, rawatJalanService rawatjalan.Service, rawatInapService rawatinap.Service, obatService obat.Service, maxEditJam int, log *logger.Logger) Service {
	if maxEditJam <= 0 {
		maxEditJam = 48
	}
	return &service{
		repo:              repo,
		rawatJalanService: rawatJalanService,
		rawatInapService:  rawatInapService,
		obatService:       obatService,
		maxEditJam:        maxEditJam,
		log:               log,
	}
}

func (s *service) DaftarResep(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut, filter FilterDaftarResep) ([]Resep, shared.PaginationMeta, error) {
	data, total, err := s.repo.DaftarResep(ctx, noRawat, statusLanjut, filter)
	if err != nil {
		s.log.Error("Gagal mengambil daftar resep untuk no_rawat %s: %v", noRawat, err)
		return nil, shared.PaginationMeta{}, err
	}

	return data, shared.NewPaginationMeta(total, filter.Page, filter.Limit), nil
}

func (s *service) DaftarResepByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterDaftarResep) ([]Resep, shared.PaginationMeta, error) {
	data, total, err := s.repo.DaftarResepByRM(ctx, noRM, statusLanjut, filter)
	if err != nil {
		s.log.Error("Gagal mengambil riwayat resep untuk no_rm %s: %v", noRM, err)
		return nil, shared.PaginationMeta{}, err
	}

	return data, shared.NewPaginationMeta(total, filter.Page, filter.Limit), nil
}

func (s *service) DetailResep(ctx context.Context, noResep string) (*Resep, error) {
	resep, err := s.repo.DetailResep(ctx, noResep)
	if err != nil {
		s.log.Error("Gagal mengambil detail resep %s: %v", noResep, err)
		return nil, err
	}
	if resep == nil {
		return nil, apperror.NewNotFoundError(
			"Resep obat tidak ditemukan",
			fmt.Sprintf("Data resep no_resep '%s' tidak ditemukan", noResep),
		)
	}

	return resep, nil
}

func (s *service) SimpanResep(ctx context.Context, kodeDokter string, statusLanjut shared.StatusLanjut, req SimpanResepRequest) (*Resep, error) {
	if err := s.validasiRegistrasiDanStatus(ctx, req.NoRawat, req.TanggalPeresepan, req.JamPeresepan, statusLanjut, "disimpan"); err != nil {
		return nil, err
	}

	if err := s.validasiKeberadaanObat(ctx, req); err != nil {
		return nil, err
	}

	if err := s.validasiMetodeRacik(ctx, req); err != nil {
		return nil, err
	}

	resep, err := s.repo.SimpanResep(ctx, kodeDokter, statusLanjut, req)
	if err != nil {
		s.log.Error("Gagal menyimpan resep obat no_rawat '%s' (%s) oleh dokter '%s': %v", req.NoRawat, statusLanjut, kodeDokter, err)
		return nil, err
	}

	s.log.Info("Berhasil menyimpan resep obat no_resep '%s' untuk no_rawat '%s' (%s %s, %s) oleh dokter '%s'", resep.NoResep, req.NoRawat, req.TanggalPeresepan, req.JamPeresepan, statusLanjut, kodeDokter)
	return resep, nil
}

func (s *service) HapusResep(ctx context.Context, kodeDokter, noRawat, noResep string, statusLanjut shared.StatusLanjut) error {
	resep, err := s.DetailResep(ctx, noResep)
	if err != nil {
		return err
	}

	if err := s.validasiAksesDanStatusResep(resep, noRawat, kodeDokter, "dihapus"); err != nil {
		return err
	}

	orderStatus := statusLanjut
	if orderStatus == "" {
		if strings.EqualFold(resep.Status, string(shared.StatusLanjutRawatInap)) {
			orderStatus = shared.StatusLanjutRawatInap
		} else {
			orderStatus = shared.StatusLanjutRawatJalan
		}
	}

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, "", "", orderStatus, "dihapus"); err != nil {
		return err
	}

	if err := s.repo.HapusResep(ctx, noResep); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError(
				"Resep obat tidak ditemukan",
				fmt.Sprintf("Data resep no_resep '%s' tidak ditemukan", noResep),
			)
		}
		s.log.Error("Gagal menghapus resep obat no_resep '%s' untuk no_rawat '%s' oleh dokter '%s': %v", noResep, noRawat, kodeDokter, err)
		return err
	}

	s.log.Info("Berhasil menghapus resep obat no_resep '%s' untuk no_rawat '%s' (%s) oleh dokter '%s'", noResep, noRawat, orderStatus, kodeDokter)
	return nil
}

func (s *service) UpdateResep(ctx context.Context, kodeDokter, noRawat, noResep string, statusLanjut shared.StatusLanjut, req SimpanResepRequest) (*Resep, error) {
	resep, err := s.DetailResep(ctx, noResep)
	if err != nil {
		return nil, err
	}

	if err := s.validasiAksesDanStatusResep(resep, noRawat, kodeDokter, "diubah"); err != nil {
		return nil, err
	}

	if err := s.validasiRegistrasiDanStatus(ctx, req.NoRawat, req.TanggalPeresepan, req.JamPeresepan, statusLanjut, "diubah"); err != nil {
		return nil, err
	}

	if err := s.validasiKeberadaanObat(ctx, req); err != nil {
		return nil, err
	}

	if err := s.validasiMetodeRacik(ctx, req); err != nil {
		return nil, err
	}

	updatedResep, err := s.repo.UpdateResep(ctx, noResep, req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError(
				"Resep obat tidak ditemukan",
				fmt.Sprintf("Data resep no_resep '%s' tidak ditemukan", noResep),
			)
		}
		s.log.Error("Gagal memperbarui resep obat no_resep '%s' untuk no_rawat '%s' (%s) oleh dokter '%s': %v", noResep, noRawat, statusLanjut, kodeDokter, err)
		return nil, err
	}

	s.log.Info("Berhasil memperbarui resep obat no_resep '%s' untuk no_rawat '%s' (%s %s, %s) oleh dokter '%s'", updatedResep.NoResep, noRawat, req.TanggalPeresepan, req.JamPeresepan, statusLanjut, kodeDokter)
	return updatedResep, nil
}

func (s *service) validasiAksesDanStatusResep(resep *Resep, noRawat, kodeDokter, action string) error {
	if resep.NoRawat != noRawat {
		return apperror.NewNotFoundError(
			"Resep obat tidak ditemukan",
			fmt.Sprintf("Data resep no_resep '%s' tidak ditemukan untuk no_rawat '%s'", resep.NoResep, noRawat),
		)
	}

	if resep.KodeDokter != kodeDokter {
		s.log.Warn("Percobaan %s resep no_resep %s oleh dokter %s ditolak: diresepkan oleh %s (%s)", action, resep.NoResep, kodeDokter, resep.KodeDokter, resep.NamaDokter)
		return apperror.NewForbiddenError(fmt.Sprintf("Resep obat dokter lain tidak dapat %s", action))
	}

	if resep.TanggalPerawatan != "0000-00-00" && resep.JamPerawatan != "00:00:00" {
		s.log.Warn("Percobaan %s resep no_resep %s ditolak: telah divalidasi farmasi pada %s %s", action, resep.NoResep, resep.TanggalPerawatan, resep.JamPerawatan)
		return apperror.NewForbiddenError(fmt.Sprintf("Resep obat telah divalidasi farmasi, tidak dapat %s", action))
	}

	if resep.TanggalPenyerahan != "0000-00-00" && resep.JamPenyerahan != "00:00:00" {
		s.log.Warn("Percobaan %s resep no_resep %s ditolak: telah diserahkan ke pasien pada %s %s", action, resep.NoResep, resep.TanggalPenyerahan, resep.JamPenyerahan)
		return apperror.NewForbiddenError(fmt.Sprintf("Resep obat telah diserahkan ke pasien, tidak dapat %s", action))
	}

	return nil
}

func (s *service) validasiKeberadaanObat(ctx context.Context, req SimpanResepRequest) error {

	listKode := make([]string, 0)
	for _, rd := range req.ResepDokter {
		if rd.KodeObat != "" {
			listKode = append(listKode, rd.KodeObat)
		}
	}
	for _, rr := range req.ResepRacikan {
		for _, d := range rr.Detail {
			if d.KodeObat != "" {
				listKode = append(listKode, d.KodeObat)
			}
		}
	}

	if len(listKode) == 0 {
		return nil
	}

	foundMap, err := s.obatService.CekKeberadaanObat(ctx, listKode)
	if err != nil {
		s.log.Error("Gagal memeriksa keberadaan obat: %v", err)
		return err
	}

	valErrs := make(apperror.ValidationError)
	for i, rd := range req.ResepDokter {
		if !foundMap[rd.KodeObat] {
			valErrs[fmt.Sprintf("resep_dokter[%d].id_obat", i)] = "Obat tidak ditemukan"
		}
	}

	for i, rr := range req.ResepRacikan {
		for j, d := range rr.Detail {
			if !foundMap[d.KodeObat] {
				valErrs[fmt.Sprintf("resep_racikan[%d].detail[%d].id_obat", i, j)] = fmt.Sprintf("Bahan racikan ke-%d: Obat tidak ditemukan", j+1)
			}
		}
	}

	if len(valErrs) > 0 {
		s.log.Warn("Validasi keberadaan obat gagal untuk no_rawat %s: %+v", req.NoRawat, valErrs)
		return valErrs
	}

	return nil
}

func (s *service) validasiMetodeRacik(ctx context.Context, req SimpanResepRequest) error {
	listKodeRacik := make([]string, 0)
	for _, rr := range req.ResepRacikan {
		if rr.KodeRacik != "" {
			listKodeRacik = append(listKodeRacik, rr.KodeRacik)
		}
	}

	if len(listKodeRacik) == 0 {
		return nil
	}

	foundMap, err := s.repo.CekKeberadaanMetodeRacik(ctx, listKodeRacik)
	if err != nil {
		s.log.Error("Gagal memeriksa keberadaan metode racik: %v", err)
		return err
	}

	valErrs := make(apperror.ValidationError)
	for i, rr := range req.ResepRacikan {
		if !foundMap[rr.KodeRacik] {
			valErrs[fmt.Sprintf("resep_racikan[%d].kode_racik", i)] = "Metode racik tidak ditemukan"
		}
	}

	if len(valErrs) > 0 {
		s.log.Warn("Validasi metode racik gagal untuk no_rawat %s: %+v", req.NoRawat, valErrs)
		return valErrs
	}

	return nil
}

func (s *service) validasiRegistrasiDanStatus(ctx context.Context, noRawat, tanggalPeresepan, jamPeresepan string, statusLanjut shared.StatusLanjut, aksi string) error {
	infoRegistrasi, err := s.rawatJalanService.GetInfoRegistrasi(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal mengambil data registrasi no_rawat '%s': %v", noRawat, err)
		return err
	}

	if infoRegistrasi.StatusBayar == "Sudah Bayar" && infoRegistrasi.KodePenjamin == "BPJ" {
		return apperror.NewBusinessError(fmt.Sprintf("Pasien BPJS sudah bayar, resep tidak dapat %s", aksi))
	}

	tanggalRegistrasi := infoRegistrasi.TanggalRegistrasi
	jamRegistrasi := infoRegistrasi.JamRegistrasi

	waktuRegistrasi, err := shared.ParseWaktu(tanggalRegistrasi, jamRegistrasi)
	if err != nil {
		s.log.Error("Gagal parse waktu registrasi no_rawat '%s' (%s %s): %v", noRawat, tanggalRegistrasi, jamRegistrasi, err)
		return err
	}

	if tanggalPeresepan != "" && jamPeresepan != "" {
		waktuPeresepan, err := shared.ParseWaktu(tanggalPeresepan, jamPeresepan)
		if err != nil {
			return apperror.NewBusinessError(err.Error())
		}

		if waktuPeresepan.Before(waktuRegistrasi) {
			return apperror.ValidationError{
				"tanggal_peresepan": fmt.Sprintf("Waktu peresepan (%s %s) tidak boleh sebelum waktu registrasi (%s %s)", tanggalPeresepan, jamPeresepan, tanggalRegistrasi, jamRegistrasi),
			}
		}
	}

	isAktifRanap, hasRecordKamar, err := s.rawatInapService.CekStatusKamarInap(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal cek status kamar inap untuk no_rawat '%s': %v", noRawat, err)
		return err
	}

	if hasRecordKamar {
		if !isAktifRanap {
			return apperror.NewBusinessError("Pasien sudah keluar dari kamar inap")
		}
		return nil
	}

	if strings.EqualFold(string(statusLanjut), string(shared.StatusLanjutRawatInap)) {
		return apperror.NewBusinessError("Pasien belum terdaftar di kamar inap, gunakan status 'ralan'")
	}

	batasWaktu := waktuRegistrasi.Add(time.Duration(s.maxEditJam) * time.Hour)
	if time.Now().After(batasWaktu) {
		return apperror.NewForbiddenError(
			fmt.Sprintf("Resep obat tidak dapat %s, melewati batas %d jam", aksi, s.maxEditJam),
			fmt.Sprintf("Resep obat no_rawat '%s' ditolak untuk %s karena melewati batas %d jam dari registrasi (%s)", noRawat, aksi, s.maxEditJam, waktuRegistrasi.Format("2006-01-02 15:04:05")),
		)
	}

	return nil
}

func (s *service) DaftarAturanPakai(ctx context.Context, keyword string) ([]AturanPakai, error) {
	data, err := s.repo.DaftarAturanPakai(ctx, keyword)
	if err != nil {
		s.log.Error("Gagal mengambil daftar aturan pakai: %v", err)
		return nil, err
	}

	return data, nil
}

func (s *service) DaftarMetodeRacik(ctx context.Context) ([]MetodeRacik, error) {
	data, err := s.repo.DaftarMetodeRacik(ctx)
	if err != nil {
		s.log.Error("Gagal mengambil daftar metode racik: %v", err)
		return nil, err
	}

	return data, nil
}
