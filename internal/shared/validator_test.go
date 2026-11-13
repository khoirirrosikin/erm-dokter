package shared_test

import (
	"testing"

	"erm-dokter/internal/shared"
	"erm-dokter/internal/shared/apperror"
)

func TestParseWaktu(t *testing.T) {
	tests := []struct {
		name        string
		tanggal     string
		jam         string
		expectError bool
		checkHour   int
		checkMin    int
		checkSec    int
	}{
		{
			name:        "Tanggal dan jam lengkap (HH:mm:ss)",
			tanggal:     "2026-08-21",
			jam:         "14:30:45",
			expectError: false,
			checkHour:   14,
			checkMin:    30,
			checkSec:    45,
		},
		{
			name:        "Tanggal dan jam tanpa detik (HH:mm)",
			tanggal:     "2026-08-21",
			jam:         "09:15",
			expectError: false,
			checkHour:   9,
			checkMin:    15,
			checkSec:    0,
		},
		{
			name:        "Tanggal saja tanpa jam (jam kosong)",
			tanggal:     "2026-08-21",
			jam:         "",
			expectError: false,
			checkHour:   0,
			checkMin:    0,
			checkSec:    0,
		},
		{
			name:        "Tanggal kosong",
			tanggal:     "",
			jam:         "12:00:00",
			expectError: true,
		},
		{
			name:        "Format tanggal salah",
			tanggal:     "21-08-2026",
			jam:         "12:00:00",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := shared.ParseWaktu(tt.tanggal, tt.jam)
			if (err != nil) != tt.expectError {
				t.Fatalf("expected error: %v, got: %v", tt.expectError, err)
			}
			if !tt.expectError {
				if res.Hour() != tt.checkHour || res.Minute() != tt.checkMin || res.Second() != tt.checkSec {
					t.Errorf("expected time %02d:%02d:%02d, got %02d:%02d:%02d",
						tt.checkHour, tt.checkMin, tt.checkSec, res.Hour(), res.Minute(), res.Second())
				}
			}
		})
	}
}

func TestValidasiRentangTanggal(t *testing.T) {
	errs := make(apperror.ValidationError)
	shared.ValidasiRentangTanggal("2026-01-01,2026-01-31", errs)
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %+v", errs)
	}

	errsInvalid := make(apperror.ValidationError)
	shared.ValidasiRentangTanggal("2026-01-31,2026-01-01", errsInvalid)
	if _, exists := errsInvalid["tanggal"]; !exists {
		t.Error("expected validation error when tglAwal > tglAkhir")
	}
}

func TestIsPenjaminTerkunciSudahBayar(t *testing.T) {
	if !shared.IsPenjaminTerkunciSudahBayar("Sudah Bayar", "BPJ", nil) {
		t.Error("expected default nil list to lock BPJ when Sudah Bayar")
	}
	if shared.IsPenjaminTerkunciSudahBayar("Sudah Bayar", "UMU", nil) {
		t.Error("expected default nil list not to lock UMU")
	}
	if shared.IsPenjaminTerkunciSudahBayar("Belum Bayar", "BPJ", []string{"BPJ", "UMU"}) {
		t.Error("expected Belum Bayar never to lock")
	}
	if !shared.IsPenjaminTerkunciSudahBayar("Sudah Bayar", "umu", []string{"BPJ", "UMU"}) {
		t.Error("expected configured UMU to lock case-insensitively")
	}
	if !shared.IsPenjaminTerkunciSudahBayar("Sudah Bayar", "A09", []string{"*"}) {
		t.Error("expected wildcard * to lock any penjamin when Sudah Bayar")
	}
	if shared.IsPenjaminTerkunciSudahBayar("Sudah Bayar", "BPJ", []string{"-"}) {
		t.Error("expected '-' to disable Sudah Bayar lock")
	}

	if got := shared.PesanTerkunciSudahBayar("BPJ", "resep", "disimpan"); got != "Pasien BPJS sudah bayar, resep tidak dapat disimpan" {
		t.Errorf("unexpected BPJ message: %s", got)
	}
	if got := shared.PesanTerkunciSudahBayar("UMU", "resep", "disimpan"); got != "Tagihan kunjungan sudah dibayar, resep tidak dapat disimpan" {
		t.Errorf("unexpected non-BPJ message: %s", got)
	}
}

func TestIsNonZeroDateAndTime(t *testing.T) {
	zeroDates := []string{"", "   ", "0000-00-00", " 0000-00-00 "}
	for _, d := range zeroDates {
		if shared.IsNonZeroDate(d) {
			t.Errorf("expected IsNonZeroDate(%q) to be false", d)
		}
	}
	if !shared.IsNonZeroDate("2026-11-13") {
		t.Error("expected IsNonZeroDate(\"2026-11-13\") to be true")
	}

	zeroTimes := []string{"", "   ", "00:00:00", " 00:00:00 "}
	for _, tm := range zeroTimes {
		if shared.IsNonZeroTime(tm) {
			t.Errorf("expected IsNonZeroTime(%q) to be false", tm)
		}
	}
	if !shared.IsNonZeroTime("00:00:01") || !shared.IsNonZeroTime("14:30:00") {
		t.Error("expected valid non-zero times to return true")
	}
}
