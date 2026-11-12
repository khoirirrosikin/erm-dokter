package laboratorium

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

func (s *service) SimpanPermintaanLabPA(ctx context.Context, kodeDokterLogin string, statusLanjut shared.StatusLanjut, req SimpanPermintaanLabPARequest) (*DetailPermintaanLabPA, error) {
	noRawat := req.NoRawat

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, req.TanggalPermintaan, req.JamPermintaan, statusLanjut, "dibuat"); err != nil {
		return nil, err
	}

	kodeTindakanList, err := s.validasiTindakanLabPA(ctx, noRawat, req)
	if err != nil {
		return nil, err
	}

	noPermintaan, err := s.repo.SimpanPermintaanLabPA(ctx, noRawat, kodeDokterLogin, statusLanjut, req, kodeTindakanList)
	if err != nil {
		s.log.Error("Gagal menyimpan permintaan laboratorium PA %s: %v", noRawat, err)
		return nil, err
	}

	detail, err := s.GetDetailPermintaanLabPA(ctx, noPermintaan)
	if err != nil {
		return nil, err
	}

	s.log.Info("Berhasil membuat permintaan laboratorium PA %s untuk no_rawat %s (%s %s, %s) oleh dokter %s", noPermintaan, noRawat, req.TanggalPermintaan, req.JamPermintaan, statusLanjut, kodeDokterLogin)
	return detail, nil
}

func (s *service) UpdatePermintaanLabPA(ctx context.Context, kodeDokterLogin, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req SimpanPermintaanLabPARequest) (*DetailPermintaanLabPA, error) {
	detail, err := s.GetDetailPermintaanLabPA(ctx, noPermintaan)
	if err != nil {
		return nil, err
	}

	if err := s.validasiAksesDanStatusPermintaanLab(detail.PermintaanLabHeader, noRawat, kodeDokterLogin, statusLanjut, "diubah"); err != nil {
		return nil, err
	}

	if err := s.validasiStatusBayarPA(detail.NoPermintaan, "diubah", detail.Pemeriksaan); err != nil {
		return nil, err
	}

	if err := s.validasiRegistrasiDanStatus(ctx, req.NoRawat, req.TanggalPermintaan, req.JamPermintaan, statusLanjut, "diubah"); err != nil {
		return nil, err
	}

	kodeTindakanList, err := s.validasiTindakanLabPA(ctx, noRawat, req)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdatePermintaanLabPA(ctx, noPermintaan, req, kodeTindakanList); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data permintaan laboratorium tidak ditemukan")
		}
		s.log.Error("Gagal memperbarui permintaan lab PA %s: %v", noPermintaan, err)
		return nil, err
	}

	updatedDetail, err := s.GetDetailPermintaanLabPA(ctx, noPermintaan)
	if err != nil {
		return nil, err
	}

	s.log.Info("Berhasil memperbarui permintaan lab PA %s untuk no_rawat %s (%s %s, %s) oleh dokter %s", noPermintaan, noRawat, req.TanggalPermintaan, req.JamPermintaan, statusLanjut, kodeDokterLogin)
	return updatedDetail, nil
}

func (s *service) validasiTindakanLabPA(ctx context.Context, noRawat string, req SimpanPermintaanLabPARequest) ([]string, error) {
	var kodeTindakanList []string
	seenTindakan := make(map[string]bool)

	for _, item := range req.Pemeriksaan {
		if !seenTindakan[item.KodeTindakan] {
			seenTindakan[item.KodeTindakan] = true
			kodeTindakanList = append(kodeTindakanList, item.KodeTindakan)
		}
	}

	foundTindakan, err := s.tindakanService.CekKeberadaanTindakanLab(ctx, shared.KategoriLabPA, kodeTindakanList)
	if err != nil {
		s.log.Error("Gagal memeriksa keberadaan tindakan lab PA: %v", err)
		return nil, err
	}

	valErrs := make(apperror.ValidationError)
	for i, item := range req.Pemeriksaan {
		if !foundTindakan[item.KodeTindakan] {
			valErrs[fmt.Sprintf("pemeriksaan[%d].id_tindakan", i)] = fmt.Sprintf("Pemeriksaan ke-%d: Data master tindakan laboratorium tidak ditemukan", i+1)
		}
	}
	if len(valErrs) > 0 {
		s.log.Warn("Validasi keberadaan tindakan lab PA gagal untuk no_rawat %s: %+v", noRawat, valErrs)
		return nil, valErrs
	}

	return kodeTindakanList, nil
}

func (s *service) GetDaftarPermintaanLabPA(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]PermintaanLabPA, error) {
	items, err := s.repo.DaftarPermintaanLabPA(ctx, noRawat, statusLanjut)
	if err != nil {
		s.log.Error("Gagal mengambil riwayat permintaan lab PA %s: %v", noRawat, err)
		return nil, err
	}

	result := make([]PermintaanLabPA, 0, len(items))
	for _, item := range items {
		result = append(result, s.formatPermintaanLabPA(item))
	}

	return result, nil
}

func (s *service) GetRiwayatPermintaanLabPAByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterRiwayatLab) ([]PermintaanLabPA, shared.PaginationMeta, error) {
	items, total, err := s.repo.DaftarPermintaanLabPAByRM(ctx, noRM, statusLanjut, filter)
	if err != nil {
		s.log.Error("Gagal mengambil riwayat permintaan lab PA by RM %s: %v", noRM, err)
		return nil, shared.PaginationMeta{}, err
	}

	result := make([]PermintaanLabPA, 0, len(items))
	for _, item := range items {
		result = append(result, s.formatPermintaanLabPA(item))
	}

	meta := shared.NewPaginationMeta(total, filter.Page, filter.Limit)
	return result, meta, nil
}

func (s *service) GetDetailPermintaanLabPA(ctx context.Context, noPermintaan string) (*DetailPermintaanLabPA, error) {
	detail, err := s.repo.DetailPermintaanLabPA(ctx, noPermintaan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data permintaan laboratorium tidak ditemukan")
		}
		s.log.Error("Gagal mengambil detail permintaan lab PA %s: %v", noPermintaan, err)
		return nil, err
	}

	detail.PermintaanLabPA = s.formatPermintaanLabPA(detail.PermintaanLabPA)
	return detail, nil
}

func (s *service) HapusPermintaanLabPA(ctx context.Context, noRawat string, noPermintaan string, statusLanjut shared.StatusLanjut, kodeDokterLogin string) error {
	detail, err := s.GetDetailPermintaanLabPA(ctx, noPermintaan)
	if err != nil {
		return err
	}

	if err := s.validasiAksesDanStatusPermintaanLab(detail.PermintaanLabHeader, noRawat, kodeDokterLogin, statusLanjut, "dihapus"); err != nil {
		return err
	}

	if err := s.validasiStatusBayarPA(detail.NoPermintaan, "dihapus", detail.Pemeriksaan); err != nil {
		return err
	}

	orderStatus := statusLanjut
	if orderStatus == "" || orderStatus == "Semua" {
		if strings.EqualFold(detail.Status, string(shared.StatusLanjutRawatInap)) {
			orderStatus = shared.StatusLanjutRawatInap
		} else {
			orderStatus = shared.StatusLanjutRawatJalan
		}
	}

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, "", "", orderStatus, "dihapus"); err != nil {
		return err
	}

	if err := s.repo.HapusPermintaanLabPA(ctx, noPermintaan); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError("Data permintaan laboratorium tidak ditemukan")
		}
		s.log.Error("Gagal menghapus permintaan lab PA di repository %s: %v", noPermintaan, err)
		return err
	}

	s.log.Info("Berhasil menghapus permintaan lab PA %s untuk no_rawat %s (%s) oleh dokter %s", noPermintaan, noRawat, orderStatus, kodeDokterLogin)
	return nil
}

func (s *service) formatPermintaanLabPA(item PermintaanLabPA) PermintaanLabPA {
	item.FormatZeroDates()
	if item.PengambilanBahan == "0000-00-00" {
		item.PengambilanBahan = ""
	}
	if item.TanggalPASebelumnya == "0000-00-00" {
		item.TanggalPASebelumnya = ""
	}
	return item
}
