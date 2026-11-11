package penilaianmedis

import (
	"context"
	"database/sql"
	"errors"

	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

func (s *service) DetailPenilaianMedisRanapKandungan(ctx context.Context, noRawat string) (*PenilaianMedisRanapKandungan, error) {
	item, err := s.repo.DetailPenilaianMedisRanapKandungan(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data penilaian awal medis rawat inap kandungan tidak ditemukan")
		}
		s.log.Error("Gagal mengambil detail penilaian medis ranap kandungan no_rawat %s: %v", noRawat, err)
		return nil, err
	}

	return item, nil
}

func (s *service) RiwayatPenilaianMedisRanapKandunganByNoRM(ctx context.Context, noRM string) ([]PenilaianMedisRanapKandungan, error) {
	list, err := s.repo.RiwayatPenilaianMedisRanapKandunganByNoRM(ctx, noRM)
	if err != nil {
		s.log.Error("Gagal mengambil riwayat penilaian medis ranap kandungan no_rkm_medis %s: %v", noRM, err)
		return nil, err
	}

	return list, nil
}

func (s *service) SimpanPenilaianMedisRanapKandungan(ctx context.Context, kodeDokterLogin, noRawat string, req SimpanPenilaianMedisRanapKandunganRequest) (*PenilaianMedisRanapKandungan, error) {
	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, req.TanggalPenilaian, shared.StatusLanjutRawatInap, "disimpan"); err != nil {
		return nil, err
	}

	ada, err := s.repo.CekPenilaianMedisRanapKandunganAda(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal memeriksa keberadaan penilaian medis ranap kandungan no_rawat %s: %v", noRawat, err)
		return nil, err
	}
	if ada {
		s.log.Warn("Percobaan duplikasi penilaian medis ranap kandungan pada no_rawat %s", noRawat)
		return nil, apperror.NewBusinessError("Penilaian awal medis rawat inap kandungan untuk kunjungan ini sudah ada")
	}

	if err := s.repo.SimpanPenilaianMedisRanapKandungan(ctx, noRawat, kodeDokterLogin, req); err != nil {
		s.log.Error("Gagal menyimpan penilaian medis ranap kandungan no_rawat %s oleh dokter %s: %v", noRawat, kodeDokterLogin, err)
		return nil, err
	}

	s.log.Info("Berhasil menyimpan penilaian medis ranap kandungan no_rawat %s oleh dokter %s", noRawat, kodeDokterLogin)
	return s.DetailPenilaianMedisRanapKandungan(ctx, noRawat)
}

func (s *service) UpdatePenilaianMedisRanapKandungan(ctx context.Context, kodeDokterLogin, noRawat string, req UpdatePenilaianMedisRanapKandunganRequest) (*PenilaianMedisRanapKandungan, error) {
	existing, err := s.repo.DetailPenilaianMedisRanapKandungan(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data penilaian awal medis rawat inap kandungan tidak ditemukan")
		}
		s.log.Error("Gagal memeriksa data penilaian medis ranap kandungan sebelum update no_rawat %s: %v", noRawat, err)
		return nil, err
	}

	if err := s.validasiKepemilikanDokter(existing.KodeDokter, existing.NamaDokter, kodeDokterLogin, noRawat, "diubah"); err != nil {
		return nil, err
	}

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, req.TanggalPenilaian, shared.StatusLanjutRawatInap, "diubah"); err != nil {
		return nil, err
	}

	if err := s.repo.UpdatePenilaianMedisRanapKandungan(ctx, noRawat, req); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data penilaian awal medis rawat inap kandungan tidak ditemukan")
		}
		s.log.Error("Gagal memperbarui penilaian medis ranap kandungan no_rawat %s oleh dokter %s: %v", noRawat, kodeDokterLogin, err)
		return nil, err
	}

	s.log.Info("Berhasil memperbarui penilaian medis ranap kandungan no_rawat %s oleh dokter %s", noRawat, kodeDokterLogin)
	return s.DetailPenilaianMedisRanapKandungan(ctx, noRawat)
}

func (s *service) HapusPenilaianMedisRanapKandungan(ctx context.Context, kodeDokterLogin, noRawat string) error {
	existing, err := s.repo.DetailPenilaianMedisRanapKandungan(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError("Data penilaian awal medis rawat inap kandungan tidak ditemukan")
		}
		s.log.Error("Gagal memeriksa data penilaian medis ranap kandungan sebelum hapus no_rawat %s: %v", noRawat, err)
		return err
	}

	if err := s.validasiKepemilikanDokter(existing.KodeDokter, existing.NamaDokter, kodeDokterLogin, noRawat, "dihapus"); err != nil {
		return err
	}

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, "", shared.StatusLanjutRawatInap, "dihapus"); err != nil {
		return err
	}

	if err := s.repo.HapusPenilaianMedisRanapKandungan(ctx, noRawat); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError("Data penilaian awal medis rawat inap kandungan tidak ditemukan")
		}
		s.log.Error("Gagal menghapus penilaian medis ranap kandungan no_rawat %s oleh dokter %s: %v", noRawat, kodeDokterLogin, err)
		return err
	}

	s.log.Info("Berhasil menghapus penilaian medis ranap kandungan no_rawat %s oleh dokter %s", noRawat, kodeDokterLogin)
	return nil
}
