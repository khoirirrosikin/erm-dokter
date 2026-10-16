package pemeriksaan_test

import (
	"testing"
	"time"

	"erm-dokter/internal/pemeriksaan"
)

func TestPemeriksaan_CompositeKey(t *testing.T) {
	id := pemeriksaan.IdPemeriksaan{
		NoRawat:            "2026/09/03/000001",
		TanggalPemeriksaan: "2026-09-03",
		JamPemeriksaan:     "10:00:00",
	}

	key := id.CompositeKey()
	if key != "2026/09/03/000001~2026-09-03~10:00:00" {
		t.Errorf("Unexpected composite key: %s", key)
	}

	parsed, err := pemeriksaan.ParseIdPemeriksaan(key)
	if err != nil {
		t.Fatalf("Failed to parse composite key: %v", err)
	}
	if parsed.NoRawat != id.NoRawat || parsed.TanggalPemeriksaan != id.TanggalPemeriksaan || parsed.JamPemeriksaan != id.JamPemeriksaan {
		t.Errorf("Parsed mismatch: %+v", parsed)
	}

	invalidKeys := []string{"invalid~format", "~~", "2026/01/01~~", "~~10:00:00", ""}
	for _, k := range invalidKeys {
		if _, err := pemeriksaan.ParseIdPemeriksaan(k); err == nil {
			t.Errorf("Expected error for invalid composite key %q", k)
		}
	}
}

func TestSimpanPemeriksaanRequest_ValidationAndSanitization(t *testing.T) {
	now := time.Now()
	nowDate := now.Format("2006-01-02")
	nowTime := now.Format("15:04:05")

	t.Run("Valid Request", func(t *testing.T) {
		req := pemeriksaan.SimpanPemeriksaanRequest{
			NoRawat: "2026/09/03/000001",
			DataPemeriksaan: pemeriksaan.DataPemeriksaan{
				TanggalPemeriksaan:  nowDate,
				JamPemeriksaan:      nowTime,
				SuhuTubuh:           "36.5",
				Tensi:               "120/80",
				Nadi:                "80",
				Respirasi:           "20",
				TinggiBadan:         "170",
				BeratBadan:          "65",
				SpO2:                "98",
				Gcs:                 "E4V5M6",
				Kesadaran:           pemeriksaan.KesadaranComposMentis,
				Keluhan:             "Demam sejak 2 hari",
				Pemeriksaan:         "Kepala dbn",
				Penilaian:           "Febris",
				Instruksi:           "Paracetamol 3x500mg",
				RencanaTindakLanjut: "Kontrol 3 hari",
				Evaluasi:            "Membaik",
			},
		}
		req.Sanitize()
		if errs := req.Validate(); errs != nil {
			t.Errorf("Expected valid request, got errs: %+v", errs)
		}
	})

	t.Run("Invalid Future Date Time", func(t *testing.T) {
		future := time.Now().Add(24 * time.Hour)
		req := pemeriksaan.SimpanPemeriksaanRequest{
			NoRawat: "2026/09/03/000001",
			DataPemeriksaan: pemeriksaan.DataPemeriksaan{
				TanggalPemeriksaan: future.Format("2006-01-02"),
				JamPemeriksaan:     future.Format("15:04:05"),
				Kesadaran:          pemeriksaan.KesadaranComposMentis,
				Keluhan:            "Keluhan",
				Pemeriksaan:        "Pemeriksaan",
				Penilaian:          "Penilaian",
				Instruksi:          "Instruksi",
			},
		}
		req.Sanitize()
		errs := req.Validate()
		if errs == nil {
			t.Fatal("Expected validation error for future time")
		}
	})

	t.Run("Invalid Empty Required Fields", func(t *testing.T) {
		req := pemeriksaan.SimpanPemeriksaanRequest{}
		req.Sanitize()
		errs := req.Validate()
		if errs == nil {
			t.Fatal("Expected validation errors for empty fields")
		}
	})

	t.Run("Sanitize HH:mm Format to HH:mm:ss", func(t *testing.T) {
		req := pemeriksaan.SimpanPemeriksaanRequest{
			NoRawat: "2026/09/03/000001",
			DataPemeriksaan: pemeriksaan.DataPemeriksaan{
				TanggalPemeriksaan:  "2026-09-03",
				JamPemeriksaan:      "08:30",
				Kesadaran:           pemeriksaan.KesadaranComposMentis,
				Keluhan:             "Keluhan",
				Pemeriksaan:         "Pemeriksaan",
				Penilaian:           "Penilaian",
				Instruksi:           "Instruksi",
				RencanaTindakLanjut: "RTL",
				Evaluasi:            "Evaluasi",
			},
		}
		req.Sanitize()
		if req.JamPemeriksaan != "08:30:00" {
			t.Errorf("Expected sanitized JamPemeriksaan '08:30:00', got '%s'", req.JamPemeriksaan)
		}
		if errs := req.Validate(); errs != nil {
			t.Errorf("Expected valid request after sanitization, got errs: %+v", errs)
		}
	})
}
