package resumepasien

import (
	"context"
	"database/sql"
	"errors"

	"erm-dokter/internal/pkg/database"
	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

func (s *service) DetailResumePasienRalan(ctx context.Context, noRawat string) (*ResumePasienRalan, error) {
	item, err := s.repo.DetailResumePasienRalan(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data resume pasien rawat jalan tidak ditemukan")
		}
		s.log.Error("Gagal mengambil detail resume pasien ralan no_rawat %s: %v", noRawat, err)
		return nil, err
	}
	return item, nil
}

func (s *service) RiwayatResumePasienRalanByNoRM(ctx context.Context, noRM string) ([]ResumePasienRalan, error) {
	list, err := s.repo.RiwayatResumePasienRalanByNoRM(ctx, noRM)
	if err != nil {
		s.log.Error("Gagal mengambil riwayat resume pasien ralan no_rkm_medis %s: %v", noRM, err)
		return nil, err
	}
	return list, nil
}

func (s *service) SimpanResumePasienRalan(ctx context.Context, noRawat, kodeDokter string, req SimpanResumePasienRalanRequest) (*ResumePasienRalan, error) {
	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, shared.StatusLanjutRawatJalan, "disimpan"); err != nil {
		return nil, err
	}

	ada, err := s.repo.CekResumePasienRalanAda(ctx, noRawat)
	if err != nil {
		s.log.Error("Gagal cek keberadaan resume pasien ralan no_rawat %s: %v", noRawat, err)
		return nil, err
	}
	if ada {
		return nil, apperror.NewBusinessError("Resume medis rawat jalan untuk kunjungan ini sudah ada")
	}

	item, err := s.repo.SimpanResumePasienRalan(ctx, noRawat, kodeDokter, req)
	if err != nil {
		if database.IsDuplicateEntry(err) {
			return nil, apperror.NewBusinessError("Resume medis rawat jalan untuk kunjungan ini sudah ada")
		}
		s.log.Error("Gagal menyimpan resume pasien ralan no_rawat %s: %v", noRawat, err)
		return nil, err
	}

	s.log.Info("Berhasil menyimpan resume pasien ralan no_rawat %s oleh dokter %s", noRawat, kodeDokter)
	return item, nil
}

func (s *service) UpdateResumePasienRalan(ctx context.Context, kodeDokterLogin, noRawat string, req UpdateResumePasienRalanRequest) (*ResumePasienRalan, error) {
	existing, err := s.repo.DetailResumePasienRalan(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data resume pasien rawat jalan tidak ditemukan")
		}
		s.log.Error("Gagal mengambil resume pasien ralan no_rawat %s untuk update: %v", noRawat, err)
		return nil, err
	}

	if err := s.validasiKepemilikanDokter(existing.KodeDokter, existing.NamaDokter, kodeDokterLogin, noRawat, "mengubah"); err != nil {
		return nil, err
	}

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, shared.StatusLanjutRawatJalan, "diubah"); err != nil {
		return nil, err
	}

	item, err := s.repo.UpdateResumePasienRalan(ctx, noRawat, req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NewNotFoundError("Data resume pasien rawat jalan tidak ditemukan")
		}
		s.log.Error("Gagal memperbarui resume pasien ralan no_rawat %s: %v", noRawat, err)
		return nil, err
	}

	s.log.Info("Berhasil memperbarui resume pasien ralan no_rawat %s oleh dokter %s", noRawat, kodeDokterLogin)
	return item, nil
}

func (s *service) HapusResumePasienRalan(ctx context.Context, kodeDokterLogin, noRawat string) error {
	existing, err := s.repo.DetailResumePasienRalan(ctx, noRawat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError("Data resume pasien rawat jalan tidak ditemukan")
		}
		s.log.Error("Gagal mengambil resume pasien ralan no_rawat %s untuk hapus: %v", noRawat, err)
		return err
	}

	if err := s.validasiKepemilikanDokter(existing.KodeDokter, existing.NamaDokter, kodeDokterLogin, noRawat, "menghapus"); err != nil {
		return err
	}

	if err := s.validasiRegistrasiDanStatus(ctx, noRawat, shared.StatusLanjutRawatJalan, "dihapus"); err != nil {
		return err
	}

	if err := s.repo.HapusResumePasienRalan(ctx, noRawat); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NewNotFoundError("Data resume pasien rawat jalan tidak ditemukan")
		}
		s.log.Error("Gagal menghapus resume pasien ralan no_rawat %s: %v", noRawat, err)
		return err
	}

	s.log.Info("Berhasil menghapus resume pasien ralan no_rawat %s oleh dokter %s", noRawat, kodeDokterLogin)
	return nil
}

func (s *service) ReferensiRalan(ctx context.Context) ReferensiResumeRalan {
	return ReferensiResumeRalan{
		KeadaanPulang: DaftarOpsiKeadaanPulangRalan(),
	}
}
