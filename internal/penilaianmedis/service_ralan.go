package penilaianmedis

import (
	"context"
	"database/sql"
	"errors"

	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

func (s *service) DetailPenilaianMedisRalan(ctx context.Context, noRawat string) (*PenilaianMedisRalan, error) {
	item, err := s.repo.DetailPenilaianMedisRalan(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data penilaian awal medis rawat jalan tidak ditemukan")
		}
		s.log.Error("Gagal mengambil detail penilaian medis ralan no_rawat %s: %v", noRawat, err)
		return nil, err
	}

	return item, nil
}

func (s *service) RiwayatPenilaianMedisRalanByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRalan, error) {
	list, err := s.repo.RiwayatPenilaianMedisRalanByNoRM(ctx, noRM)
	if err != nil {
		s.log.Error("Gagal mengambil riwayat penilaian medis ralan no_rkm_medis %s: %v", noRM, err)
		return nil, err
	}

	return list, nil
}

func (s *service) SimpanPenilaianMedisRalan(ctx context.Context, kodeDokterLogin, noRawat string, req SimpanPenilaianMedisRalanRequest) (*PenilaianMedisRalan, error) {
	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, req.TanggalPenilaian, shared.StatusLanjutRawatJalan, "disimpan"); err != nil {
		return nil, err
	}

	ada, err := s.repo.CekPenilaianMedisRalanAda(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal memeriksa keberadaan penilaian medis ralan no_rawat %s: %v", noRawat, err)
		return nil, err
	}
	if ada {
		s.log.Warn("Percobaan duplikasi penilaian medis ralan pada no_rawat %s", noRawat)
		return nil, apperror.NewBusinessError("Penilaian awal medis rawat jalan untuk kunjungan ini sudah ada")
	}

	if err := s.repo.SimpanPenilaianMedisRalan(ctx, noRawat, kodeDokterLogin, req); err != nil {
		s.log.Error("Gagal menyimpan penilaian medis ralan no_rawat %s oleh dokter %s: %v", noRawat, kodeDokterLogin, err)
		return nil, err
	}

	s.log.Info("Berhasil menyimpan penilaian medis ralan no_rawat %s oleh dokter %s", noRawat, kodeDokterLogin)
	return s.DetailPenilaianMedisRalan(ctx, noRawat)
}

func (s *service) UpdatePenilaianMedisRalan(ctx context.Context, kodeDokterLogin, noRawat string, req UpdatePenilaianMedisRalanRequest) (*PenilaianMedisRalan, error) {
	existing, err := s.repo.DetailPenilaianMedisRalan(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data penilaian awal medis rawat jalan tidak ditemukan")
		}
		s.log.Error("Gagal memeriksa data penilaian medis ralan sebelum update no_rawat %s: %v", noRawat, err)
		return nil, err
	}

	if err := s.validasiKepemilikanDokter(existing.KodeDokter, existing.NamaDokter, kodeDokterLogin, noRawat, "diubah"); err != nil {
		return nil, err
	}

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, req.TanggalPenilaian, shared.StatusLanjutRawatJalan, "diubah"); err != nil {
		return nil, err
	}

	if err := s.repo.UpdatePenilaianMedisRalan(ctx, noRawat, req); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data penilaian awal medis rawat jalan tidak ditemukan")
		}
		s.log.Error("Gagal memperbarui penilaian medis ralan no_rawat %s oleh dokter %s: %v", noRawat, kodeDokterLogin, err)
		return nil, err
	}

	s.log.Info("Berhasil memperbarui penilaian medis ralan no_rawat %s oleh dokter %s", noRawat, kodeDokterLogin)
	return s.DetailPenilaianMedisRalan(ctx, noRawat)
}

func (s *service) HapusPenilaianMedisRalan(ctx context.Context, kodeDokterLogin, noRawat string) error {
	existing, err := s.repo.DetailPenilaianMedisRalan(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError("Data penilaian awal medis rawat jalan tidak ditemukan")
		}
		s.log.Error("Gagal memeriksa data penilaian medis ralan sebelum hapus no_rawat %s: %v", noRawat, err)
		return err
	}

	if err := s.validasiKepemilikanDokter(existing.KodeDokter, existing.NamaDokter, kodeDokterLogin, noRawat, "dihapus"); err != nil {
		return err
	}

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, "", shared.StatusLanjutRawatJalan, "dihapus"); err != nil {
		return err
	}

	if err := s.repo.HapusPenilaianMedisRalan(ctx, noRawat); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError("Data penilaian awal medis rawat jalan tidak ditemukan")
		}
		s.log.Error("Gagal menghapus penilaian medis ralan no_rawat %s oleh dokter %s: %v", noRawat, kodeDokterLogin, err)
		return err
	}

	s.log.Info("Berhasil menghapus penilaian medis ralan no_rawat %s oleh dokter %s", noRawat, kodeDokterLogin)
	return nil
}
