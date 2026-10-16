package resep_test

import (
	"testing"
	"time"

	"erm-dokter/internal/resep"
)

func TestSimpanResepRequest_Sanitize(t *testing.T) {
	req := resep.SimpanResepRequest{
		NoRawat: " 2026/08/28/000001 ",
		ResepDokter: []resep.ResepDokterInput{
			{
				ItemObatInput: resep.ItemObatInput{
					IdObat: "  OBAT123  ",
					Jumlah: 10,
				},
				AturanPakai: "  3x1 sehari  ",
			},
		},
		ResepRacikan: []resep.ResepRacikanInput{
			{
				NamaRacik:     "  Puyer Flu  ",
				KodeRacik:     "  R01  ",
				JumlahRacikan: 10,
				AturanPakai:   "  3x1  ",
				Keterangan:    "  Sebelum tidur  ",
				Detail: []resep.ResepRacikanDetailInput{
					{
						ItemObatInput: resep.ItemObatInput{
							IdObat: "  OBAT456  ",
							Jumlah: 5,
						},
						Kandungan: "  500mg  ",
					},
				},
			},
		},
	}

	req.Sanitize()

	if req.NoRawat != "2026/08/28/000001" {
		t.Errorf("expected trimmed no_rawat, got %q", req.NoRawat)
	}
	if req.TanggalPeresepan == "" {
		t.Error("expected default tanggal_peresepan to be set")
	}
	if req.JamPeresepan == "" {
		t.Error("expected default jam_peresepan to be set")
	}
	if req.ResepDokter[0].IdObat != "OBAT123" {
		t.Errorf("expected trimmed id_obat, got %q", req.ResepDokter[0].IdObat)
	}
	if req.ResepDokter[0].AturanPakai != "3x1 sehari" {
		t.Errorf("expected trimmed aturan_pakai, got %q", req.ResepDokter[0].AturanPakai)
	}
	if req.ResepRacikan[0].NamaRacik != "Puyer Flu" {
		t.Errorf("expected trimmed nama_racik, got %q", req.ResepRacikan[0].NamaRacik)
	}
	if req.ResepRacikan[0].KodeRacik != "R01" {
		t.Errorf("expected trimmed kode_racik, got %q", req.ResepRacikan[0].KodeRacik)
	}
	if req.ResepRacikan[0].Detail[0].IdObat != "OBAT456" {
		t.Errorf("expected trimmed detail id_obat, got %q", req.ResepRacikan[0].Detail[0].IdObat)
	}
}

func TestSimpanResepRequest_Validate_ValidNonRacikan(t *testing.T) {
	pastTime := time.Now().Add(-5 * time.Minute)
	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: pastTime.Format("2006-01-02"),
		JamPeresepan:     pastTime.Format("15:04:05"),
		ResepDokter: []resep.ResepDokterInput{
			{
				ItemObatInput: resep.ItemObatInput{
					IdObat: "encrypted-obat-1",
					Jumlah: 10,
				},
				AturanPakai: "3 x 1 tablet",
			},
		},
	}

	errs := req.Validate()
	if errs != nil {
		t.Fatalf("expected no errors, got %+v", errs)
	}
}

func TestSimpanResepRequest_Validate_ValidRacikan(t *testing.T) {
	pastTime := time.Now().Add(-5 * time.Minute)
	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: pastTime.Format("2006-01-02"),
		JamPeresepan:     pastTime.Format("15:04:05"),
		ResepRacikan: []resep.ResepRacikanInput{
			{
				NamaRacik:     "Puyer Batuk",
				KodeRacik:     "R01",
				JumlahRacikan: 10,
				AturanPakai:   "3 x 1 bungkus",
				Keterangan:    "Sesudah makan",
				Detail: []resep.ResepRacikanDetailInput{
					{
						ItemObatInput: resep.ItemObatInput{
							IdObat: "encrypted-obat-1",
							Jumlah: 5,
						},
						Kandungan: "500 mg",
					},
					{
						ItemObatInput: resep.ItemObatInput{
							IdObat: "encrypted-obat-2",
							Jumlah: 5,
						},
						Kandungan: "4 mg",
					},
				},
			},
		},
	}

	errs := req.Validate()
	if errs != nil {
		t.Fatalf("expected no errors, got %+v", errs)
	}
}

func TestSimpanResepRequest_Validate_EmptyFields(t *testing.T) {
	req := resep.SimpanResepRequest{}
	errs := req.Validate()

	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}

	if _, ok := errs["no_rawat"]; !ok {
		t.Error("expected error for no_rawat")
	}
	if _, ok := errs["resep"]; !ok {
		t.Error("expected error for resep (empty list)")
	}
}

func TestSimpanResepRequest_Validate_InvalidDateTime(t *testing.T) {
	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: "2026-13-45",
		JamPeresepan:     "25:70:99",
		ResepDokter: []resep.ResepDokterInput{
			{
				ItemObatInput: resep.ItemObatInput{
					IdObat: "obat-1",
					Jumlah: 5,
				},
				AturanPakai: "1x1",
			},
		},
	}

	errs := req.Validate()
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}

	if _, ok := errs["tanggal_peresepan"]; !ok {
		t.Error("expected error for tanggal_peresepan")
	}
	if _, ok := errs["jam_peresepan"]; !ok {
		t.Error("expected error for jam_peresepan")
	}
}

func TestSimpanResepRequest_Validate_FutureTime(t *testing.T) {
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	req := resep.SimpanResepRequest{
		NoRawat:          "2026/08/28/000001",
		TanggalPeresepan: tomorrow,
		JamPeresepan:     "12:00:00",
		ResepDokter: []resep.ResepDokterInput{
			{
				ItemObatInput: resep.ItemObatInput{
					IdObat: "obat-1",
					Jumlah: 5,
				},
				AturanPakai: "1x1",
			},
		},
	}

	errs := req.Validate()
	if errs == nil {
		t.Fatal("expected validation error for future time, got nil")
	}

	if _, ok := errs["tanggal_peresepan"]; !ok {
		t.Error("expected error for future tanggal_peresepan")
	}
}

func TestSimpanResepRequest_Validate_InvalidResepDokter(t *testing.T) {
	req := resep.SimpanResepRequest{
		NoRawat: "2026/08/28/000001",
		ResepDokter: []resep.ResepDokterInput{
			{
				ItemObatInput: resep.ItemObatInput{
					IdObat: "",
					Jumlah: 0,
				},
				AturanPakai: "",
			},
		},
	}

	errs := req.Validate()
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}

	if _, ok := errs["resep_dokter[0].id_obat"]; !ok {
		t.Error("expected error for resep_dokter[0].id_obat")
	}
	if _, ok := errs["resep_dokter[0].jumlah"]; !ok {
		t.Error("expected error for resep_dokter[0].jumlah")
	}
	if _, ok := errs["resep_dokter[0].aturan_pakai"]; !ok {
		t.Error("expected error for resep_dokter[0].aturan_pakai")
	}
}

func TestSimpanResepRequest_Validate_InvalidRacikan(t *testing.T) {
	req := resep.SimpanResepRequest{
		NoRawat: "2026/08/28/000001",
		ResepRacikan: []resep.ResepRacikanInput{
			{
				NamaRacik:     "",
				KodeRacik:     "",
				JumlahRacikan: 0,
				AturanPakai:   "",
				Detail:        []resep.ResepRacikanDetailInput{},
			},
		},
	}

	errs := req.Validate()
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}

	if _, ok := errs["resep_racikan[0].nama_racik"]; !ok {
		t.Error("expected error for nama_racik")
	}
	if _, ok := errs["resep_racikan[0].kode_racik"]; !ok {
		t.Error("expected error for kode_racik")
	}
	if _, ok := errs["resep_racikan[0].jumlah_racikan"]; !ok {
		t.Error("expected error for jumlah_racikan")
	}
	if _, ok := errs["resep_racikan[0].aturan_pakai"]; !ok {
		t.Error("expected error for aturan_pakai")
	}
	if _, ok := errs["resep_racikan[0].detail"]; !ok {
		t.Error("expected error for detail")
	}
}

func TestSimpanResepRequest_Validate_InvalidRacikanDetail(t *testing.T) {
	req := resep.SimpanResepRequest{
		NoRawat: "2026/08/28/000001",
		ResepRacikan: []resep.ResepRacikanInput{
			{
				NamaRacik:     "Puyer",
				KodeRacik:     "R01",
				JumlahRacikan: 10,
				AturanPakai:   "3x1",
				Detail: []resep.ResepRacikanDetailInput{
					{
						ItemObatInput: resep.ItemObatInput{
							IdObat: "",
							Jumlah: -1,
						},
					},
				},
			},
		},
	}

	errs := req.Validate()
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}

	if _, ok := errs["resep_racikan[0].detail[0].id_obat"]; !ok {
		t.Error("expected error for detail id_obat")
	}
	if _, ok := errs["resep_racikan[0].detail[0].jumlah"]; !ok {
		t.Error("expected error for detail jumlah")
	}
}
