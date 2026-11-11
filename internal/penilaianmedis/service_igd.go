package penilaianmedis

import (
	"context"
	"database/sql"
	"errors"

	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

func (s *service) DetailPenilaianMedisIGD(ctx context.Context, noRawat string) (*PenilaianMedisIGD, error) {
	item, err := s.repo.DetailPenilaianMedisIGD(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data penilaian awal medis IGD tidak ditemukan")
		}
		s.log.Error("Gagal mengambil detail penilaian medis IGD no_rawat %s: %v", noRawat, err)
		return nil, err
	}

	return item, nil
}

func (s *service) RiwayatPenilaianMedisIGDByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisIGD, error) {
	list, err := s.repo.RiwayatPenilaianMedisIGDByNoRM(ctx, noRM)
	if err != nil {
		s.log.Error("Gagal mengambil riwayat penilaian medis IGD no_rkm_medis %s: %v", noRM, err)
		return nil, err
	}

	return list, nil
}

func (s *service) SimpanPenilaianMedisIGD(ctx context.Context, kodeDokterLogin, noRawat string, req SimpanPenilaianMedisIGDRequest) (*PenilaianMedisIGD, error) {
	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, req.TanggalPenilaian, shared.StatusLanjutRawatJalan, "disimpan"); err != nil {
		return nil, err
	}

	ada, err := s.repo.CekPenilaianMedisIGDAda(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal memeriksa keberadaan penilaian medis IGD no_rawat %s: %v", noRawat, err)
		return nil, err
	}
	if ada {
		s.log.Warn("Percobaan duplikasi penilaian medis IGD pada no_rawat %s", noRawat)
		return nil, apperror.NewBusinessError("Penilaian awal medis IGD untuk kunjungan ini sudah ada")
	}

	if err := s.repo.SimpanPenilaianMedisIGD(ctx, noRawat, kodeDokterLogin, req); err != nil {
		s.log.Error("Gagal menyimpan penilaian medis IGD no_rawat %s oleh dokter %s: %v", noRawat, kodeDokterLogin, err)
		return nil, err
	}

	s.log.Info("Berhasil menyimpan penilaian medis IGD no_rawat %s oleh dokter %s", noRawat, kodeDokterLogin)
	return s.DetailPenilaianMedisIGD(ctx, noRawat)
}

func (s *service) UpdatePenilaianMedisIGD(ctx context.Context, kodeDokterLogin, noRawat string, req UpdatePenilaianMedisIGDRequest) (*PenilaianMedisIGD, error) {
	existing, err := s.repo.DetailPenilaianMedisIGD(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data penilaian awal medis IGD tidak ditemukan")
		}
		s.log.Error("Gagal memeriksa data penilaian medis IGD sebelum update no_rawat %s: %v", noRawat, err)
		return nil, err
	}

	if err := s.validasiKepemilikanDokter(existing.KodeDokter, existing.NamaDokter, kodeDokterLogin, noRawat, "diubah"); err != nil {
		return nil, err
	}

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, req.TanggalPenilaian, shared.StatusLanjutRawatJalan, "diubah"); err != nil {
		return nil, err
	}

	if err := s.repo.UpdatePenilaianMedisIGD(ctx, noRawat, req); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data penilaian awal medis IGD tidak ditemukan")
		}
		s.log.Error("Gagal memperbarui penilaian medis IGD no_rawat %s oleh dokter %s: %v", noRawat, kodeDokterLogin, err)
		return nil, err
	}

	s.log.Info("Berhasil memperbarui penilaian medis IGD no_rawat %s oleh dokter %s", noRawat, kodeDokterLogin)
	return s.DetailPenilaianMedisIGD(ctx, noRawat)
}

func (s *service) HapusPenilaianMedisIGD(ctx context.Context, kodeDokterLogin, noRawat string) error {
	existing, err := s.repo.DetailPenilaianMedisIGD(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError("Data penilaian awal medis IGD tidak ditemukan")
		}
		s.log.Error("Gagal memeriksa data penilaian medis IGD sebelum hapus no_rawat %s: %v", noRawat, err)
		return err
	}

	if err := s.validasiKepemilikanDokter(existing.KodeDokter, existing.NamaDokter, kodeDokterLogin, noRawat, "dihapus"); err != nil {
		return err
	}

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, "", shared.StatusLanjutRawatJalan, "dihapus"); err != nil {
		return err
	}

	if err := s.repo.HapusPenilaianMedisIGD(ctx, noRawat); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError("Data penilaian awal medis IGD tidak ditemukan")
		}
		s.log.Error("Gagal menghapus penilaian medis IGD no_rawat %s oleh dokter %s: %v", noRawat, kodeDokterLogin, err)
		return err
	}

	s.log.Info("Berhasil menghapus penilaian medis IGD no_rawat %s oleh dokter %s", noRawat, kodeDokterLogin)
	return nil
}
