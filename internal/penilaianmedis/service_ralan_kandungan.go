package penilaianmedis

import (
	"context"
	"database/sql"
	"errors"

	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

func (s *service) DetailPenilaianMedisRalanKandungan(ctx context.Context, noRawat string) (*PenilaianMedisRalanKandungan, error) {
	item, err := s.repo.DetailPenilaianMedisRalanKandungan(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data penilaian awal medis rawat jalan kandungan tidak ditemukan")
		}
		s.log.Error("Gagal mengambil detail penilaian medis ralan kandungan no_rawat %s: %v", noRawat, err)
		return nil, err
	}

	return item, nil
}

func (s *service) RiwayatPenilaianMedisRalanKandunganByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRalanKandungan, error) {
	list, err := s.repo.RiwayatPenilaianMedisRalanKandunganByNoRM(ctx, noRM)
	if err != nil {
		s.log.Error("Gagal mengambil riwayat penilaian medis ralan kandungan no_rkm_medis %s: %v", noRM, err)
		return nil, err
	}

	return list, nil
}

func (s *service) SimpanPenilaianMedisRalanKandungan(ctx context.Context, kodeDokterLogin, noRawat string, req SimpanPenilaianMedisRalanKandunganRequest) (*PenilaianMedisRalanKandungan, error) {
	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, req.TanggalPenilaian, shared.StatusLanjutRawatJalan, "disimpan"); err != nil {
		return nil, err
	}

	ada, err := s.repo.CekPenilaianMedisRalanKandunganAda(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal memeriksa keberadaan penilaian medis ralan kandungan no_rawat %s: %v", noRawat, err)
		return nil, err
	}
	if ada {
		s.log.Warn("Percobaan duplikasi penilaian medis ralan kandungan pada no_rawat %s", noRawat)
		return nil, apperror.NewBusinessError("Penilaian awal medis rawat jalan kandungan untuk kunjungan ini sudah ada")
	}

	if err := s.repo.SimpanPenilaianMedisRalanKandungan(ctx, noRawat, kodeDokterLogin, req); err != nil {
		s.log.Error("Gagal menyimpan penilaian medis ralan kandungan no_rawat %s oleh dokter %s: %v", noRawat, kodeDokterLogin, err)
		return nil, err
	}

	s.log.Info("Berhasil menyimpan penilaian medis ralan kandungan no_rawat %s oleh dokter %s", noRawat, kodeDokterLogin)
	return s.DetailPenilaianMedisRalanKandungan(ctx, noRawat)
}

func (s *service) UpdatePenilaianMedisRalanKandungan(ctx context.Context, kodeDokterLogin, noRawat string, req UpdatePenilaianMedisRalanKandunganRequest) (*PenilaianMedisRalanKandungan, error) {
	existing, err := s.repo.DetailPenilaianMedisRalanKandungan(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data penilaian awal medis rawat jalan kandungan tidak ditemukan")
		}
		s.log.Error("Gagal memeriksa data penilaian medis ralan kandungan sebelum update no_rawat %s: %v", noRawat, err)
		return nil, err
	}

	if err := s.validasiKepemilikanDokter(existing.KodeDokter, existing.NamaDokter, kodeDokterLogin, noRawat, "diubah"); err != nil {
		return nil, err
	}

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, req.TanggalPenilaian, shared.StatusLanjutRawatJalan, "diubah"); err != nil {
		return nil, err
	}

	if err := s.repo.UpdatePenilaianMedisRalanKandungan(ctx, noRawat, req); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data penilaian awal medis rawat jalan kandungan tidak ditemukan")
		}
		s.log.Error("Gagal memperbarui penilaian medis ralan kandungan no_rawat %s oleh dokter %s: %v", noRawat, kodeDokterLogin, err)
		return nil, err
	}

	s.log.Info("Berhasil memperbarui penilaian medis ralan kandungan no_rawat %s oleh dokter %s", noRawat, kodeDokterLogin)
	return s.DetailPenilaianMedisRalanKandungan(ctx, noRawat)
}

func (s *service) HapusPenilaianMedisRalanKandungan(ctx context.Context, kodeDokterLogin, noRawat string) error {
	existing, err := s.repo.DetailPenilaianMedisRalanKandungan(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError("Data penilaian awal medis rawat jalan kandungan tidak ditemukan")
		}
		s.log.Error("Gagal memeriksa data penilaian medis ralan kandungan sebelum hapus no_rawat %s: %v", noRawat, err)
		return err
	}

	if err := s.validasiKepemilikanDokter(existing.KodeDokter, existing.NamaDokter, kodeDokterLogin, noRawat, "dihapus"); err != nil {
		return err
	}

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, "", shared.StatusLanjutRawatJalan, "dihapus"); err != nil {
		return err
	}

	if err := s.repo.HapusPenilaianMedisRalanKandungan(ctx, noRawat); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError("Data penilaian awal medis rawat jalan kandungan tidak ditemukan")
		}
		s.log.Error("Gagal menghapus penilaian medis ralan kandungan no_rawat %s oleh dokter %s: %v", noRawat, kodeDokterLogin, err)
		return err
	}

	s.log.Info("Berhasil menghapus penilaian medis ralan kandungan no_rawat %s oleh dokter %s", noRawat, kodeDokterLogin)
	return nil
}
