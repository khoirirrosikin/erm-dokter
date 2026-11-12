package master_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"erm-dokter/internal/master"
	"erm-dokter/internal/pkg/logger"
)

type mockRepository struct {
	penjaminData   []master.Penjamin
	depoData       []master.Depo
	poliklinikData []master.Poliklinik
	bangsalData    []master.Bangsal
	kelasData      []master.KelasKamar
	icd10Data      []master.ICD10
	icd10Total     int
	icd9Data       []master.ICD9
	icd9Total      int
	checkICD10Map  map[string]bool
	checkICD9Map   map[string]bool
	err            error
}

func (m *mockRepository) DaftarPenjamin(ctx context.Context) ([]master.Penjamin, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.penjaminData, nil
}

func (m *mockRepository) DaftarDepo(ctx context.Context) ([]master.Depo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.depoData, nil
}

func (m *mockRepository) DaftarPoliklinik(ctx context.Context) ([]master.Poliklinik, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.poliklinikData, nil
}

func (m *mockRepository) DaftarBangsal(ctx context.Context) ([]master.Bangsal, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.bangsalData, nil
}

func (m *mockRepository) DaftarKelas(ctx context.Context) ([]master.KelasKamar, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.kelasData, nil
}

func (m *mockRepository) DaftarICD10(ctx context.Context, filter master.FilterMasterICD) ([]master.ICD10, int, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	return m.icd10Data, m.icd10Total, nil
}

func (m *mockRepository) DaftarICD9(ctx context.Context, filter master.FilterMasterICD) ([]master.ICD9, int, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	return m.icd9Data, m.icd9Total, nil
}

func (m *mockRepository) FetchAllICD10(ctx context.Context) ([]master.ICD10, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.icd10Data, nil
}

func (m *mockRepository) FetchAllICD9(ctx context.Context) ([]master.ICD9, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.icd9Data, nil
}

func (m *mockRepository) CekKeberadaanICD10(ctx context.Context, listKode []string) (map[string]bool, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.checkICD10Map, nil
}

func (m *mockRepository) CekKeberadaanICD9(ctx context.Context, listKode []string) (map[string]bool, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.checkICD9Map, nil
}

func TestMasterService_DaftarPenjamin(t *testing.T) {
	repo := &mockRepository{
		penjaminData: []master.Penjamin{
			{ItemMaster: master.ItemMaster{Kode: "BPJ", Nama: "BPJS KESEHATAN"}},
		},
	}
	svc := master.NewService(repo, logger.New())

	data, err := svc.DaftarPenjamin(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data) != 1 || data[0].Kode != "BPJ" {
		t.Errorf("unexpected penjamin data: %+v", data)
	}

	repo.err = errors.New("db error")
	_, err = svc.DaftarPenjamin(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMasterService_DaftarDepo(t *testing.T) {
	repo := &mockRepository{
		depoData: []master.Depo{
			{ItemMaster: master.ItemMaster{Kode: "DPRJ", Nama: "Depo Rawat Jalan"}},
		},
	}
	svc := master.NewService(repo, logger.New())

	data, err := svc.DaftarDepo(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data) != 1 || data[0].Kode != "DPRJ" {
		t.Errorf("unexpected depo data: %+v", data)
	}

	repo.err = errors.New("db error")
	_, err = svc.DaftarDepo(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMasterService_DaftarPoliklinik(t *testing.T) {
	repo := &mockRepository{
		poliklinikData: []master.Poliklinik{
			{ItemMaster: master.ItemMaster{Kode: "INT", Nama: "Poli Penyakit Dalam"}},
		},
	}
	svc := master.NewService(repo, logger.New())

	data, err := svc.DaftarPoliklinik(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data) != 1 || data[0].Kode != "INT" {
		t.Errorf("unexpected poliklinik data: %+v", data)
	}

	repo.err = errors.New("db error")
	_, err = svc.DaftarPoliklinik(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMasterService_DaftarBangsal(t *testing.T) {
	repo := &mockRepository{
		bangsalData: []master.Bangsal{
			{ItemMaster: master.ItemMaster{Kode: "B01", Nama: "Melati"}},
		},
	}
	svc := master.NewService(repo, logger.New())

	data, err := svc.DaftarBangsal(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data) != 1 || data[0].Kode != "B01" || data[0].Nama != "Melati" {
		t.Errorf("unexpected bangsal data: %+v", data)
	}

	repo.err = errors.New("db error")
	_, err = svc.DaftarBangsal(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMasterService_DaftarKelas(t *testing.T) {
	repo := &mockRepository{
		kelasData: []master.KelasKamar{
			{ItemMaster: master.ItemMaster{Kode: "Kelas 1", Nama: "Kelas 1"}},
		},
	}
	svc := master.NewService(repo, logger.New())

	data, err := svc.DaftarKelas(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data) != 1 || data[0].Kode != "Kelas 1" || data[0].Nama != "Kelas 1" {
		t.Errorf("unexpected kelas data: %+v", data)
	}

	repo.err = errors.New("db error")
	_, err = svc.DaftarKelas(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMasterService_DaftarICD10(t *testing.T) {
	repo := &mockRepository{
		icd10Data: []master.ICD10{
			{ItemMaster: master.ItemMaster{Kode: "I63.9", Nama: "Cerebral infarction, unspecified"}},
		},
		icd10Total: 1,
	}
	svc := master.NewService(repo, logger.New())

	filter := master.FilterMasterICD{Keyword: "cerebral", Page: 1, Limit: 20}
	data, meta, err := svc.DaftarICD10(context.Background(), filter)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data) != 1 || data[0].Kode != "I63.9" {
		t.Errorf("unexpected ICD-10 data: %+v", data)
	}
	if meta.TotalRecords != 1 || meta.CurrentPage != 1 {
		t.Errorf("unexpected meta: %+v", meta)
	}

	repoErr := &mockRepository{err: errors.New("db error")}
	svcErr := master.NewService(repoErr, logger.New())
	_, _, err = svcErr.DaftarICD10(context.Background(), filter)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMasterService_DaftarICD9(t *testing.T) {
	repo := &mockRepository{
		icd9Data: []master.ICD9{
			{ItemMaster: master.ItemMaster{Kode: "87.03", Nama: "Computerized axial tomography of head"}},
		},
		icd9Total: 1,
	}
	svc := master.NewService(repo, logger.New())

	filter := master.FilterMasterICD{Keyword: "tomography", Page: 1, Limit: 20}
	data, meta, err := svc.DaftarICD9(context.Background(), filter)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data) != 1 || data[0].Kode != "87.03" {
		t.Errorf("unexpected ICD-9 data: %+v", data)
	}
	if meta.TotalRecords != 1 || meta.CurrentPage != 1 {
		t.Errorf("unexpected meta: %+v", meta)
	}

	repoErr := &mockRepository{err: errors.New("db error")}
	svcErr := master.NewService(repoErr, logger.New())
	_, _, err = svcErr.DaftarICD9(context.Background(), filter)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMasterService_CekKeberadaanICD(t *testing.T) {
	repo := &mockRepository{
		icd10Data: []master.ICD10{
			{ItemMaster: master.ItemMaster{Kode: "I63.9", Nama: "Cerebral infarction"}},
		},
		icd9Data: []master.ICD9{
			{ItemMaster: master.ItemMaster{Kode: "87.03", Nama: "CT Head"}},
		},
		checkICD10Map: map[string]bool{"I63.9": true, "A00": false},
		checkICD9Map:  map[string]bool{"87.03": true, "99.99": false},
	}
	svc := master.NewService(repo, logger.New())

	res10, err := svc.CekKeberadaanICD10(context.Background(), []string{"I63.9", "A00"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res10["I63.9"] || res10["A00"] {
		t.Errorf("unexpected res10: %+v", res10)
	}

	res9, err := svc.CekKeberadaanICD9(context.Background(), []string{"87.03", "99.99"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res9["87.03"] || res9["99.99"] {
		t.Errorf("unexpected res9: %+v", res9)
	}

	repoErr := &mockRepository{err: errors.New("db error")}
	svcErr := master.NewService(repoErr, logger.New())
	_, err = svcErr.CekKeberadaanICD10(context.Background(), []string{"I63.9"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	_, err = svcErr.CekKeberadaanICD9(context.Background(), []string{"87.03"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMasterService_SyncICD(t *testing.T) {
	repo := &mockRepository{
		icd10Data: []master.ICD10{
			{ItemMaster: master.ItemMaster{Kode: "I63.9", Nama: "Cerebral infarction"}},
		},
		icd9Data: []master.ICD9{
			{ItemMaster: master.ItemMaster{Kode: "87.03", Nama: "CT Head"}},
		},
	}
	svc := master.NewService(repo, logger.New())

	res, err := svc.SyncICD(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.TotalICD10 != 1 || res.TotalICD9 != 1 {
		t.Errorf("unexpected sync result: %+v", res)
	}

	repo.err = errors.New("db error")
	_, err = svc.SyncICD(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

type concurrencyMockRepo struct {
	mockRepository
	mu              sync.RWMutex
	fetchICD10Calls atomic.Int32
	fetchICD9Calls  atomic.Int32
	onFetchICD10    func()
}

func (m *concurrencyMockRepo) FetchAllICD10(ctx context.Context) ([]master.ICD10, error) {
	m.fetchICD10Calls.Add(1)
	if m.onFetchICD10 != nil {
		m.onFetchICD10()
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	out := make([]master.ICD10, len(m.icd10Data))
	copy(out, m.icd10Data)
	return out, nil
}

func (m *concurrencyMockRepo) FetchAllICD9(ctx context.Context) ([]master.ICD9, error) {
	m.fetchICD9Calls.Add(1)
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.err != nil {
		return nil, m.err
	}
	out := make([]master.ICD9, len(m.icd9Data))
	copy(out, m.icd9Data)
	return out, nil
}

func TestMasterService_SyncICDVisibleToCekKeberadaanICD(t *testing.T) {
	repo := &concurrencyMockRepo{
		mockRepository: mockRepository{
			icd10Data: []master.ICD10{{ItemMaster: master.ItemMaster{Kode: "A09", Nama: "Gastroenteritis"}}},
			icd9Data:  []master.ICD9{{ItemMaster: master.ItemMaster{Kode: "87.03", Nama: "CT Head"}}},
		},
	}
	sharedSvc := master.NewService(repo, logger.New())

	res10Before, err := sharedSvc.CekKeberadaanICD10(context.Background(), []string{"A09", "U07.1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res10Before["A09"] || res10Before["U07.1"] {
		t.Fatalf("unexpected initial ICD10 map: %+v", res10Before)
	}

	repo.mu.Lock()
	repo.icd10Data = append(repo.icd10Data, master.ICD10{ItemMaster: master.ItemMaster{Kode: "U07.1", Nama: "COVID-19"}})
	repo.icd9Data = append(repo.icd9Data, master.ICD9{ItemMaster: master.ItemMaster{Kode: "93.90", Nama: "Non-invasive mechanical ventilation"}})
	repo.mu.Unlock()

	if _, err := sharedSvc.SyncICD(context.Background()); err != nil {
		t.Fatalf("unexpected sync error: %v", err)
	}

	res10After, err := sharedSvc.CekKeberadaanICD10(context.Background(), []string{"U07.1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res10After["U07.1"] {
		t.Fatalf("expected U07.1 to be visible immediately after SyncICD, got %+v", res10After)
	}

	res9After, err := sharedSvc.CekKeberadaanICD9(context.Background(), []string{"93.90"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res9After["93.90"] {
		t.Fatalf("expected 93.90 to be visible immediately after SyncICD, got %+v", res9After)
	}
}

func TestMasterService_ICDKodeNormalization_CacheAndSQLFallback(t *testing.T) {
	t.Run("InMemoryCache UPPER(TRIM)", func(t *testing.T) {
		repo := &mockRepository{
			icd10Data: []master.ICD10{
				{ItemMaster: master.ItemMaster{Kode: "a09 ", Nama: "Gastroenteritis"}},
			},
			icd9Data: []master.ICD9{
				{ItemMaster: master.ItemMaster{Kode: " 87.03 ", Nama: "CT Head"}},
			},
		}
		svc := master.NewService(repo, logger.New())

		got10, err := svc.CekKeberadaanICD10(context.Background(), []string{"A09", "a09", " A09 ", "   "})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got10["A09"] || !got10["a09"] || !got10[" A09 "] || got10["   "] {
			t.Fatalf("unexpected normalized ICD-10 result: %+v", got10)
		}

		got9, err := svc.CekKeberadaanICD9(context.Background(), []string{"87.03", " 87.03", ""})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got9["87.03"] || !got9[" 87.03"] || got9[""] {
			t.Fatalf("unexpected normalized ICD-9 result: %+v", got9)
		}
	})

	t.Run("SQLRepositoryFallback UPPER(TRIM)", func(t *testing.T) {
		db := newStubMasterDB(t, func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error) {
			if strings.Contains(query, "FROM penyakit") {
				if len(args) != 3 || args[0].Value != "a09" || args[1].Value != "A09" || args[2].Value != "" {
					t.Fatalf("unexpected ICD-10 query args: %+v", args)
				}
				return []string{"kd_penyakit"}, [][]driver.Value{{"A09 "}}, nil
			}
			if strings.Contains(query, "FROM icd9") {
				if len(args) != 2 || args[0].Value != "87.03" || args[1].Value != "" {
					t.Fatalf("unexpected ICD-9 query args: %+v", args)
				}
				return []string{"kode"}, [][]driver.Value{{" 87.03 "}}, nil
			}
			return nil, nil, errors.New("unexpected query: " + query)
		})

		sqlRepo := master.NewRepository(db)

		got10, err := sqlRepo.CekKeberadaanICD10(context.Background(), []string{"a09", " A09 ", "   "})
		if err != nil {
			t.Fatalf("unexpected ICD-10 repo error: %v", err)
		}
		if !got10["a09"] || !got10[" A09 "] || got10["   "] {
			t.Fatalf("unexpected SQL fallback ICD-10 map: %+v", got10)
		}

		got9, err := sqlRepo.CekKeberadaanICD9(context.Background(), []string{" 87.03", " "})
		if err != nil {
			t.Fatalf("unexpected ICD-9 repo error: %v", err)
		}
		if !got9[" 87.03"] || got9[" "] {
			t.Fatalf("unexpected SQL fallback ICD-9 map: %+v", got9)
		}
	})
}

type stubMasterQueryFunc func(query string, args []driver.NamedValue) ([]string, [][]driver.Value, error)

type stubMasterConnector struct {
	fn stubMasterQueryFunc
}

func (c *stubMasterConnector) Connect(context.Context) (driver.Conn, error) {
	return &stubMasterConn{fn: c.fn}, nil
}
func (c *stubMasterConnector) Driver() driver.Driver { return nil }

type stubMasterConn struct {
	fn stubMasterQueryFunc
}

func (c *stubMasterConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unsupported") }
func (c *stubMasterConn) Close() error                        { return nil }
func (c *stubMasterConn) Begin() (driver.Tx, error)           { return nil, errors.New("unsupported") }
func (c *stubMasterConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	cols, rows, err := c.fn(query, args)
	if err != nil {
		return nil, err
	}
	return &stubMasterRows{cols: cols, rows: rows}, nil
}

type stubMasterRows struct {
	cols []string
	rows [][]driver.Value
	idx  int
}

func (r *stubMasterRows) Columns() []string { return r.cols }
func (r *stubMasterRows) Close() error      { return nil }
func (r *stubMasterRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.idx])
	r.idx++
	return nil
}

func newStubMasterDB(t *testing.T, fn stubMasterQueryFunc) *sql.DB {
	t.Helper()
	db := sql.OpenDB(&stubMasterConnector{fn: fn})
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestMasterService_ColdLoadSingleflight(t *testing.T) {
	repo := &concurrencyMockRepo{
		mockRepository: mockRepository{
			icd10Data: []master.ICD10{{ItemMaster: master.ItemMaster{Kode: "I10", Nama: "Essential hypertension"}}},
			icd9Data:  []master.ICD9{{ItemMaster: master.ItemMaster{Kode: "89.52", Nama: "Electrocardiogram"}}},
		},
		onFetchICD10: func() {
			time.Sleep(25 * time.Millisecond)
		},
	}
	svc := master.NewService(repo, logger.New())

	const workers = 20
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(idx int) {
			defer wg.Done()
			<-start
			if idx%2 == 0 {
				res, err := svc.CekKeberadaanICD10(context.Background(), []string{"I10"})
				if err != nil || !res["I10"] {
					t.Errorf("worker %d unexpected ICD10 result: %v, err=%v", idx, res, err)
				}
			} else {
				res, err := svc.CekKeberadaanICD9(context.Background(), []string{"89.52"})
				if err != nil || !res["89.52"] {
					t.Errorf("worker %d unexpected ICD9 result: %v, err=%v", idx, res, err)
				}
			}
		}(i)
	}
	close(start)
	wg.Wait()

	if got := repo.fetchICD10Calls.Load(); got != 1 {
		t.Fatalf("expected exactly 1 cold-load FetchAllICD10 call, got %d", got)
	}
	if got := repo.fetchICD9Calls.Load(); got != 1 {
		t.Fatalf("expected exactly 1 cold-load FetchAllICD9 call, got %d", got)
	}
}

func TestMasterService_AutoRefreshOnCekKeberadaanICD_AntiStampede(t *testing.T) {
	refreshEntered := make(chan struct{}, 1)
	releaseRefresh := make(chan struct{})
	var gateActive atomic.Bool

	repo := &concurrencyMockRepo{
		mockRepository: mockRepository{
			icd10Data: []master.ICD10{{ItemMaster: master.ItemMaster{Kode: "A09", Nama: "Gastroenteritis"}}},
			icd9Data:  []master.ICD9{{ItemMaster: master.ItemMaster{Kode: "87.03", Nama: "CT Head"}}},
		},
	}
	repo.onFetchICD10 = func() {
		if gateActive.Load() {
			select {
			case refreshEntered <- struct{}{}:
			default:
			}
			<-releaseRefresh
		}
	}

	svc := master.NewServiceWithTTL(repo, logger.New(), 15*time.Millisecond)

	if _, err := svc.CekKeberadaanICD10(context.Background(), []string{"A09"}); err != nil {
		t.Fatalf("unexpected initial error: %v", err)
	}
	if got := repo.fetchICD10Calls.Load(); got != 1 {
		t.Fatalf("expected 1 initial FetchAllICD10 call, got %d", got)
	}

	repo.mu.Lock()
	repo.icd10Data = append(repo.icd10Data, master.ICD10{ItemMaster: master.ItemMaster{Kode: "J06.9", Nama: "ISPA"}})
	repo.icd9Data = append(repo.icd9Data, master.ICD9{ItemMaster: master.ItemMaster{Kode: "96.04", Nama: "Endotracheal intubation"}})
	repo.mu.Unlock()
	gateActive.Store(true)
	time.Sleep(30 * time.Millisecond)

	var wg sync.WaitGroup
	wg.Add(16)
	for i := 0; i < 16; i++ {
		go func(idx int) {
			defer wg.Done()
			if idx%2 == 0 {
				_, _ = svc.CekKeberadaanICD10(context.Background(), []string{"A09"})
			} else {
				_, _ = svc.CekKeberadaanICD9(context.Background(), []string{"87.03"})
			}
		}(i)
	}

	select {
	case <-refreshEntered:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for background refresh to trigger from CekKeberadaanICD")
	}
	wg.Wait()
	close(releaseRefresh)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		res10, _ := svc.CekKeberadaanICD10(context.Background(), []string{"J06.9"})
		res9, _ := svc.CekKeberadaanICD9(context.Background(), []string{"96.04"})
		if res10["J06.9"] && res9["96.04"] {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if got := repo.fetchICD10Calls.Load(); got != 2 {
		t.Fatalf("expected exactly 2 FetchAllICD10 calls (1 cold + 1 single background refresh), got %d", got)
	}
	res10, err := svc.CekKeberadaanICD10(context.Background(), []string{"J06.9"})
	if err != nil || !res10["J06.9"] {
		t.Fatalf("expected J06.9 to be loaded after TTL refresh, got %+v (err=%v)", res10, err)
	}
	res9, err := svc.CekKeberadaanICD9(context.Background(), []string{"96.04"})
	if err != nil || !res9["96.04"] {
		t.Fatalf("expected 96.04 to be loaded after TTL refresh, got %+v (err=%v)", res9, err)
	}
}
