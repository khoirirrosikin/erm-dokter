package eklaim_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"erm-dokter/internal/pkg/eklaim"
)

func TestClient_SimulasiGrouper_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		decrypted, err := eklaim.Decrypt(string(body), testKeyHex)
		if err != nil {
			http.Error(w, "decrypt error: "+err.Error(), http.StatusBadRequest)
			return
		}

		var reqMap map[string]any
		if err := json.Unmarshal(decrypted, &reqMap); err != nil {
			http.Error(w, "unmarshal error: "+err.Error(), http.StatusBadRequest)
			return
		}

		metadata, _ := reqMap["metadata"].(map[string]any)
		method, _ := metadata["method"].(string)

		var respMap map[string]any

		switch method {
		case "new_claim":
			respMap = map[string]any{
				"metadata": map[string]any{
					"code":    200,
					"message": "Ok",
				},
			}
		case "set_claim_data":
			respMap = map[string]any{
				"metadata": map[string]any{
					"code":    200,
					"message": "Ok",
				},
			}
		case "inacbg_diagnosa_set", "inacbg_procedure_set", "idrg_diagnosa_set", "idrg_procedure_set":
			respMap = map[string]any{
				"metadata": map[string]any{
					"code":    200,
					"message": "Ok",
				},
			}
		case "grouper":
			respMap = map[string]any{
				"metadata": map[string]any{
					"code":    200,
					"message": "Ok",
				},
				"response_inacbg": map[string]any{
					"cbg": map[string]any{
						"code":        "I-4-17-I",
						"description": "GAGAL JANTUNG & SYOK KARDIOGENIK RINGAN",
					},
					"tariff":      "4250000",
					"base_tariff": "4250000",
					"kelas":       "kelas_3",
					"special_cmg": []map[string]any{
						{
							"code":        "YY-01-I",
							"description": "Special Procedure",
							"tariff":      500000,
							"type":        "Special Procedure",
						},
					},
				},
			}
		case "delete_claim":
			respMap = map[string]any{
				"metadata": map[string]any{
					"code":    200,
					"message": "Ok",
				},
			}
		default:
			respMap = map[string]any{
				"metadata": map[string]any{
					"code":    400,
					"message": "Unknown method: " + method,
				},
			}
		}

		respBytes, _ := json.Marshal(respMap)
		encryptedResp, _ := eklaim.Encrypt(respBytes, testKeyHex)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(encryptedResp))
	}))
	defer mockServer.Close()

	client := eklaim.NewClient(mockServer.URL, testKeyHex, "1234567", "CP", "1234567890123456", 2*time.Second)

	param := eklaim.ParameterSimulasi{
		NomorSEP:      "SIM-2026-09-18-000001",
		NomorKartu:    "0001234567890",
		NomorRM:       "123456",
		NamaPasien:    "Budi Santoso",
		TanggalLahir:  "1980-01-01",
		Gender:        "1",
		JenisRawat:    "2",
		KelasRawat:    "regular",
		TanggalMasuk:  "2026-09-18",
		TanggalPulang: "2026-09-18",
		CaraPulang:    "1",
		Diagnosa:      []string{"I50.9", "I10"},
		Prosedur:      []string{"88.72"},
		NamaDokter:    "dr. Dokter Spesialis",
	}

	result, err := client.SimulasiGrouper(context.Background(), param)
	if err != nil {
		t.Fatalf("SimulasiGrouper failed: %v", err)
	}

	if result.KodeCBG != "I-4-17-I" {
		t.Errorf("Expected KodeCBG 'I-4-17-I', got '%s'", result.KodeCBG)
	}
	if result.DeskripsiCBG != "GAGAL JANTUNG & SYOK KARDIOGENIK RINGAN" {
		t.Errorf("Expected DeskripsiCBG match, got '%s'", result.DeskripsiCBG)
	}
	if result.Tarif != 4250000 {
		t.Errorf("Expected Tarif 4250000, got %d", result.Tarif)
	}
	if result.SeverityLevel != "I" {
		t.Errorf("Expected SeverityLevel 'I', got '%s'", result.SeverityLevel)
	}
	if result.Kelas != "kelas_3" {
		t.Errorf("Expected Kelas 'kelas_3', got '%s'", result.Kelas)
	}
	if len(result.SpecialCMG) != 1 || result.SpecialCMG[0].Tariff != 500000 {
		t.Errorf("Unexpected SpecialCMG: %+v", result.SpecialCMG)
	}
}

func TestClient_SimulasiGrouper_MissingConfig(t *testing.T) {
	client := eklaim.NewClient("", "", "", "", "", 0)
	_, err := client.SimulasiGrouper(context.Background(), eklaim.ParameterSimulasi{NomorSEP: "SIM-01"})
	if err == nil {
		t.Fatal("Expected error for missing config, got nil")
	}
}

func TestClient_SimulasiGrouper_GrouperError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		decrypted, _ := eklaim.Decrypt(string(body), testKeyHex)

		var reqMap map[string]any
		_ = json.Unmarshal(decrypted, &reqMap)
		metadata, _ := reqMap["metadata"].(map[string]any)
		method, _ := metadata["method"].(string)

		var respMap map[string]any
		if method == "grouper" {
			respMap = map[string]any{
				"metadata": map[string]any{
					"code":    500,
					"message": "Kode ICD-10 tidak ditemukan dalam grouper",
				},
			}
		} else {
			respMap = map[string]any{
				"metadata": map[string]any{
					"code":    200,
					"message": "Ok",
				},
			}
		}

		respBytes, _ := json.Marshal(respMap)
		encryptedResp, _ := eklaim.Encrypt(respBytes, testKeyHex)
		_, _ = w.Write([]byte(encryptedResp))
	}))
	defer mockServer.Close()

	client := eklaim.NewClient(mockServer.URL, testKeyHex, "1234567", "CP", "1234567890123456", 2*time.Second)
	_, err := client.SimulasiGrouper(context.Background(), eklaim.ParameterSimulasi{
		NomorSEP: "SIM-01",
		Diagnosa: []string{"INVALID.CODE"},
	})
	if err == nil {
		t.Fatal("Expected grouper error, got nil")
	}
}

func TestClient_SimulasiGrouper_DecimalTariffAndCaraPulangAndDuplicateSEP(t *testing.T) {
	var capturedDischargeStatus float64

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		decrypted, _ := eklaim.Decrypt(string(body), testKeyHex)

		var reqMap map[string]any
		_ = json.Unmarshal(decrypted, &reqMap)
		metadata, _ := reqMap["metadata"].(map[string]any)
		method, _ := metadata["method"].(string)

		var respMap map[string]any
		switch method {
		case "new_claim":
			respMap = map[string]any{
				"metadata": map[string]any{
					"code":    400,
					"message": "Duplikasi nomor SEP",
				},
			}
		case "set_claim_data":
			dataMap, _ := reqMap["data"].(map[string]any)
			capturedDischargeStatus, _ = dataMap["discharge_status"].(float64)
			respMap = map[string]any{
				"metadata": map[string]any{"code": 200, "message": "Ok"},
			}
		case "grouper":
			respMap = map[string]any{
				"metadata": map[string]any{"code": 200, "message": "Ok"},
				"response_inacbg": map[string]any{
					"cbg": map[string]any{
						"code":        "I-4-10-II",
						"description": "INFARK MIOKARD AKUT SEDANG",
					},
					"tariff":      "1250000.00",
					"base_tariff": "1.100.000,00",
					"kelas":       "kelas_2",
					"special_cmg": []map[string]any{
						{
							"code":        "DD-01-II",
							"description": "Special Drug",
							"tariff":      "1,250,000.00",
							"type":        "Special Drug",
						},
					},
				},
			}
		default:
			respMap = map[string]any{
				"metadata": map[string]any{"code": 200, "message": "Ok"},
			}
		}

		respBytes, _ := json.Marshal(respMap)
		encryptedResp, _ := eklaim.Encrypt(respBytes, testKeyHex)
		_, _ = w.Write([]byte(encryptedResp))
	}))
	defer mockServer.Close()

	client := eklaim.NewClient(mockServer.URL, testKeyHex, "1234567", "CP", "1234567890123456", 2*time.Second)
	res, err := client.SimulasiGrouper(context.Background(), eklaim.ParameterSimulasi{
		NomorSEP:   "SIM-DEC-01",
		CaraPulang: "4",
		Diagnosa:   []string{"I21.9"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedDischargeStatus != 4 {
		t.Errorf("expected discharge_status 4 from CaraPulang '4', got %v", capturedDischargeStatus)
	}
	if res.Tarif != 1250000 {
		t.Errorf("expected Tarif 1250000 (not 100x), got %d", res.Tarif)
	}
	if res.BaseTarif != 1100000 {
		t.Errorf("expected BaseTarif 1100000, got %d", res.BaseTarif)
	}
	if len(res.SpecialCMG) != 1 || res.SpecialCMG[0].Tariff != 1250000 {
		t.Errorf("expected SpecialCMG Tariff 1250000, got %+v", res.SpecialCMG)
	}
}

func TestClient_SimulasiGrouper_NewClaimNetworkErrorExitsEarly(t *testing.T) {
	var callCount atomic.Int32
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		time.Sleep(150 * time.Millisecond)
	}))
	defer mockServer.Close()

	client := eklaim.NewClient(mockServer.URL, testKeyHex, "1234567", "CP", "1234567890123456", 40*time.Millisecond)
	_, err := client.SimulasiGrouper(context.Background(), eklaim.ParameterSimulasi{
		NomorSEP: "SIM-TIMEOUT-01",
		Diagnosa: []string{"I10"},
	})
	if err == nil {
		t.Fatal("expected timeout error on new_claim")
	}
	if got := callCount.Load(); got != 1 {
		t.Errorf("expected exactly 1 call (early exit after new_claim timeout without set_claim_data/delete_claim), got %d", got)
	}
}
