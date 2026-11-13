package shared

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"erm-dokter/internal/shared/apperror"
)

func ValidasiRentangTanggal(tanggal string, errs apperror.ValidationError) {
	const layout = "2006-01-02"
	parts := strings.Split(tanggal, ",")

	if len(parts) != 2 {
		errs["tanggal"] = "Format tanggal harus rentang 2 tanggal (YYYY-MM-DD,YYYY-MM-DD)"
		return
	}

	tglAwal, err := time.Parse(layout, strings.TrimSpace(parts[0]))
	if err != nil {
		errs["tanggal"] = "Format tanggal awal tidak valid (YYYY-MM-DD)"
		return
	}

	tglAkhir, err := time.Parse(layout, strings.TrimSpace(parts[1]))
	if err != nil {
		errs["tanggal"] = "Format tanggal akhir tidak valid (YYYY-MM-DD)"
		return
	}

	if tglAwal.After(tglAkhir) {
		errs["tanggal"] = "Tanggal awal harus ≤ tanggal akhir"
	}
}

func ParseWaktu(tanggal, jam string) (time.Time, error) {
	tanggal = strings.TrimSpace(tanggal)
	jam = strings.TrimSpace(jam)
	if tanggal == "" {
		return time.Time{}, errors.New("Tanggal tidak boleh kosong")
	}

	tgl, err := time.Parse("2006-01-02", tanggal)
	if err != nil {
		return time.Time{}, fmt.Errorf("Format tanggal tidak valid (YYYY-MM-DD): %w", err)
	}

	hour, min, sec := 0, 0, 0
	if jam != "" {
		jamParts := strings.Split(jam, ":")
		if len(jamParts) >= 1 {
			hour, _ = strconv.Atoi(jamParts[0])
		}
		if len(jamParts) >= 2 {
			min, _ = strconv.Atoi(jamParts[1])
		}
		if len(jamParts) >= 3 {
			sec, _ = strconv.Atoi(jamParts[2])
		}
	}

	return time.Date(tgl.Year(), tgl.Month(), tgl.Day(), hour, min, sec, 0, time.Local), nil
}

func NormalizeKodePenjaminKunciBayar(daftarKode []string) []string {
	if len(daftarKode) == 0 {
		return []string{"BPJ"}
	}
	cleaned := make([]string, 0, len(daftarKode))
	for _, k := range daftarKode {
		if trimmed := strings.TrimSpace(k); trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	if len(cleaned) == 0 {
		return []string{"BPJ"}
	}
	return cleaned
}

func IsPenjaminTerkunciSudahBayar(statusBayar, kodePenjamin string, daftarKode []string) bool {
	if !strings.EqualFold(strings.TrimSpace(statusBayar), "Sudah Bayar") {
		return false
	}
	kp := strings.TrimSpace(kodePenjamin)
	if kp == "" {
		return false
	}
	for _, k := range NormalizeKodePenjaminKunciBayar(daftarKode) {
		if k == "-" || strings.EqualFold(k, "NONE") {
			return false
		}
		if k == "*" || strings.EqualFold(k, "ALL") || strings.EqualFold(k, kp) {
			return true
		}
	}
	return false
}

func PesanTerkunciSudahBayar(kodePenjamin, namaModul, aksi string) string {
	if strings.EqualFold(strings.TrimSpace(kodePenjamin), "BPJ") {
		return fmt.Sprintf("Pasien BPJS sudah bayar, %s tidak dapat %s", namaModul, aksi)
	}
	return fmt.Sprintf("Tagihan kunjungan sudah dibayar, %s tidak dapat %s", namaModul, aksi)
}

func IsNonZeroDate(s string) bool {
	trimmed := strings.TrimSpace(s)
	return trimmed != "" && trimmed != "0000-00-00"
}

func IsNonZeroTime(s string) bool {
	trimmed := strings.TrimSpace(s)
	return trimmed != "" && trimmed != "00:00:00"
}
