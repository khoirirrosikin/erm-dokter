package radiologi_test

import (
	"testing"

	"erm-dokter/internal/radiologi"
)

func TestFilterRiwayatRadiologi_Sanitize(t *testing.T) {
	f := radiologi.FilterRiwayatRadiologi{
		Tanggal: "   2026-01-01,2026-01-31   ",
		Page:    -1,
		Limit:   0,
	}
	f.Sanitize()

	if f.Page != 1 {
		t.Errorf("Expected page to default to 1, got %d", f.Page)
	}
	if f.Limit != 5 {
		t.Errorf("Expected limit to default to 5, got %d", f.Limit)
	}
	if f.Tanggal != "2026-01-01,2026-01-31" {
		t.Errorf("Expected trimmed tanggal, got %q", f.Tanggal)
	}

	fOver := radiologi.FilterRiwayatRadiologi{
		Page:  2,
		Limit: 200,
	}
	fOver.Sanitize()
	if fOver.Limit != 100 {
		t.Errorf("Expected max limit 100, got %d", fOver.Limit)
	}
}

func TestFilterRiwayatRadiologi_Offset(t *testing.T) {
	f := radiologi.FilterRiwayatRadiologi{Page: 1, Limit: 5}
	if f.Offset() != 0 {
		t.Errorf("Expected offset 0 for page 1, got %d", f.Offset())
	}

	f2 := radiologi.FilterRiwayatRadiologi{Page: 3, Limit: 5}
	if f2.Offset() != 10 {
		t.Errorf("Expected offset 10 for page 3 with limit 5, got %d", f2.Offset())
	}
}

func TestFilterRiwayatRadiologi_Validate(t *testing.T) {
	tests := []struct {
		name        string
		filter      radiologi.FilterRiwayatRadiologi
		expectError bool
		errorField  string
	}{
		{
			name: "Valid Filter without tanggal",
			filter: radiologi.FilterRiwayatRadiologi{
				Page:  1,
				Limit: 5,
			},
			expectError: false,
		},
		{
			name: "Valid Filter with single date range",
			filter: radiologi.FilterRiwayatRadiologi{
				Tanggal: "2026-01-01,2026-01-01",
				Page:    1,
				Limit:   5,
			},
			expectError: false,
		},
		{
			name: "Valid Filter with date range",
			filter: radiologi.FilterRiwayatRadiologi{
				Tanggal: "2026-01-01,2026-01-31",
				Page:    1,
				Limit:   5,
			},
			expectError: false,
		},
		{
			name: "Invalid Single Date without comma",
			filter: radiologi.FilterRiwayatRadiologi{
				Tanggal: "2026-01-01",
				Page:    1,
				Limit:   5,
			},
			expectError: true,
			errorField:  "tanggal",
		},
		{
			name: "Invalid Date format",
			filter: radiologi.FilterRiwayatRadiologi{
				Tanggal: "invalid-date",
				Page:    1,
				Limit:   5,
			},
			expectError: true,
			errorField:  "tanggal",
		},
		{
			name: "Invalid Range start > end",
			filter: radiologi.FilterRiwayatRadiologi{
				Tanggal: "2026-02-01,2026-01-01",
				Page:    1,
				Limit:   5,
			},
			expectError: true,
			errorField:  "tanggal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.filter.Validate()
			if tt.expectError {
				if errs == nil {
					t.Fatalf("Expected validation error, got nil")
				}
				if _, ok := errs[tt.errorField]; !ok {
					t.Errorf("Expected error in field %s, got errors: %v", tt.errorField, errs)
				}
				return
			}
			if errs != nil {
				t.Fatalf("Expected no validation errors, got: %v", errs)
			}
		})
	}
}

func TestIdHasilRadiologi_CompositeKey(t *testing.T) {
	id := radiologi.IdHasilRadiologi{
		NoRawat:        "2026/04/22/000001",
		KodeTindakan:   "RAD001",
		TanggalPeriksa: "2026-04-22",
		JamPeriksa:     "10:00:00",
	}

	expected := "2026/04/22/000001~RAD001~2026-04-22~10:00:00"
	if key := id.CompositeKey(); key != expected {
		t.Errorf("Expected composite key %s, got %s", expected, key)
	}
}

func TestParseIdHasilRadiologi(t *testing.T) {
	validKey := "2026/04/22/000001~RAD001~2026-04-22~10:00:00"
	id, err := radiologi.ParseIdHasilRadiologi(validKey)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if id.NoRawat != "2026/04/22/000001" || id.KodeTindakan != "RAD001" || id.TanggalPeriksa != "2026-04-22" || id.JamPeriksa != "10:00:00" {
		t.Errorf("Parsed fields mismatched: %+v", id)
	}

	invalidKey := "2026/04/22/000001~RAD001"
	_, err = radiologi.ParseIdHasilRadiologi(invalidKey)
	if err == nil {
		t.Fatalf("Expected error for invalid key, got nil")
	}
}

func TestHasilRadiologi_CompositeKey(t *testing.T) {
	hasil := radiologi.HasilRadiologi{
		NoRawat:        "2026/04/22/000001",
		KodeTindakan:   "RAD001",
		TanggalPeriksa: "2026-04-22",
		JamPeriksa:     "10:00:00",
	}

	expected := "2026/04/22/000001~RAD001~2026-04-22~10:00:00"
	if key := hasil.CompositeKey(); key != expected {
		t.Errorf("Expected composite key %s, got %s", expected, key)
	}
}

func TestSimpanPermintaanRadiologiRequest_SanitizeAndValidate(t *testing.T) {
	req := radiologi.SimpanPermintaanRadiologiRequest{
		NoRawat:           "  2026/09/05/000001  ",
		TanggalPermintaan: "  2026-09-05  ",
		JamPermintaan:     "  10:00  ",
		InformasiTambahan: "",
		DiagnosaKlinis:    "",
		Pemeriksaan: []radiologi.ItemPemeriksaanRadiologiRequest{
			{IdTindakan: "  enc123  "},
		},
	}
	req.Sanitize()

	if req.NoRawat != "2026/09/05/000001" {
		t.Errorf("Expected trimmed NoRawat, got %q", req.NoRawat)
	}
	if req.JamPermintaan != "10:00:00" {
		t.Errorf("Expected auto-appended :00, got %q", req.JamPermintaan)
	}
	if req.InformasiTambahan != "-" {
		t.Errorf("Expected '-' for empty InformasiTambahan, got %q", req.InformasiTambahan)
	}
	if req.DiagnosaKlinis != "-" {
		t.Errorf("Expected '-' for empty DiagnosaKlinis, got %q", req.DiagnosaKlinis)
	}
	if req.Pemeriksaan[0].IdTindakan != "enc123" {
		t.Errorf("Expected trimmed IdTindakan, got %q", req.Pemeriksaan[0].IdTindakan)
	}

	errs := req.Validate()
	if errs != nil {
		t.Fatalf("Expected valid request, got errs: %v", errs)
	}

	invalidReq := radiologi.SimpanPermintaanRadiologiRequest{
		NoRawat: "",
	}
	errsInvalid := invalidReq.Validate()
	if errsInvalid == nil {
		t.Fatalf("Expected validation error for empty fields, got nil")
	}
	if _, ok := errsInvalid["no_rawat"]; !ok {
		t.Errorf("Expected no_rawat error, got %v", errsInvalid)
	}
	if _, ok := errsInvalid["pemeriksaan"]; !ok {
		t.Errorf("Expected pemeriksaan error, got %v", errsInvalid)
	}

	dupReq := radiologi.SimpanPermintaanRadiologiRequest{
		NoRawat:           "2026/09/05/000001",
		TanggalPermintaan: "2026-09-05",
		JamPermintaan:     "10:00:00",
		DiagnosaKlinis:    "Batuk",
		Pemeriksaan: []radiologi.ItemPemeriksaanRadiologiRequest{
			{IdTindakan: "enc123"},
			{IdTindakan: "enc123"},
		},
	}
	errsDup := dupReq.Validate()
	if errsDup == nil || errsDup["pemeriksaan[1].id_tindakan"] == "" {
		t.Fatalf("Expected duplicate id_tindakan validation error on pemeriksaan[1].id_tindakan, got %v", errsDup)
	}
}

func TestFilterRiwayatPermintaanRadiologi_SanitizeAndValidate(t *testing.T) {
	f := radiologi.FilterRiwayatPermintaanRadiologi{
		Tanggal: "  2026-09-01,2026-09-05  ",
		Page:    0,
		Limit:   200,
	}
	f.Sanitize()

	if f.Page != 1 {
		t.Errorf("Expected page 1, got %d", f.Page)
	}
	if f.Limit != 100 {
		t.Errorf("Expected limit capped at 100, got %d", f.Limit)
	}
	if f.Offset() != 0 {
		t.Errorf("Expected offset 0, got %d", f.Offset())
	}

	errs := f.Validate()
	if errs != nil {
		t.Fatalf("Expected valid filter, got %v", errs)
	}
}
