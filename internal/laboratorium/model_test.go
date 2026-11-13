package laboratorium_test

import (
	"testing"
	"time"

	"erm-dokter/internal/laboratorium"
)

func TestFilterRiwayatLab_Sanitize(t *testing.T) {
	filter := laboratorium.FilterRiwayatLab{
		Page:  -1,
		Limit: 0,
	}
	filter.Sanitize()

	if filter.Page != 1 {
		t.Errorf("Expected Page 1, got %d", filter.Page)
	}
	if filter.Limit != 5 {
		t.Errorf("Expected Limit 5, got %d", filter.Limit)
	}

	filterHigh := laboratorium.FilterRiwayatLab{
		Page:  2,
		Limit: 200,
	}
	filterHigh.Sanitize()
	if filterHigh.Limit != 100 {
		t.Errorf("Expected Limit 100, got %d", filterHigh.Limit)
	}
}

func TestFilterRiwayatLab_Offset(t *testing.T) {
	filter := laboratorium.FilterRiwayatLab{
		Page:  3,
		Limit: 10,
	}
	if filter.Offset() != 20 {
		t.Errorf("Expected Offset 20, got %d", filter.Offset())
	}
}

func TestFilterRiwayatLab_Validate(t *testing.T) {
	tests := []struct {
		name    string
		filter  laboratorium.FilterRiwayatLab
		wantErr bool
	}{
		{
			name: "Valid without date",
			filter: laboratorium.FilterRiwayatLab{
				Page:  1,
				Limit: 10,
			},
			wantErr: false,
		},
		{
			name: "Valid with same date range",
			filter: laboratorium.FilterRiwayatLab{
				Tanggal: "2026-04-22,2026-04-22",
				Page:    1,
				Limit:   10,
			},
			wantErr: false,
		},
		{
			name: "Valid with date range",
			filter: laboratorium.FilterRiwayatLab{
				Tanggal: "2026-04-01,2026-04-30",
				Page:    1,
				Limit:   10,
			},
			wantErr: false,
		},
		{
			name: "Invalid single date without comma",
			filter: laboratorium.FilterRiwayatLab{
				Tanggal: "2026-04-22",
				Page:    1,
				Limit:   10,
			},
			wantErr: true,
		},
		{
			name: "Invalid date format",
			filter: laboratorium.FilterRiwayatLab{
				Tanggal: "invalid-date",
				Page:    1,
				Limit:   10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.filter.Validate()
			if (errs != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestSimpanPermintaanLabPKRequest_ValidationAndSanitization(t *testing.T) {
	now := time.Now()
	nowStr := now.Format("2006-01-02")
	jamNowStr := now.Format("15:04:05")

	t.Run("Valid request with full fields", func(t *testing.T) {
		req := laboratorium.SimpanPermintaanLabPKRequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				NoRawat:           "2026/09/03/000001",
				TanggalPermintaan: nowStr,
				JamPermintaan:     jamNowStr,
				DiagnosaKlinis:    "Febris H-3",
				InformasiTambahan: "Cek rutin",
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabPKRequest{
				{
					IdTindakan: "enc-tindakan-1",
					IdTemplate: []string{"enc-tmpl-1", "enc-tmpl-2"},
				},
			},
		}

		errs := req.Validate()
		if errs != nil {
			t.Fatalf("Expected valid, got errors: %v", errs)
		}
		if req.InformasiTambahan != "Cek rutin" {
			t.Errorf("Expected informasi tambahan tetap, got %s", req.InformasiTambahan)
		}
	})

	t.Run("Valid request with HH:MM format and empty informasi tambahan defaults to hyphen", func(t *testing.T) {
		jamHHMM := now.Format("15:04")
		req := laboratorium.SimpanPermintaanLabPKRequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				NoRawat:           "2026/09/03/000001",
				TanggalPermintaan: nowStr,
				JamPermintaan:     jamHHMM,
				DiagnosaKlinis:    "Susp TBC",
				InformasiTambahan: "",
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabPKRequest{
				{
					IdTindakan: "enc-tindakan-2",
				},
			},
		}

		req.Sanitize()
		errs := req.Validate()
		if errs != nil {
			t.Fatalf("Expected valid, got errors: %v", errs)
		}
		if req.JamPermintaan != jamHHMM+":00" {
			t.Errorf("Expected jam %s, got %s", jamHHMM+":00", req.JamPermintaan)
		}
		if req.InformasiTambahan != "-" {
			t.Errorf("Expected default '-', got %s", req.InformasiTambahan)
		}
	})

	t.Run("Invalid empty required fields", func(t *testing.T) {
		req := laboratorium.SimpanPermintaanLabPKRequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				NoRawat:           "",
				TanggalPermintaan: "",
				JamPermintaan:     "",
				DiagnosaKlinis:    "",
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabPKRequest{},
		}

		errs := req.Validate()
		if errs == nil {
			t.Fatal("Expected validation errors, got nil")
		}
		if _, ok := errs["no_rawat"]; !ok {
			t.Error("Expected error on no_rawat")
		}
		if _, ok := errs["tanggal_permintaan"]; !ok {
			t.Error("Expected error on tanggal_permintaan")
		}
		if _, ok := errs["jam_permintaan"]; !ok {
			t.Error("Expected error on jam_permintaan")
		}
		if _, ok := errs["diagnosa_klinis"]; !ok {
			t.Error("Expected error on diagnosa_klinis")
		}
		if _, ok := errs["pemeriksaan"]; !ok {
			t.Error("Expected error on pemeriksaan")
		}
	})

	t.Run("Invalid date and time formats", func(t *testing.T) {
		req := laboratorium.SimpanPermintaanLabPKRequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				TanggalPermintaan: "03-09-2026",
				JamPermintaan:     "jam-10",
				DiagnosaKlinis:    "Anemia",
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabPKRequest{
				{
					IdTindakan: "",
				},
			},
		}

		errs := req.Validate()
		if errs == nil {
			t.Fatal("Expected validation errors, got nil")
		}
		if _, ok := errs["tanggal_permintaan"]; !ok {
			t.Error("Expected error on format tanggal_permintaan")
		}
		if _, ok := errs["jam_permintaan"]; !ok {
			t.Error("Expected error on format jam_permintaan")
		}
		if _, ok := errs["pemeriksaan[0].id_tindakan"]; !ok {
			t.Error("Expected error on empty id_tindakan")
		}
	})

	t.Run("Invalid future time", func(t *testing.T) {
		besok := time.Now().Add(24 * time.Hour).Format("2006-01-02")
		req := laboratorium.SimpanPermintaanLabPKRequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				TanggalPermintaan: besok,
				JamPermintaan:     "10:00:00",
				DiagnosaKlinis:    "Anemia",
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabPKRequest{
				{IdTindakan: "enc-tindakan-1"},
			},
		}

		errs := req.Validate()
		if errs == nil {
			t.Fatal("Expected validation errors for future time, got nil")
		}
		if _, ok := errs["tanggal_permintaan"]; !ok {
			t.Error("Expected error on future tanggal_permintaan")
		}
	})

	t.Run("Invalid empty id_template in slice", func(t *testing.T) {
		req := laboratorium.SimpanPermintaanLabPKRequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				TanggalPermintaan: nowStr,
				JamPermintaan:     jamNowStr,
				DiagnosaKlinis:    "Anemia",
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabPKRequest{
				{
					IdTindakan: "enc-tindakan-1",
					IdTemplate: []string{""},
				},
			},
		}

		errs := req.Validate()
		if errs == nil {
			t.Fatal("Expected validation errors for empty id_template, got nil")
		}
		if _, ok := errs["pemeriksaan[0].id_template[0]"]; !ok {
			t.Error("Expected error on pemeriksaan[0].id_template[0]")
		}
	})
}

func TestSimpanPermintaanLabPARequest_Sanitize(t *testing.T) {
	req := laboratorium.SimpanPermintaanLabPARequest{
		PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
			NoRawat:           "  2026/04/22/000001  ",
			TanggalPermintaan: "  2026-04-22  ",
			JamPermintaan:     "  14:30  ",
			DiagnosaKlinis:    "  Susp. Ca Mammae  ",
			InformasiTambahan: "   ",
		},
		DataKlinisJaringanPA: laboratorium.DataKlinisJaringanPA{
			PengambilanBahan: "  2026-04-21  ",
			DiperolehDengan:  "  Biopsi  ",
			LokasiJaringan:   "  Mammae Dextra  ",
			DiawetkanDengan:  "  Formalin 10%  ",
		},
		Pemeriksaan: []laboratorium.ItemPemeriksaanLabPARequest{
			{IdTindakan: "  enc-tindakan-pa  "},
		},
	}

	req.Sanitize()

	if req.NoRawat != "2026/04/22/000001" {
		t.Errorf("Expected sanitized NoRawat '2026/04/22/000001', got '%s'", req.NoRawat)
	}
	if req.JamPermintaan != "14:30:00" {
		t.Errorf("Expected sanitized JamPermintaan '14:30:00', got '%s'", req.JamPermintaan)
	}
	if req.InformasiTambahan != "-" {
		t.Errorf("Expected sanitized InformasiTambahan '-', got '%s'", req.InformasiTambahan)
	}
	if req.DiperolehDengan != "Biopsi" {
		t.Errorf("Expected DiperolehDengan 'Biopsi', got '%s'", req.DiperolehDengan)
	}
	if req.Pemeriksaan[0].IdTindakan != "enc-tindakan-pa" {
		t.Errorf("Expected IdTindakan 'enc-tindakan-pa', got '%s'", req.Pemeriksaan[0].IdTindakan)
	}
	if req.TanggalPASebelumnya != "0000-00-00" {
		t.Errorf("Expected TanggalPASebelumnya default '0000-00-00', got '%s'", req.TanggalPASebelumnya)
	}

	reqKosongBahan := laboratorium.SimpanPermintaanLabPARequest{
		PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
			TanggalPermintaan: "2026-04-22",
		},
		DataKlinisJaringanPA: laboratorium.DataKlinisJaringanPA{
			PengambilanBahan: "   ",
		},
	}
	reqKosongBahan.Sanitize()
	if reqKosongBahan.PengambilanBahan != "2026-04-22" {
		t.Errorf("Expected PengambilanBahan default to TanggalPermintaan '2026-04-22', got '%s'", reqKosongBahan.PengambilanBahan)
	}
}

func TestSimpanPermintaanLabPARequest_Validate(t *testing.T) {
	now := time.Now()
	nowStr := now.Format("2006-01-02")
	jamNowStr := now.Format("15:04:05")

	t.Run("Valid request with all optional fields", func(t *testing.T) {
		req := laboratorium.SimpanPermintaanLabPARequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				NoRawat:           "2026/04/22/000001",
				TanggalPermintaan: nowStr,
				JamPermintaan:     jamNowStr,
				DiagnosaKlinis:    "Susp. Tumor Mammae",
				InformasiTambahan: "Nodul teraba kenyal",
			},
			DataKlinisJaringanPA: laboratorium.DataKlinisJaringanPA{
				PengambilanBahan:     nowStr,
				DiperolehDengan:      "Biopsi Eksisi",
				LokasiJaringan:       "Mammae Dextra",
				DiawetkanDengan:      "Formalin Buffer 10%",
				PernahDilakukanDi:    "RSUD Lain",
				TanggalPASebelumnya:  "2025-01-10",
				NomorPASebelumnya:    "PA250001",
				DiagnosaPASebelumnya: "Fibroadenoma",
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabPARequest{
				{IdTindakan: "enc-tindakan-1"},
			},
		}

		errs := req.Validate()
		if errs != nil {
			t.Fatalf("Expected valid request, got errors: %v", errs)
		}
	})

	t.Run("Valid request minimal without optional fields", func(t *testing.T) {
		req := laboratorium.SimpanPermintaanLabPARequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				NoRawat:           "2026/04/22/000001",
				TanggalPermintaan: nowStr,
				JamPermintaan:     jamNowStr,
				DiagnosaKlinis:    "Susp. Tumor Mammae",
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabPARequest{
				{IdTindakan: "enc-tindakan-1"},
			},
		}

		errs := req.Validate()
		if errs != nil {
			t.Fatalf("Expected valid request, got errors: %v", errs)
		}
	})

	t.Run("Invalid empty fields", func(t *testing.T) {
		req := laboratorium.SimpanPermintaanLabPARequest{}
		errs := req.Validate()
		if errs == nil {
			t.Fatal("Expected validation errors, got nil")
		}
		if _, ok := errs["no_rawat"]; !ok {
			t.Error("Expected error on empty no_rawat")
		}
		if _, ok := errs["tanggal_permintaan"]; !ok {
			t.Error("Expected error on empty tanggal_permintaan")
		}
		if _, ok := errs["jam_permintaan"]; !ok {
			t.Error("Expected error on empty jam_permintaan")
		}
		if _, ok := errs["diagnosa_klinis"]; !ok {
			t.Error("Expected error on empty diagnosa_klinis")
		}
		if _, ok := errs["pemeriksaan"]; !ok {
			t.Error("Expected error on empty pemeriksaan")
		}
	})

	t.Run("Invalid format and future dates", func(t *testing.T) {
		besok := time.Now().Add(24 * time.Hour).Format("2006-01-02")
		req := laboratorium.SimpanPermintaanLabPARequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				NoRawat:           "2026/04/22/000001",
				TanggalPermintaan: besok,
				JamPermintaan:     "10:00:00",
				DiagnosaKlinis:    "Tumor",
			},
			DataKlinisJaringanPA: laboratorium.DataKlinisJaringanPA{
				PengambilanBahan:    besok,
				TanggalPASebelumnya: "invalid-date",
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabPARequest{
				{IdTindakan: ""},
			},
		}

		errs := req.Validate()
		if errs == nil {
			t.Fatal("Expected validation errors, got nil")
		}
		if _, ok := errs["tanggal_permintaan"]; !ok {
			t.Error("Expected error on future tanggal_permintaan")
		}
		if _, ok := errs["pengambilan_bahan"]; !ok {
			t.Error("Expected error on future pengambilan_bahan")
		}
		if _, ok := errs["tanggal_pa_sebelumnya"]; !ok {
			t.Error("Expected error on invalid tanggal_pa_sebelumnya")
		}
		if _, ok := errs["pemeriksaan[0].id_tindakan"]; !ok {
			t.Error("Expected error on empty pemeriksaan[0].id_tindakan")
		}
	})

	t.Run("Invalid character length limits", func(t *testing.T) {
		req := laboratorium.SimpanPermintaanLabPARequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				NoRawat:           "2026/04/22/000001",
				TanggalPermintaan: nowStr,
				JamPermintaan:     jamNowStr,
				DiagnosaKlinis:    string(make([]byte, 85)),
				InformasiTambahan: string(make([]byte, 65)),
			},
			DataKlinisJaringanPA: laboratorium.DataKlinisJaringanPA{
				DiperolehDengan:   string(make([]byte, 45)),
				LokasiJaringan:    string(make([]byte, 45)),
				DiawetkanDengan:   string(make([]byte, 45)),
				PernahDilakukanDi: string(make([]byte, 105)),
				NomorPASebelumnya: string(make([]byte, 25)),
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabPARequest{
				{IdTindakan: "enc-1"},
			},
		}

		errs := req.Validate()
		if errs == nil {
			t.Fatal("Expected length validation errors, got nil")
		}
		if _, ok := errs["diagnosa_klinis"]; !ok {
			t.Error("Expected error on long diagnosa_klinis")
		}
		if _, ok := errs["informasi_tambahan"]; !ok {
			t.Error("Expected error on long informasi_tambahan")
		}
		if _, ok := errs["diperoleh_dengan"]; !ok {
			t.Error("Expected error on long diperoleh_dengan")
		}
		if _, ok := errs["lokasi_jaringan"]; !ok {
			t.Error("Expected error on long lokasi_jaringan")
		}
		if _, ok := errs["diawetkan_dengan"]; !ok {
			t.Error("Expected error on long diawetkan_dengan")
		}
		if _, ok := errs["pernah_dilakukan_di"]; !ok {
			t.Error("Expected error on long pernah_dilakukan_di")
		}
		if _, ok := errs["nomor_pa_sebelumnya"]; !ok {
			t.Error("Expected error on long nomor_pa_sebelumnya")
		}
	})
}

func TestSimpanPermintaanLabMBRequest_ValidationAndSanitization(t *testing.T) {
	now := time.Now().Add(-1 * time.Minute)
	nowStr := now.Format("2006-01-02")
	jamNowStr := now.Format("15:04:05")

	t.Run("Valid request with full fields", func(t *testing.T) {
		req := laboratorium.SimpanPermintaanLabMBRequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				NoRawat:           "2026/09/04/000001",
				TanggalPermintaan: nowStr,
				JamPermintaan:     jamNowStr,
				DiagnosaKlinis:    "Kultur Sputum",
				InformasiTambahan: "Curiga TB",
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabMBRequest{
				{
					IdTindakan: "enc-tindakan-mb-1",
					IdTemplate: []string{"enc-tmpl-1", "enc-tmpl-2"},
				},
			},
		}

		req.Sanitize()
		errs := req.Validate()
		if errs != nil {
			t.Fatalf("Expected nil errors, got %v", errs)
		}
		if req.InformasiTambahan != "Curiga TB" {
			t.Errorf("Expected 'Curiga TB', got %s", req.InformasiTambahan)
		}
	})

	t.Run("Valid request with HH:MM format and empty informasi_tambahan defaults to hyphen", func(t *testing.T) {
		req := laboratorium.SimpanPermintaanLabMBRequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				NoRawat:           "2026/09/04/000001",
				TanggalPermintaan: nowStr,
				JamPermintaan:     now.Format("15:04"),
				DiagnosaKlinis:    "Kultur Darah",
				InformasiTambahan: "   ",
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabMBRequest{
				{
					IdTindakan: "  enc-tindakan-mb-1  ",
				},
			},
		}

		req.Sanitize()
		errs := req.Validate()
		if errs != nil {
			t.Fatalf("Expected nil errors, got %v", errs)
		}
		if req.InformasiTambahan != "-" {
			t.Errorf("Expected '-', got %s", req.InformasiTambahan)
		}
		if len(req.JamPermintaan) != 8 {
			t.Errorf("Expected jam format with seconds (8 chars), got %s", req.JamPermintaan)
		}
		if req.Pemeriksaan[0].IdTindakan != "enc-tindakan-mb-1" {
			t.Errorf("Expected trimmed id_tindakan, got %s", req.Pemeriksaan[0].IdTindakan)
		}
	})

	t.Run("Invalid empty required fields", func(t *testing.T) {
		req := laboratorium.SimpanPermintaanLabMBRequest{}
		req.Sanitize()
		errs := req.Validate()
		if errs == nil {
			t.Fatal("Expected validation errors, got nil")
		}
		if _, ok := errs["no_rawat"]; !ok {
			t.Error("Expected error on no_rawat")
		}
		if _, ok := errs["tanggal_permintaan"]; !ok {
			t.Error("Expected error on tanggal_permintaan")
		}
		if _, ok := errs["jam_permintaan"]; !ok {
			t.Error("Expected error on jam_permintaan")
		}
		if _, ok := errs["diagnosa_klinis"]; !ok {
			t.Error("Expected error on diagnosa_klinis")
		}
		if _, ok := errs["pemeriksaan"]; !ok {
			t.Error("Expected error on pemeriksaan")
		}
	})

	t.Run("Invalid date and time formats", func(t *testing.T) {
		req := laboratorium.SimpanPermintaanLabMBRequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				TanggalPermintaan: "04-09-2026",
				JamPermintaan:     "jam-10",
				DiagnosaKlinis:    "Sepsis",
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabMBRequest{
				{
					IdTindakan: "",
				},
			},
		}

		errs := req.Validate()
		if errs == nil {
			t.Fatal("Expected validation errors, got nil")
		}
		if _, ok := errs["tanggal_permintaan"]; !ok {
			t.Error("Expected error on format tanggal_permintaan")
		}
		if _, ok := errs["jam_permintaan"]; !ok {
			t.Error("Expected error on format jam_permintaan")
		}
		if _, ok := errs["pemeriksaan[0].id_tindakan"]; !ok {
			t.Error("Expected error on empty id_tindakan")
		}
	})

	t.Run("Invalid future time", func(t *testing.T) {
		besok := time.Now().Add(24 * time.Hour).Format("2006-01-02")
		req := laboratorium.SimpanPermintaanLabMBRequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				TanggalPermintaan: besok,
				JamPermintaan:     "10:00:00",
				DiagnosaKlinis:    "Sepsis",
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabMBRequest{
				{IdTindakan: "enc-tindakan-mb-1"},
			},
		}

		errs := req.Validate()
		if errs == nil {
			t.Fatal("Expected validation errors for future time, got nil")
		}
		if _, ok := errs["tanggal_permintaan"]; !ok {
			t.Error("Expected error on future tanggal_permintaan")
		}
	})

	t.Run("Invalid empty id_template in slice", func(t *testing.T) {
		req := laboratorium.SimpanPermintaanLabMBRequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				TanggalPermintaan: nowStr,
				JamPermintaan:     jamNowStr,
				DiagnosaKlinis:    "Sepsis",
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabMBRequest{
				{
					IdTindakan: "enc-tindakan-mb-1",
					IdTemplate: []string{""},
				},
			},
		}

		errs := req.Validate()
		if errs == nil {
			t.Fatal("Expected validation errors for empty template id, got nil")
		}
		if _, ok := errs["pemeriksaan[0].id_template[0]"]; !ok {
			t.Error("Expected error on empty id_template[0]")
		}
	})
}

func TestSimpanPermintaanLab_DuplicateTindakanAndTemplate_Validate(t *testing.T) {
	past := time.Now().Add(-5 * time.Minute)
	tgl := past.Format("2006-01-02")
	jam := past.Format("15:04:05")

	t.Run("PK menolak duplikat id_tindakan dan id_template", func(t *testing.T) {
		req := laboratorium.SimpanPermintaanLabPKRequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				NoRawat:           "2026/09/03/000001",
				TanggalPermintaan: tgl,
				JamPermintaan:     jam,
				DiagnosaKlinis:    "Febris",
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabPKRequest{
				{IdTindakan: "enc-pk-1", IdTemplate: []string{"enc-t1", "enc-t1"}},
				{IdTindakan: "enc-pk-1"},
			},
		}
		errs := req.Validate()
		if errs == nil {
			t.Fatal("expected validation errors, got nil")
		}
		if errs["pemeriksaan[0].id_template[1]"] == "" {
			t.Errorf("expected duplicate template error on pemeriksaan[0].id_template[1], got %v", errs)
		}
		if errs["pemeriksaan[1].id_tindakan"] == "" {
			t.Errorf("expected duplicate tindakan error on pemeriksaan[1].id_tindakan, got %v", errs)
		}
	})

	t.Run("PA menolak duplikat id_tindakan", func(t *testing.T) {
		req := laboratorium.SimpanPermintaanLabPARequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				NoRawat:           "2026/09/03/000001",
				TanggalPermintaan: tgl,
				JamPermintaan:     jam,
				DiagnosaKlinis:    "Tumor",
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabPARequest{
				{IdTindakan: "enc-pa-1"},
				{IdTindakan: "enc-pa-1"},
			},
		}
		req.Sanitize()
		errs := req.Validate()
		if errs == nil || errs["pemeriksaan[1].id_tindakan"] == "" {
			t.Fatalf("expected duplicate tindakan error on pemeriksaan[1].id_tindakan, got %v", errs)
		}
	})

	t.Run("MB menolak duplikat id_tindakan dan id_template", func(t *testing.T) {
		req := laboratorium.SimpanPermintaanLabMBRequest{
			PermintaanLabHeaderRequest: laboratorium.PermintaanLabHeaderRequest{
				NoRawat:           "2026/09/03/000001",
				TanggalPermintaan: tgl,
				JamPermintaan:     jam,
				DiagnosaKlinis:    "Sepsis",
			},
			Pemeriksaan: []laboratorium.ItemPemeriksaanLabMBRequest{
				{IdTindakan: "enc-mb-1", IdTemplate: []string{"enc-t1", "enc-t1"}},
				{IdTindakan: "enc-mb-1"},
			},
		}
		errs := req.Validate()
		if errs == nil {
			t.Fatal("expected validation errors, got nil")
		}
		if errs["pemeriksaan[0].id_template[1]"] == "" {
			t.Errorf("expected duplicate template error on pemeriksaan[0].id_template[1], got %v", errs)
		}
		if errs["pemeriksaan[1].id_tindakan"] == "" {
			t.Errorf("expected duplicate tindakan error on pemeriksaan[1].id_tindakan, got %v", errs)
		}
	})
}
