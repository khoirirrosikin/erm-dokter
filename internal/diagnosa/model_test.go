package diagnosa_test

import (
	"testing"

	"erm-dokter/internal/diagnosa"
)

func intPtr(v int) *int {
	return &v
}

func TestDiagnosaProsedurRequests_PrioritasBounds(t *testing.T) {
	t.Run("TambahDiagnosaRequest batas prioritas 1..127", func(t *testing.T) {
		okReq := diagnosa.TambahDiagnosaRequest{Kode: "I10", Prioritas: intPtr(127)}
		if errs := okReq.Validate(); errs != nil {
			t.Fatalf("expected prioritas 127 valid, got %v", errs)
		}

		zeroReq := diagnosa.TambahDiagnosaRequest{Kode: "I10", Prioritas: intPtr(0)}
		if errs := zeroReq.Validate(); errs == nil || errs["prioritas"] == "" {
			t.Fatalf("expected prioritas < 1 error, got %v", errs)
		}

		overReq := diagnosa.TambahDiagnosaRequest{Kode: "I10", Prioritas: intPtr(128)}
		if errs := overReq.Validate(); errs == nil || errs["prioritas"] == "" {
			t.Fatalf("expected prioritas > 127 error, got %v", errs)
		}
	})

	t.Run("UpdateDiagnosaRequest batas prioritas 1..127", func(t *testing.T) {
		okReq := diagnosa.UpdateDiagnosaRequest{Prioritas: intPtr(127)}
		if errs := okReq.Validate(); errs != nil {
			t.Fatalf("expected prioritas 127 valid, got %v", errs)
		}

		overReq := diagnosa.UpdateDiagnosaRequest{Prioritas: intPtr(128)}
		if errs := overReq.Validate(); errs == nil || errs["prioritas"] == "" {
			t.Fatalf("expected prioritas > 127 error, got %v", errs)
		}
	})

	t.Run("TambahProsedurRequest batas prioritas 1..127", func(t *testing.T) {
		okReq := diagnosa.TambahProsedurRequest{Kode: "87.03", Prioritas: intPtr(127)}
		if errs := okReq.Validate(); errs != nil {
			t.Fatalf("expected prioritas 127 valid, got %v", errs)
		}

		overReq := diagnosa.TambahProsedurRequest{Kode: "87.03", Prioritas: intPtr(128)}
		if errs := overReq.Validate(); errs == nil || errs["prioritas"] == "" {
			t.Fatalf("expected prioritas > 127 error, got %v", errs)
		}
	})

	t.Run("UpdateProsedurRequest batas prioritas 1..127", func(t *testing.T) {
		okReq := diagnosa.UpdateProsedurRequest{Prioritas: intPtr(127)}
		if errs := okReq.Validate(); errs != nil {
			t.Fatalf("expected prioritas 127 valid, got %v", errs)
		}

		overReq := diagnosa.UpdateProsedurRequest{Prioritas: intPtr(128)}
		if errs := overReq.Validate(); errs == nil || errs["prioritas"] == "" {
			t.Fatalf("expected prioritas > 127 error, got %v", errs)
		}
	})

	t.Run("ReorderRequest batas prioritas 1..127", func(t *testing.T) {
		okReq := diagnosa.ReorderRequest{
			Items: []diagnosa.ReorderItemRequest{{Kode: "I10", Prioritas: 127}},
		}
		if errs := okReq.Validate(); errs != nil {
			t.Fatalf("expected prioritas 127 valid, got %v", errs)
		}

		overReq := diagnosa.ReorderRequest{
			Items: []diagnosa.ReorderItemRequest{{Kode: "I10", Prioritas: 128}},
		}
		if errs := overReq.Validate(); errs == nil || errs["items[0].prioritas"] == "" {
			t.Fatalf("expected items[0].prioritas > 127 error, got %v", errs)
		}
	})
}

