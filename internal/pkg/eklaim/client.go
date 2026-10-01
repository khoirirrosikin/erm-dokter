package eklaim

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client interface {
	SimulasiGrouper(ctx context.Context, param ParameterSimulasi) (*HasilSimulasi, error)
}

type ParameterSimulasi struct {
	NomorSEP      string
	NomorKartu    string
	NomorRM       string
	NamaPasien    string
	TanggalLahir  string
	Gender        string
	JenisRawat    string
	KelasRawat    string
	TanggalMasuk  string
	TanggalPulang string
	CaraPulang    string
	Diagnosa      []string
	Prosedur      []string
	NamaDokter    string
	CoderNIK      string
}

type SpecialCMGDetail struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Tariff      int64  `json:"tariff"`
	Type        string `json:"type"`
}

type HasilSimulasi struct {
	KodeCBG       string             `json:"kode_cbg"`
	DeskripsiCBG  string             `json:"deskripsi_cbg"`
	Tarif         int64              `json:"tarif"`
	BaseTarif     int64              `json:"base_tarif"`
	Kelas         string             `json:"kelas"`
	JenisRawat    string             `json:"jenis_rawat"`
	SeverityLevel string             `json:"severity_level"`
	SpecialCMG    []SpecialCMGDetail `json:"special_cmg,omitempty"`
}

type client struct {
	baseURL         string
	encryptionKey   string
	kodeRS          string
	kodeTarif       string
	defaultCoderNIK string
	httpClient      *http.Client
}

func NewClient(
	baseURL string,
	encryptionKey string,
	kodeRS string,
	kodeTarif string,
	defaultCoderNIK string,
	timeout time.Duration,
) Client {
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	if kodeTarif == "" {
		kodeTarif = "CP"
	}
	if defaultCoderNIK == "" {
		defaultCoderNIK = "1234567890123456"
	}

	return &client{
		baseURL:         strings.TrimSpace(baseURL),
		encryptionKey:   strings.TrimSpace(encryptionKey),
		kodeRS:          strings.TrimSpace(kodeRS),
		kodeTarif:       strings.TrimSpace(kodeTarif),
		defaultCoderNIK: strings.TrimSpace(defaultCoderNIK),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *client) SimulasiGrouper(ctx context.Context, param ParameterSimulasi) (*HasilSimulasi, error) {
	if c.baseURL == "" || c.encryptionKey == "" {
		return nil, errors.New("konfigurasi E-Klaim belum disetel (EKLAIM_BASE_URL atau EKLAIM_ENCRYPTION_KEY kosong)")
	}

	if strings.TrimSpace(param.NomorSEP) == "" {
		return nil, errors.New("nomor SEP simulasi wajib diisi")
	}

	coderNIK := strings.TrimSpace(param.CoderNIK)
	if coderNIK == "" {
		coderNIK = c.defaultCoderNIK
	}

	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		deleteReq := map[string]any{
			"metadata": map[string]string{
				"method": "delete_claim",
			},
			"data": map[string]string{
				"nomor_sep": param.NomorSEP,
				"coder_nik": coderNIK,
			},
		}
		_, _ = c.sendRequest(cleanupCtx, deleteReq)
	}()

	nomorKartu := strings.TrimSpace(param.NomorKartu)
	if nomorKartu == "" {
		nomorKartu = "0000000000000"
	}

	newClaimReq := map[string]any{
		"metadata": map[string]string{
			"method": "new_claim",
		},
		"data": map[string]string{
			"nomor_kartu": nomorKartu,
			"nomor_sep":   param.NomorSEP,
			"nomor_rm":    param.NomorRM,
			"nama_pasien": param.NamaPasien,
			"tgl_lahir":   param.TanggalLahir,
			"gender":      param.Gender,
		},
	}

	_, _ = c.sendRequest(ctx, newClaimReq)

	caraPulang := param.CaraPulang
	if caraPulang == "" {
		caraPulang = "1"
	}

	jenisRawat := param.JenisRawat
	if jenisRawat == "" {
		jenisRawat = "2"
	}

	kelasRawat := param.KelasRawat
	if kelasRawat == "" {
		if jenisRawat == "1" {
			kelasRawat = "3"
		} else {
			kelasRawat = "3"
		}
	}

	tglMasuk := param.TanggalMasuk
	if len(tglMasuk) == 10 {
		tglMasuk += " 00:00:00"
	}
	tglPulang := param.TanggalPulang
	if len(tglPulang) == 10 {
		tglPulang += " 23:59:59"
	}

	setClaimReq := map[string]any{
		"metadata": map[string]string{
			"method":    "set_claim_data",
			"nomor_sep": param.NomorSEP,
		},
		"data": map[string]any{
			"nomor_sep":         param.NomorSEP,
			"nomor_kartu":       nomorKartu,
			"tgl_masuk":         tglMasuk,
			"tgl_pulang":        tglPulang,
			"cara_masuk":        "gp",
			"jenis_rawat":       jenisRawat,
			"kelas_rawat":       kelasRawat,
			"adl_sub_acute":     0,
			"adl_chronic":       0,
			"icu_indikator":     0,
			"icu_los":           0,
			"upgrade_class_ind": 0,
			"add_payment_pct":   0,
			"birth_weight":      0,
			"discharge_status":  1,
			"tarif_rs": map[string]any{
				"prosedur_non_bedah": 0,
				"prosedur_bedah":     0,
				"konsultasi":         0,
				"tenaga_ahli":        0,
				"keperawatan":        0,
				"penunjang":          0,
				"radiologi":          0,
				"laboratorium":       0,
				"pelayanan_darah":    0,
				"rehabilitasi":       0,
				"kamar":              0,
				"rawat_intensif":     0,
				"obat":               0,
				"obat_kronis":        0,
				"obat_kemoterapi":    0,
				"alkes":              0,
				"bmhp":               0,
				"sewa_alat":          0,
			},
			"pemulasaraan_jenazah": 0,
			"kantong_darah":        0,
			"tarif_poli_eks":       0,
			"nama_dokter":          param.NamaDokter,
			"kode_tarif":           c.kodeTarif,
			"payor_id":             3,
			"payor_cd":             "JKN",
			"cob_cd":               0,
			"coder_nik":            coderNIK,
		},
	}

	setClaimResp, err := c.sendRequest(ctx, setClaimReq)
	if err != nil {
		return nil, fmt.Errorf("gagal set data klaim E-Klaim: %w", err)
	}

	if setClaimResp.Metadata.Message != "Ok" && !strings.Contains(strings.ToLower(setClaimResp.Metadata.Message), "ok") {
		return nil, fmt.Errorf("E-Klaim menolak set_claim_data: %s", setClaimResp.Metadata.Message)
	}

	diagnosaStr := strings.Join(param.Diagnosa, "#")
	diagnosaSetReq := map[string]any{
		"metadata": map[string]string{
			"method":    "idrg_diagnosa_set",
			"nomor_sep": param.NomorSEP,
		},
		"data": map[string]string{
			"diagnosa": diagnosaStr,
		},
	}

	diagnosaResp, err := c.sendRequest(ctx, diagnosaSetReq)
	if err != nil {
		return nil, fmt.Errorf("gagal set diagnosa INACBG E-Klaim: %w", err)
	}

	if diagnosaResp.Metadata.Message != "Ok" && !strings.Contains(strings.ToLower(diagnosaResp.Metadata.Message), "ok") {
		return nil, fmt.Errorf("E-Klaim menolak inacbg_diagnosa_set: %s", diagnosaResp.Metadata.Message)
	}

	if len(param.Prosedur) > 0 {
		prosedurStr := strings.Join(param.Prosedur, "#")
		procedureSetReq := map[string]any{
			"metadata": map[string]string{
				"method":    "idrg_procedure_set",
				"nomor_sep": param.NomorSEP,
			},
			"data": map[string]string{
				"procedure": prosedurStr,
			},
		}

		procedureResp, err := c.sendRequest(ctx, procedureSetReq)
		if err != nil {
			return nil, fmt.Errorf("gagal set prosedur INACBG E-Klaim: %w", err)
		}

		if procedureResp.Metadata.Message != "Ok" && !strings.Contains(strings.ToLower(procedureResp.Metadata.Message), "ok") {
			return nil, fmt.Errorf("E-Klaim menolak inacbg_procedure_set: %s", procedureResp.Metadata.Message)
		}
	}

	grouperReq := map[string]any{
		"metadata": map[string]string{
			"method":  "grouper",
			"stage":   "1",
			"grouper": "idrg",
		},
		"data": map[string]string{
			"nomor_sep": param.NomorSEP,
		},
	}

	grouperResp, err := c.sendRequest(ctx, grouperReq)
	if err != nil {
		return nil, fmt.Errorf("gagal mengeksekusi grouper E-Klaim: %w", err)
	}

	if grouperResp.Metadata.Message != "Ok" && !strings.Contains(strings.ToLower(grouperResp.Metadata.Message), "ok") {
		return nil, fmt.Errorf("E-Klaim grouper error: %s", grouperResp.Metadata.Message)
	}

	return c.parseGrouperResult(grouperResp, jenisRawat)
}

type wsResponse struct {
	Metadata struct {
		Code    any    `json:"code"`
		Message string `json:"message"`
	} `json:"metadata"`
	Response struct {
		CBG struct {
			Code        string `json:"code"`
			Description string `json:"description"`
			Tariff      any    `json:"tariff"`
			BaseTariff  any    `json:"base_tariff"`
		} `json:"cbg"`
		SubAcute struct {
			Code        string `json:"code"`
			Description string `json:"description"`
			Tariff      any    `json:"tariff"`
		} `json:"sub_acute"`
		Chronic struct {
			Code        string `json:"code"`
			Description string `json:"description"`
			Tariff      any    `json:"tariff"`
		} `json:"chronic"`
		SpecialCMG []struct {
			Code        string `json:"code"`
			Description string `json:"description"`
			Tariff      any    `json:"tariff"`
			Type        string `json:"type"`
		} `json:"special_cmg"`
		Kelas string `json:"kelas"`
	} `json:"response"`
	ResponseINACBG struct {
		CBG struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"cbg"`
		BaseTariff any    `json:"base_tariff"`
		Tariff     any    `json:"tariff"`
		Kelas      string `json:"kelas"`
		SpecialCMG []struct {
			Code        string `json:"code"`
			Description string `json:"description"`
			Tariff      any    `json:"tariff"`
			Type        string `json:"type"`
		} `json:"special_cmg"`
	} `json:"response_inacbg"`
	ResponseIDRG struct {
		DRGCode        string `json:"drg_code"`
		DRGDescription string `json:"drg_description"`
		MDCNumber      string `json:"mdc_number"`
		MDCDescription string `json:"mdc_description"`
		CostWeight     string `json:"cost_weight"`
		TotalTarif     string `json:"total_tarif"`
		StatusCd       string `json:"status_cd"`
		ScriptVersion  string `json:"script_version"`
		LogicVersion   string `json:"logic_version"`
	} `json:"response_idrg"`
}

func (c *client) sendRequest(ctx context.Context, payload any) (*wsResponse, error) {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("gagal marshal request payload: %w", err)
	}

	encryptedPayload, err := Encrypt(jsonBytes, c.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("gagal enkripsi payload E-Klaim: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewBufferString(encryptedPayload))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("koneksi ke server E-Klaim gagal / timeout: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca respons server E-Klaim: %w", err)
	}

	if len(respBody) == 0 {
		return nil, errors.New("respons dari server E-Klaim kosong")
	}

	decrypted, err := Decrypt(string(respBody), c.encryptionKey)
	if err != nil {
		var rawErr wsResponse
		if jsonErr := json.Unmarshal(respBody, &rawErr); jsonErr == nil && rawErr.Metadata.Message != "" {
			return &rawErr, nil
		}
		return nil, fmt.Errorf("gagal mendekripsi respons E-Klaim: %w (body: %s)", err, strings.TrimSpace(string(respBody)))
	}

	var parsed wsResponse
	if err := json.Unmarshal(decrypted, &parsed); err != nil {
		return nil, fmt.Errorf("gagal unmarshal JSON respons E-Klaim: %w", err)
	}

	return &parsed, nil
}

func (c *client) parseGrouperResult(resp *wsResponse, jenisRawat string) (*HasilSimulasi, error) {
	cbgCode := resp.ResponseINACBG.CBG.Code
	cbgDesc := resp.ResponseINACBG.CBG.Description
	tarif := parseTariffValue(resp.ResponseINACBG.Tariff)
	baseTarif := parseTariffValue(resp.ResponseINACBG.BaseTariff)
	kelas := resp.ResponseINACBG.Kelas
	specialCMGSource := resp.ResponseINACBG.SpecialCMG

	if cbgCode == "" && resp.ResponseIDRG.DRGCode != "" {
		cbgCode = resp.ResponseIDRG.DRGCode
		cbgDesc = resp.ResponseIDRG.DRGDescription
		tarif = parseTariffValue(resp.ResponseIDRG.TotalTarif)
		baseTarif = tarif
	}

	if cbgCode == "" && resp.Response.CBG.Code != "" {
		cbgCode = resp.Response.CBG.Code
		cbgDesc = resp.Response.CBG.Description
		tarif = parseTariffValue(resp.Response.CBG.Tariff)
		baseTarif = parseTariffValue(resp.Response.CBG.BaseTariff)
		kelas = resp.Response.Kelas
		specialCMGSource = resp.Response.SpecialCMG
	}

	if cbgCode == "" {
		return nil, errors.New("hasil grouper E-Klaim tidak menghasilkan kode CBG/DRG")
	}

	if baseTarif == 0 {
		baseTarif = tarif
	}

	severityLevel := "-"
	parts := strings.Split(cbgCode, "-")
	if len(parts) >= 4 {
		severityLevel = parts[3]
	}

	namaJenisRawat := "Rawat Jalan"
	if jenisRawat == "1" {
		namaJenisRawat = "Rawat Inap"
	}

	if kelas == "" {
		if jenisRawat == "1" {
			kelas = "Rawat Inap"
		} else {
			kelas = "Rawat Jalan"
		}
	}

	var specialCMGs []SpecialCMGDetail
	for _, sc := range specialCMGSource {
		specialCMGs = append(specialCMGs, SpecialCMGDetail{
			Code:        sc.Code,
			Description: sc.Description,
			Tariff:      parseTariffValue(sc.Tariff),
			Type:        sc.Type,
		})
	}

	return &HasilSimulasi{
		KodeCBG:       cbgCode,
		DeskripsiCBG:  cbgDesc,
		Tarif:         tarif,
		BaseTarif:     baseTarif,
		Kelas:         kelas,
		JenisRawat:    namaJenisRawat,
		SeverityLevel: severityLevel,
		SpecialCMG:    specialCMGs,
	}, nil
}

func parseTariffValue(val any) int64 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	case string:
		clean := strings.TrimSpace(v)
		clean = strings.ReplaceAll(clean, ",", "")
		clean = strings.ReplaceAll(clean, ".", "")
		if parsed, err := strconv.ParseInt(clean, 10, 64); err == nil {
			return parsed
		}
	}
	return 0
}
