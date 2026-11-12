package master

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"erm-dokter/internal/shared"
)

type Repository interface {
	DaftarPenjamin(ctx context.Context) ([]Penjamin, error)
	DaftarDepo(ctx context.Context) ([]Depo, error)
	DaftarPoliklinik(ctx context.Context) ([]Poliklinik, error)
	DaftarBangsal(ctx context.Context) ([]Bangsal, error)
	DaftarKelas(ctx context.Context) ([]KelasKamar, error)
	DaftarICD10(ctx context.Context, filter FilterMasterICD) ([]ICD10, int, error)
	DaftarICD9(ctx context.Context, filter FilterMasterICD) ([]ICD9, int, error)
	FetchAllICD10(ctx context.Context) ([]ICD10, error)
	FetchAllICD9(ctx context.Context) ([]ICD9, error)
	CekKeberadaanICD10(ctx context.Context, listKode []string) (map[string]bool, error)
	CekKeberadaanICD9(ctx context.Context, listKode []string) (map[string]bool, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) DaftarPenjamin(ctx context.Context) ([]Penjamin, error) {
	query := `SELECT kd_pj AS kode, png_jawab AS nama FROM penjab WHERE status = '1' ORDER BY png_jawab ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]Penjamin, 0)
	for rows.Next() {
		var p Penjamin
		if err := rows.Scan(&p.Kode, &p.Nama); err != nil {
			return nil, err
		}
		list = append(list, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *repository) DaftarDepo(ctx context.Context) ([]Depo, error) {
	query := fmt.Sprintf(`SELECT kd_bangsal AS kode, nm_bangsal AS nama 
		FROM bangsal 
		WHERE %s 
		ORDER BY nm_bangsal ASC`, shared.InClauseDepoFarmasi("kd_bangsal"))
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]Depo, 0)
	for rows.Next() {
		var d Depo
		if err := rows.Scan(&d.Kode, &d.Nama); err != nil {
			return nil, err
		}
		list = append(list, d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *repository) DaftarPoliklinik(ctx context.Context) ([]Poliklinik, error) {
	query := `SELECT kd_poli AS kode, nm_poli AS nama FROM poliklinik WHERE status = '1' ORDER BY nm_poli ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]Poliklinik, 0)
	for rows.Next() {
		var p Poliklinik
		if err := rows.Scan(&p.Kode, &p.Nama); err != nil {
			return nil, err
		}
		list = append(list, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *repository) DaftarBangsal(ctx context.Context) ([]Bangsal, error) {
	query := `
		SELECT DISTINCT b.kd_bangsal AS kode, b.nm_bangsal AS nama 
		FROM bangsal b 
		INNER JOIN kamar k ON b.kd_bangsal = k.kd_bangsal 
		WHERE b.status = '1' AND k.statusdata = '1' 
		ORDER BY b.nm_bangsal ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]Bangsal, 0)
	for rows.Next() {
		var b Bangsal
		if err := rows.Scan(&b.Kode, &b.Nama); err != nil {
			return nil, err
		}
		list = append(list, b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *repository) DaftarKelas(ctx context.Context) ([]KelasKamar, error) {
	query := `
		SELECT DISTINCT k.kelas AS kode, k.kelas AS nama 
		FROM kamar k 
		INNER JOIN bangsal b ON k.kd_bangsal = b.kd_bangsal 
		WHERE k.statusdata = '1' AND b.status = '1' AND k.kelas <> '' 
		ORDER BY k.kelas ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]KelasKamar, 0)
	for rows.Next() {
		var k KelasKamar
		if err := rows.Scan(&k.Kode, &k.Nama); err != nil {
			return nil, err
		}
		list = append(list, k)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *repository) DaftarICD10(ctx context.Context, filter FilterMasterICD) ([]ICD10, int, error) {
	baseWhere := "WHERE tampil = 'YA' AND kd_penyakit NOT IN ('-', '00')"
	var args []any
	var countArgs []any

	whereClause := baseWhere
	if filter.Keyword != "" {
		whereClause += " AND (kd_penyakit LIKE ? OR nm_penyakit LIKE ?)"
		pattern := "%" + filter.Keyword + "%"
		args = append(args, pattern, pattern)
		countArgs = append(countArgs, pattern, pattern)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM penyakit %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []ICD10{}, 0, nil
	}

	query := fmt.Sprintf(`SELECT kd_penyakit AS kode, nm_penyakit AS nama 
		FROM penyakit 
		%s 
		ORDER BY kd_penyakit ASC 
		LIMIT ? OFFSET ?`, whereClause)
	args = append(args, filter.Limit, filter.Offset())

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]ICD10, 0, filter.Limit)
	for rows.Next() {
		var item ICD10
		if err := rows.Scan(&item.Kode, &item.Nama); err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *repository) DaftarICD9(ctx context.Context, filter FilterMasterICD) ([]ICD9, int, error) {
	whereClause := ""
	var args []any
	var countArgs []any

	if filter.Keyword != "" {
		whereClause = "WHERE (kode LIKE ? OR deskripsi_panjang LIKE ?)"
		pattern := "%" + filter.Keyword + "%"
		args = append(args, pattern, pattern)
		countArgs = append(countArgs, pattern, pattern)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM icd9 %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []ICD9{}, 0, nil
	}

	query := fmt.Sprintf(`SELECT kode, deskripsi_panjang AS nama 
		FROM icd9 
		%s 
		ORDER BY kode ASC 
		LIMIT ? OFFSET ?`, whereClause)
	args = append(args, filter.Limit, filter.Offset())

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]ICD9, 0, filter.Limit)
	for rows.Next() {
		var item ICD9
		if err := rows.Scan(&item.Kode, &item.Nama); err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *repository) CekKeberadaanICD10(ctx context.Context, listKode []string) (map[string]bool, error) {
	result := make(map[string]bool, len(listKode))
	if len(listKode) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(listKode))
	args := make([]any, len(listKode))
	for i, kode := range listKode {
		result[kode] = false
		placeholders[i] = "?"
		args[i] = strings.TrimSpace(kode)
	}

	query := fmt.Sprintf(`SELECT kd_penyakit 
		FROM penyakit 
		WHERE tampil = 'YA' AND kd_penyakit NOT IN ('-', '00') AND kd_penyakit IN (%s)`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	found := make(map[string]bool, len(listKode))
	for rows.Next() {
		var kode string
		if err := rows.Scan(&kode); err != nil {
			return nil, err
		}
		if key := normalizeICDKode(kode); key != "" {
			found[key] = true
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, kode := range listKode {
		key := normalizeICDKode(kode)
		result[kode] = key != "" && found[key]
	}

	return result, nil
}

func (r *repository) CekKeberadaanICD9(ctx context.Context, listKode []string) (map[string]bool, error) {
	result := make(map[string]bool, len(listKode))
	if len(listKode) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(listKode))
	args := make([]any, len(listKode))
	for i, kode := range listKode {
		result[kode] = false
		placeholders[i] = "?"
		args[i] = strings.TrimSpace(kode)
	}

	query := fmt.Sprintf(`SELECT kode 
		FROM icd9 
		WHERE kode IN (%s)`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	found := make(map[string]bool, len(listKode))
	for rows.Next() {
		var kode string
		if err := rows.Scan(&kode); err != nil {
			return nil, err
		}
		if key := normalizeICDKode(kode); key != "" {
			found[key] = true
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, kode := range listKode {
		key := normalizeICDKode(kode)
		result[kode] = key != "" && found[key]
	}

	return result, nil
}

func (r *repository) FetchAllICD10(ctx context.Context) ([]ICD10, error) {
	query := `SELECT kd_penyakit AS kode, nm_penyakit AS nama 
		FROM penyakit 
		WHERE tampil = 'YA' AND kd_penyakit NOT IN ('-', '00') 
		ORDER BY kd_penyakit ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]ICD10, 0, 15000)
	for rows.Next() {
		var item ICD10
		if err := rows.Scan(&item.Kode, &item.Nama); err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *repository) FetchAllICD9(ctx context.Context) ([]ICD9, error) {
	query := `SELECT kode, deskripsi_panjang AS nama 
		FROM icd9 
		ORDER BY kode ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]ICD9, 0, 4000)
	for rows.Next() {
		var item ICD9
		if err := rows.Scan(&item.Kode, &item.Nama); err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}
