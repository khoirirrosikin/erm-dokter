package penilaianmedis

import (
	"context"
	"database/sql"
	"errors"

	"erm-dokter/internal/pkg/database"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

func (s *service) DetailPenilaianMedisRanap(ctx context.Context, noRawat string) (*PenilaianMedisRanap, error) {
	item, err := s.repo.DetailPenilaianMedisRanap(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data penilaian awal medis rawat inap tidak ditemukan")
		}
		s.log.Error("Gagal mengambil detail penilaian medis ranap no_rawat %s: %v", noRawat, err)
		return nil, err
	}

	return item, nil
}

func (s *service) RiwayatPenilaianMedisRanapByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRanap, error) {
	list, err := s.repo.RiwayatPenilaianMedisRanapByNoRM(ctx, noRM)
	if err != nil {
		s.log.Error("Gagal mengambil riwayat penilaian medis ranap no_rkm_medis %s: %v", noRM, err)
		return nil, err
	}

	return list, nil
}

func (s *service) SimpanPenilaianMedisRanap(ctx context.Context, kodeDokterLogin, noRawat string, req SimpanPenilaianMedisRanapRequest) (*PenilaianMedisRanap, error) {
	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, req.TanggalPenilaian, shared.StatusLanjutRawatInap, "disimpan"); err != nil {
		return nil, err
	}

	ada, err := s.repo.CekPenilaianMedisRanapAda(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal memeriksa keberadaan penilaian medis ranap no_rawat %s: %v", noRawat, err)
		return nil, err
	}
	if ada {
		s.log.Warn("Percobaan duplikasi penilaian medis ranap pada no_rawat %s", noRawat)
		return nil, apperror.NewBusinessError("Penilaian awal medis rawat inap untuk kunjungan ini sudah ada")
	}

	if err := s.repo.SimpanPenilaianMedisRanap(ctx, noRawat, kodeDokterLogin, req); err != nil {
		if database.IsDuplicateEntry(err) {
			s.log.Warn("Percobaan duplikasi penilaian medis ranap pada no_rawat %s", noRawat)
			return nil, apperror.NewBusinessError("Penilaian awal medis rawat inap untuk kunjungan ini sudah ada")
		}
		s.log.Error("Gagal menyimpan penilaian medis ranap no_rawat %s oleh dokter %s: %v", noRawat, kodeDokterLogin, err)
		return nil, err
	}

	s.log.Info("Berhasil menyimpan penilaian medis ranap no_rawat %s oleh dokter %s", noRawat, kodeDokterLogin)
	return s.DetailPenilaianMedisRanap(ctx, noRawat)
}

func (s *service) UpdatePenilaianMedisRanap(ctx context.Context, kodeDokterLogin, noRawat string, req UpdatePenilaianMedisRanapRequest) (*PenilaianMedisRanap, error) {
	existing, err := s.repo.DetailPenilaianMedisRanap(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data penilaian awal medis rawat inap tidak ditemukan")
		}
		s.log.Error("Gagal memeriksa data penilaian medis ranap sebelum update no_rawat %s: %v", noRawat, err)
		return nil, err
	}

	if err := s.validasiKepemilikanDokter(existing.KodeDokter, existing.NamaDokter, kodeDokterLogin, noRawat, "diubah"); err != nil {
		return nil, err
	}

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, req.TanggalPenilaian, shared.StatusLanjutRawatInap, "diubah"); err != nil {
		return nil, err
	}

	if err := s.repo.UpdatePenilaianMedisRanap(ctx, noRawat, req); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data penilaian awal medis rawat inap tidak ditemukan")
		}
		s.log.Error("Gagal memperbarui penilaian medis ranap no_rawat %s oleh dokter %s: %v", noRawat, kodeDokterLogin, err)
		return nil, err
	}

	s.log.Info("Berhasil memperbarui penilaian medis ranap no_rawat %s oleh dokter %s", noRawat, kodeDokterLogin)
	return s.DetailPenilaianMedisRanap(ctx, noRawat)
}

func (s *service) HapusPenilaianMedisRanap(ctx context.Context, kodeDokterLogin, noRawat string) error {
	existing, err := s.repo.DetailPenilaianMedisRanap(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError("Data penilaian awal medis rawat inap tidak ditemukan")
		}
		s.log.Error("Gagal memeriksa data penilaian medis ranap sebelum hapus no_rawat %s: %v", noRawat, err)
		return err
	}

	if err := s.validasiKepemilikanDokter(existing.KodeDokter, existing.NamaDokter, kodeDokterLogin, noRawat, "dihapus"); err != nil {
		return err
	}

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, "", shared.StatusLanjutRawatInap, "dihapus"); err != nil {
		return err
	}

	if err := s.repo.HapusPenilaianMedisRanap(ctx, noRawat); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError("Data penilaian awal medis rawat inap tidak ditemukan")
		}
		s.log.Error("Gagal menghapus penilaian medis ranap no_rawat %s oleh dokter %s: %v", noRawat, kodeDokterLogin, err)
		return err
	}

	s.log.Info("Berhasil menghapus penilaian medis ranap no_rawat %s oleh dokter %s", noRawat, kodeDokterLogin)
	return nil
}
