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

func (s *service) SimpanPermintaanLabMB(ctx context.Context, kodeDokterLogin string, statusLanjut shared.StatusLanjut, req SimpanPermintaanLabMBRequest) (*DetailPermintaanLabMB, error) {
	noRawat := req.NoRawat

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, req.TanggalPermintaan, req.JamPermintaan, statusLanjut, "dibuat"); err != nil {
		return nil, err
	}

	kodeTindakanList, templateMap, err := s.validasiTindakanDanTemplateLabMB(ctx, noRawat, req)
	if err != nil {
		return nil, err
	}

	noPermintaan, err := s.repo.SimpanPermintaanLabMB(ctx, noRawat, kodeDokterLogin, statusLanjut, req, kodeTindakanList, templateMap)
	if err != nil {
		s.log.Error("Gagal menyimpan permintaan laboratorium MB %s: %v", noRawat, err)
		return nil, err
	}

	detail, err := s.GetDetailPermintaanLabMB(ctx, noPermintaan)
	if err != nil {
		return nil, err
	}

	s.log.Info("Berhasil membuat permintaan laboratorium MB %s untuk no_rawat %s (%s %s, %s) oleh dokter %s", noPermintaan, noRawat, req.TanggalPermintaan, req.JamPermintaan, statusLanjut, kodeDokterLogin)
	return detail, nil
}

func (s *service) UpdatePermintaanLabMB(ctx context.Context, kodeDokterLogin, noRawat, noPermintaan string, statusLanjut shared.StatusLanjut, req SimpanPermintaanLabMBRequest) (*DetailPermintaanLabMB, error) {
	detail, err := s.GetDetailPermintaanLabMB(ctx, noPermintaan)
	if err != nil {
		return nil, err
	}

	if err := s.validasiAksesDanStatusPermintaanLab(detail.PermintaanLabHeader, noRawat, kodeDokterLogin, statusLanjut, "diubah"); err != nil {
		return nil, err
	}

	if err := s.validasiStatusBayarMB(detail.NoPermintaan, "diubah", detail.Pemeriksaan); err != nil {
		return nil, err
	}

	if err := s.validasiRegistrasiDanStatus(ctx, req.NoRawat, req.TanggalPermintaan, req.JamPermintaan, statusLanjut, "diubah"); err != nil {
		return nil, err
	}

	kodeTindakanList, templateMap, err := s.validasiTindakanDanTemplateLabMB(ctx, noRawat, req)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdatePermintaanLabMB(ctx, noPermintaan, req, kodeTindakanList, templateMap); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data permintaan laboratorium tidak ditemukan")
		}
		s.log.Error("Gagal memperbarui permintaan lab MB %s: %v", noPermintaan, err)
		return nil, err
	}

	updatedDetail, err := s.GetDetailPermintaanLabMB(ctx, noPermintaan)
	if err != nil {
		return nil, err
	}

	s.log.Info("Berhasil memperbarui permintaan lab MB %s untuk no_rawat %s (%s %s, %s) oleh dokter %s", noPermintaan, noRawat, req.TanggalPermintaan, req.JamPermintaan, statusLanjut, kodeDokterLogin)
	return updatedDetail, nil
}

func (s *service) validasiTindakanDanTemplateLabMB(ctx context.Context, noRawat string, req SimpanPermintaanLabMBRequest) ([]string, map[string][]int, error) {
	var kodeTindakanList []string
	templateMap := make(map[string][]int)
	seenTindakan := make(map[string]bool)

	for _, item := range req.Pemeriksaan {
		if !seenTindakan[item.KodeTindakan] {
			seenTindakan[item.KodeTindakan] = true
			kodeTindakanList = append(kodeTindakanList, item.KodeTindakan)
		}
		if len(item.KodeTemplate) > 0 {
			templateMap[item.KodeTindakan] = append(templateMap[item.KodeTindakan], item.KodeTemplate...)
		}
	}

	foundTindakan, err := s.tindakanService.CekKeberadaanTindakanLab(ctx, shared.KategoriLabMB, kodeTindakanList)
	if err != nil {
		s.log.Error("Gagal memeriksa keberadaan tindakan lab MB: %v", err)
		return nil, nil, err
	}

	valErrs := make(apperror.ValidationError)
	for i, item := range req.Pemeriksaan {
		if !foundTindakan[item.KodeTindakan] {
			valErrs[fmt.Sprintf("pemeriksaan[%d].id_tindakan", i)] = fmt.Sprintf("Pemeriksaan ke-%d: Data master tindakan laboratorium tidak ditemukan", i+1)
		}
	}
	if len(valErrs) > 0 {
		s.log.Warn("Validasi keberadaan tindakan lab MB gagal untuk no_rawat %s: %+v", noRawat, valErrs)
		return nil, nil, valErrs
	}

	foundTemplates, err := s.tindakanService.CekKeberadaanTemplateLab(ctx, kodeTindakanList, templateMap)
	if err != nil {
		s.log.Error("Gagal memeriksa keberadaan template lab MB: %v", err)
		return nil, nil, err
	}

	for i, item := range req.Pemeriksaan {
		prefix := fmt.Sprintf("pemeriksaan[%d]", i)
		tmplMapForTindakan := foundTemplates[item.KodeTindakan]
		for j, idTmpl := range item.KodeTemplate {
			if tmplMapForTindakan == nil || !tmplMapForTindakan[idTmpl] {
				valErrs[fmt.Sprintf("%s.id_template[%d]", prefix, j)] = fmt.Sprintf("Pemeriksaan ke-%d parameter ke-%d: Template pengujian tidak terdaftar pada tindakan ini", i+1, j+1)
			}
		}
	}
	if len(valErrs) > 0 {
		s.log.Warn("Validasi template lab MB gagal untuk no_rawat %s: %+v", noRawat, valErrs)
		return nil, nil, valErrs
	}

	return kodeTindakanList, templateMap, nil
}

func (s *service) GetDaftarPermintaanLabMB(ctx context.Context, noRawat string, statusLanjut shared.StatusLanjut) ([]PermintaanLabMB, error) {
	items, err := s.repo.DaftarPermintaanLabMB(ctx, noRawat, statusLanjut)
	if err != nil {
		s.log.Error("Gagal mengambil riwayat permintaan lab MB %s: %v", noRawat, err)
		return nil, err
	}

	result := make([]PermintaanLabMB, 0, len(items))
	for _, item := range items {
		result = append(result, s.formatPermintaanLabMB(item))
	}

	return result, nil
}

func (s *service) GetRiwayatPermintaanLabMBByRM(ctx context.Context, noRM string, statusLanjut shared.StatusLanjut, filter FilterRiwayatLab) ([]PermintaanLabMB, shared.PaginationMeta, error) {
	items, total, err := s.repo.DaftarPermintaanLabMBByRM(ctx, noRM, statusLanjut, filter)
	if err != nil {
		s.log.Error("Gagal mengambil riwayat permintaan lab MB by RM %s: %v", noRM, err)
		return nil, shared.PaginationMeta{}, err
	}

	result := make([]PermintaanLabMB, 0, len(items))
	for _, item := range items {
		result = append(result, s.formatPermintaanLabMB(item))
	}

	meta := shared.NewPaginationMeta(total, filter.Page, filter.Limit)
	return result, meta, nil
}

func (s *service) GetDetailPermintaanLabMB(ctx context.Context, noPermintaan string) (*DetailPermintaanLabMB, error) {
	detail, err := s.repo.DetailPermintaanLabMB(ctx, noPermintaan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data permintaan laboratorium tidak ditemukan")
		}
		s.log.Error("Gagal mengambil detail permintaan lab MB %s: %v", noPermintaan, err)
		return nil, err
	}

	detail.PermintaanLabMB = s.formatPermintaanLabMB(detail.PermintaanLabMB)
	return detail, nil
}

func (s *service) HapusPermintaanLabMB(ctx context.Context, noRawat string, noPermintaan string, statusLanjut shared.StatusLanjut, kodeDokterLogin string) error {
	detail, err := s.GetDetailPermintaanLabMB(ctx, noPermintaan)
	if err != nil {
		return err
	}

	if err := s.validasiAksesDanStatusPermintaanLab(detail.PermintaanLabHeader, noRawat, kodeDokterLogin, statusLanjut, "dihapus"); err != nil {
		return err
	}

	if err := s.validasiStatusBayarMB(detail.NoPermintaan, "dihapus", detail.Pemeriksaan); err != nil {
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

	if err := s.repo.HapusPermintaanLabMB(ctx, noPermintaan); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError("Data permintaan laboratorium tidak ditemukan")
		}
		s.log.Error("Gagal menghapus permintaan lab MB di repository %s: %v", noPermintaan, err)
		return err
	}

	s.log.Info("Berhasil menghapus permintaan lab MB %s untuk no_rawat %s (%s) oleh dokter %s", noPermintaan, noRawat, orderStatus, kodeDokterLogin)
	return nil
}

func (s *service) formatPermintaanLabMB(item PermintaanLabMB) PermintaanLabMB {
	item.FormatZeroDates()
	return item
}
