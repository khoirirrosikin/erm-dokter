package pasien_test

import (
	"testing"
	"time"

	"erm-dokter/internal/pasien"
)

func TestPasienModelFormatters(t *testing.T) {
	birthDate := time.Now().AddDate(-25, -2, -10).Format("2006-01-02")

	p := pasien.Pasien{
		NoRM:         "00123456",
		Nama:         "Budi Santoso",
		JenisKelamin: "L",
		TanggalLahir: birthDate,
	}

	if got := p.FormatNoRekamMedis(); got != "123456" {
		t.Errorf("Expected FormatNoRekamMedis '123456', got '%s'", got)
	}

	if got := p.FormatJenisKelamin(); got != "Laki-Laki" {
		t.Errorf("Expected FormatJenisKelamin 'Laki-Laki', got '%s'", got)
	}

	p2 := pasien.Pasien{
		NoRM:         "654321",
		Nama:         "Siti Rahma",
		JenisKelamin: "P",
		TanggalLahir: birthDate,
	}

	if got := p2.FormatNoRekamMedis(); got != "654321" {
		t.Errorf("Expected FormatNoRekamMedis '654321', got '%s'", got)
	}

	if got := p2.FormatJenisKelamin(); got != "Perempuan" {
		t.Errorf("Expected FormatJenisKelamin 'Perempuan', got '%s'", got)
	}

	umur := p.HitungUmurLengkap()
	if umur.Tahun != 25 {
		t.Errorf("Expected age in years to be 25, got %d", umur.Tahun)
	}

	umurStr := p.FormatUmur()
	if umurStr == "" {
		t.Errorf("Expected non-empty format umur")
	}
}

func TestFilterRiwayatKunjungan_SanitizeAndValidate(t *testing.T) {
	t.Run("Sanitize default values and limits", func(t *testing.T) {
		f := pasien.FilterRiwayatKunjungan{
			Tanggal: "  2026-01-01,2026-01-31  ",
			Page:    0,
			Limit:   0,
		}
		f.Sanitize()

		if f.Tanggal != "2026-01-01,2026-01-31" {
			t.Errorf("expected trimmed Tanggal, got %s", f.Tanggal)
		}
		if f.Page != 1 {
			t.Errorf("expected default page 1, got %d", f.Page)
		}
		if f.Limit != 3 {
			t.Errorf("expected default limit 3, got %d", f.Limit)
		}

		fMax := pasien.FilterRiwayatKunjungan{Limit: 100}
		fMax.Sanitize()
		if fMax.Limit != 50 {
			t.Errorf("expected capped limit 50, got %d", fMax.Limit)
		}

		if f.Offset() != 0 {
			t.Errorf("expected offset 0, got %d", f.Offset())
		}
	})

	t.Run("Validation rejects invalid tanggal", func(t *testing.T) {
		for _, bad := range []string{"not-a-date", "abc,xyz", "2026-02-30,2026-03-01", "2026-12-31,2026-01-01"} {
			f := pasien.FilterRiwayatKunjungan{
				Tanggal: bad,
			}
			f.Sanitize()
			errs := f.Validate()
			if errs == nil {
				t.Fatalf("expected validation error for invalid tanggal %q", bad)
			}
		}
	})

	t.Run("Validation passes with valid parameters", func(t *testing.T) {
		f := pasien.FilterRiwayatKunjungan{
			Tanggal: "2026-01-01,2026-01-31",
		}
		f.Sanitize()
		errs := f.Validate()
		if errs != nil {
			t.Fatalf("unexpected validation error: %v", errs)
		}
	})

	t.Run("Validation passes with empty filter", func(t *testing.T) {
		f := pasien.FilterRiwayatKunjungan{}
		f.Sanitize()
		errs := f.Validate()
		if errs != nil {
			t.Fatalf("unexpected validation error: %v", errs)
		}
	})
}
