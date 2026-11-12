package master

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"erm-dokter/internal/pkg/logger"
	"erm-dokter/internal/shared"
)

const defaultICDCacheTTL = 24 * time.Hour

type Service interface {
	DaftarPenjamin(ctx context.Context) ([]Penjamin, error)
	DaftarDepo(ctx context.Context) ([]Depo, error)
	DaftarPoliklinik(ctx context.Context) ([]Poliklinik, error)
	DaftarBangsal(ctx context.Context) ([]Bangsal, error)
	DaftarKelas(ctx context.Context) ([]KelasKamar, error)
	DaftarICD10(ctx context.Context, filter FilterMasterICD) ([]ICD10, shared.PaginationMeta, error)
	DaftarICD9(ctx context.Context, filter FilterMasterICD) ([]ICD9, shared.PaginationMeta, error)
	CekKeberadaanICD10(ctx context.Context, listKode []string) (map[string]bool, error)
	CekKeberadaanICD9(ctx context.Context, listKode []string) (map[string]bool, error)
	SyncICD(ctx context.Context) (*SyncICDResult, error)
}

type service struct {
	repo       Repository
	log        *logger.Logger
	cache      *ICDCache
	cacheTTL   time.Duration
	loadMu     sync.Mutex
	refreshing atomic.Bool
}

func NewService(repo Repository, log *logger.Logger) Service {
	return NewServiceWithTTL(repo, log, defaultICDCacheTTL)
}

func NewServiceWithTTL(repo Repository, log *logger.Logger, cacheTTL time.Duration) Service {
	if cacheTTL <= 0 {
		cacheTTL = defaultICDCacheTTL
	}
	return &service{
		repo:     repo,
		log:      log,
		cache:    NewICDCache(),
		cacheTTL: cacheTTL,
	}
}

func (s *service) DaftarPenjamin(ctx context.Context) ([]Penjamin, error) {
	data, err := s.repo.DaftarPenjamin(ctx)
	if err != nil {
		s.log.Error("Gagal mengambil daftar penjamin: %v", err)
		return nil, err
	}
	return data, nil
}

func (s *service) DaftarDepo(ctx context.Context) ([]Depo, error) {
	data, err := s.repo.DaftarDepo(ctx)
	if err != nil {
		s.log.Error("Gagal mengambil daftar depo: %v", err)
		return nil, err
	}
	return data, nil
}

func (s *service) DaftarPoliklinik(ctx context.Context) ([]Poliklinik, error) {
	data, err := s.repo.DaftarPoliklinik(ctx)
	if err != nil {
		s.log.Error("Gagal mengambil daftar poliklinik: %v", err)
		return nil, err
	}
	return data, nil
}

func (s *service) DaftarBangsal(ctx context.Context) ([]Bangsal, error) {
	data, err := s.repo.DaftarBangsal(ctx)
	if err != nil {
		s.log.Error("Gagal mengambil daftar bangsal: %v", err)
		return nil, err
	}
	return data, nil
}

func (s *service) DaftarKelas(ctx context.Context) ([]KelasKamar, error) {
	data, err := s.repo.DaftarKelas(ctx)
	if err != nil {
		s.log.Error("Gagal mengambil daftar kelas kamar: %v", err)
		return nil, err
	}
	return data, nil
}

func (s *service) ensureCacheLoaded(ctx context.Context) {
	if s.cache.IsLoaded() {
		return
	}

	s.loadMu.Lock()
	defer s.loadMu.Unlock()

	if s.cache.IsLoaded() {
		return
	}

	icd10, err10 := s.repo.FetchAllICD10(ctx)
	if err10 != nil {
		s.log.Warn("Gagal memuat awal cache ICD-10: %v", err10)
		return
	}
	icd9, err9 := s.repo.FetchAllICD9(ctx)
	if err9 != nil {
		s.log.Warn("Gagal memuat awal cache ICD-9: %v", err9)
		return
	}
	s.cache.Load(icd10, icd9)
	s.log.Info("In-Memory Cache ICD berhasil dimuat (%d ICD-10, %d ICD-9)", len(icd10), len(icd9))
}

func (s *service) triggerBackgroundRefresh() {
	if !s.refreshing.CompareAndSwap(false, true) {
		return
	}

	go func() {
		defer s.refreshing.Store(false)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		icd10, err10 := s.repo.FetchAllICD10(ctx)
		if err10 != nil {
			s.log.Warn("Gagal auto-refresh cache ICD-10: %v", err10)
			return
		}
		icd9, err9 := s.repo.FetchAllICD9(ctx)
		if err9 != nil {
			s.log.Warn("Gagal auto-refresh cache ICD-9: %v", err9)
			return
		}
		s.cache.Load(icd10, icd9)
		s.log.Info("Auto-refresh In-Memory Cache ICD selesai (%d ICD-10, %d ICD-9)", len(icd10), len(icd9))
	}()
}

func (s *service) DaftarICD10(ctx context.Context, filter FilterMasterICD) ([]ICD10, shared.PaginationMeta, error) {
	if !s.cache.IsLoaded() {
		s.ensureCacheLoaded(ctx)
	}

	if s.cache.IsLoaded() {
		if s.cache.NeedsRefresh(s.cacheTTL) {
			s.triggerBackgroundRefresh()
		}
		data, total := s.cache.SearchICD10(filter)
		meta := shared.NewPaginationMeta(total, filter.Page, filter.Limit)
		return data, meta, nil
	}

	data, total, err := s.repo.DaftarICD10(ctx, filter)
	if err != nil {
		s.log.Error("Gagal mengambil daftar master ICD-10: %v", err)
		return nil, shared.PaginationMeta{}, err
	}
	meta := shared.NewPaginationMeta(total, filter.Page, filter.Limit)
	return data, meta, nil
}

func (s *service) DaftarICD9(ctx context.Context, filter FilterMasterICD) ([]ICD9, shared.PaginationMeta, error) {
	if !s.cache.IsLoaded() {
		s.ensureCacheLoaded(ctx)
	}

	if s.cache.IsLoaded() {
		if s.cache.NeedsRefresh(s.cacheTTL) {
			s.triggerBackgroundRefresh()
		}
		data, total := s.cache.SearchICD9(filter)
		meta := shared.NewPaginationMeta(total, filter.Page, filter.Limit)
		return data, meta, nil
	}

	data, total, err := s.repo.DaftarICD9(ctx, filter)
	if err != nil {
		s.log.Error("Gagal mengambil daftar master ICD-9: %v", err)
		return nil, shared.PaginationMeta{}, err
	}
	meta := shared.NewPaginationMeta(total, filter.Page, filter.Limit)
	return data, meta, nil
}

func (s *service) CekKeberadaanICD10(ctx context.Context, listKode []string) (map[string]bool, error) {
	if !s.cache.IsLoaded() {
		s.ensureCacheLoaded(ctx)
	}

	if s.cache.IsLoaded() {
		if s.cache.NeedsRefresh(s.cacheTTL) {
			s.triggerBackgroundRefresh()
		}
		return s.cache.CheckICD10(listKode), nil
	}

	result, err := s.repo.CekKeberadaanICD10(ctx, listKode)
	if err != nil {
		s.log.Error("Gagal mengecek keberadaan ICD-10: %v", err)
		return nil, err
	}
	return result, nil
}

func (s *service) CekKeberadaanICD9(ctx context.Context, listKode []string) (map[string]bool, error) {
	if !s.cache.IsLoaded() {
		s.ensureCacheLoaded(ctx)
	}

	if s.cache.IsLoaded() {
		if s.cache.NeedsRefresh(s.cacheTTL) {
			s.triggerBackgroundRefresh()
		}
		return s.cache.CheckICD9(listKode), nil
	}

	result, err := s.repo.CekKeberadaanICD9(ctx, listKode)
	if err != nil {
		s.log.Error("Gagal mengecek keberadaan ICD-9: %v", err)
		return nil, err
	}
	return result, nil
}

func (s *service) SyncICD(ctx context.Context) (*SyncICDResult, error) {
	icd10, err := s.repo.FetchAllICD10(ctx)
	if err != nil {
		s.log.Error("Gagal sinkronisasi ICD-10: %v", err)
		return nil, err
	}

	icd9, err := s.repo.FetchAllICD9(ctx)
	if err != nil {
		s.log.Error("Gagal sinkronisasi ICD-9: %v", err)
		return nil, err
	}

	s.cache.Load(icd10, icd9)
	s.log.Info("Manual sinkronisasi cache ICD berhasil (%d ICD-10, %d ICD-9)", len(icd10), len(icd9))

	return &SyncICDResult{
		TotalICD10: len(icd10),
		TotalICD9:  len(icd9),
		SyncedAt:   s.cache.LastSync(),
	}, nil
}
