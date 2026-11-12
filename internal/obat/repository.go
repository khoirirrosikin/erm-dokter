package obat

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"erm-dokter/internal/shared"
)

type Repository interface {
	DaftarObat(ctx context.Context, filter FilterDaftarObat) ([]Obat, int64, error)
	DetailObat(ctx context.Context, kodeObat string) (*Obat, error)
	DaftarJenis(ctx context.Context) ([]JenisObat, error)
	DaftarGolongan(ctx context.Context) ([]GolonganObat, error)
	DaftarKategori(ctx context.Context) ([]KategoriObat, error)
	CekKeberadaanObat(ctx context.Context, listKodeObat []string) (map[string]bool, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func buildOrderClause(orderBy, sortOrder string, hasDepo bool) string {
	switch orderBy {
	case "stok":
		if hasDepo {
			return fmt.Sprintf("gdb.stok %s, dtb.nama_brng ASC, dtb.kode_brng ASC", sortOrder)
		}
		return fmt.Sprintf("stok %s, dtb.nama_brng ASC, dtb.kode_brng ASC", sortOrder)
	case "nama_depo":
		if hasDepo {
			return fmt.Sprintf("bg.nm_bangsal %s, dtb.nama_brng ASC, dtb.kode_brng ASC", sortOrder)
		}
		return fmt.Sprintf("dtb.nama_brng %s, dtb.kode_brng ASC", sortOrder)
	case "nama_jenis":
		return fmt.Sprintf("jn.nama %s, dtb.nama_brng ASC, dtb.kode_brng ASC", sortOrder)
	case "nama_golongan":
		return fmt.Sprintf("gb.nama %s, dtb.nama_brng ASC, dtb.kode_brng ASC", sortOrder)
	case "nama_kategori":
		return fmt.Sprintf("kb.nama %s, dtb.nama_brng ASC, dtb.kode_brng ASC", sortOrder)
	default:
		return fmt.Sprintf("dtb.nama_brng %s, dtb.kode_brng ASC", sortOrder)
	}
}

func (r *repository) DaftarObat(ctx context.Context, filter FilterDaftarObat) ([]Obat, int64, error) {
	orderBy := strings.ToLower(strings.TrimSpace(filter.OrderBy))
	if orderBy == "" {
		orderBy = "nama_obat"
	}
	sortOrder := "ASC"
	if strings.EqualFold(strings.TrimSpace(filter.SortOrder), "DESC") {
		sortOrder = "DESC"
	}

	var baseFrom string
	var selectCols string
	var conditions []string
	var args []any

	if filter.Depo != "" {
		baseFrom = `FROM databarang dtb 
			INNER JOIN gudangbarang gdb ON dtb.kode_brng = gdb.kode_brng 
			INNER JOIN bangsal bg ON gdb.kd_bangsal = bg.kd_bangsal
			INNER JOIN kategori_barang kb ON dtb.kode_kategori = kb.kode 
			INNER JOIN golongan_barang gb ON dtb.kode_golongan = gb.kode 
			INNER JOIN jenis jn ON dtb.kdjns = jn.kdjns
			WHERE dtb.status = '1'`

		selectCols = `SELECT 
			dtb.kode_brng AS kode_obat,
			dtb.nama_brng AS nama_obat,
			dtb.letak_barang AS komposisi,
			dtb.ralan AS harga,
			dtb.kode_sat AS satuan,
			gdb.stok,
			dtb.kapasitas,
			gdb.kd_bangsal AS kode_depo,
			bg.nm_bangsal AS nama_depo,
			dtb.kdjns AS kode_jenis,
			jn.nama AS nama_jenis,
			dtb.kode_golongan AS kode_golongan,
			gb.nama AS nama_golongan,
			dtb.kode_kategori AS kode_kategori,
			kb.nama AS nama_kategori `

		conditions = append(conditions, "gdb.kd_bangsal = ?")
		args = append(args, filter.Depo)
	} else {
		baseFrom = `FROM databarang dtb 
			INNER JOIN kategori_barang kb ON dtb.kode_kategori = kb.kode 
			INNER JOIN golongan_barang gb ON dtb.kode_golongan = gb.kode 
			INNER JOIN jenis jn ON dtb.kdjns = jn.kdjns
			WHERE dtb.status = '1'`

		stokExpr := "0 AS stok"
		if orderBy == "stok" {
			stokExpr = "COALESCE((SELECT SUM(gdb.stok) FROM gudangbarang gdb WHERE gdb.kode_brng = dtb.kode_brng AND " + shared.InClauseDepoFarmasi("gdb.kd_bangsal") + "), 0) AS stok"
		}

		selectCols = fmt.Sprintf(`SELECT 
			dtb.kode_brng AS kode_obat,
			dtb.nama_brng AS nama_obat,
			dtb.letak_barang AS komposisi,
			dtb.ralan AS harga,
			dtb.kode_sat AS satuan,
			%s,
			dtb.kapasitas,
			'' AS kode_depo,
			'Semua Depo' AS nama_depo,
			dtb.kdjns AS kode_jenis,
			jn.nama AS nama_jenis,
			dtb.kode_golongan AS kode_golongan,
			gb.nama AS nama_golongan,
			dtb.kode_kategori AS kode_kategori,
			kb.nama AS nama_kategori `, stokExpr)
	}

	if filter.Jenis != "" {
		conditions = append(conditions, "dtb.kdjns = ?")
		args = append(args, filter.Jenis)
	}

	if filter.Golongan != "" {
		conditions = append(conditions, "dtb.kode_golongan = ?")
		args = append(args, filter.Golongan)
	}

	if filter.Kategori != "" {
		conditions = append(conditions, "dtb.kode_kategori = ?")
		args = append(args, filter.Kategori)
	}

	if filter.Keyword != "" {
		keywordPattern := filter.Keyword + "%"
		conditions = append(conditions, "(dtb.nama_brng LIKE ? OR dtb.letak_barang LIKE ?)")
		args = append(args, keywordPattern, keywordPattern)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	countFrom := r.buildCountFrom(filter)
	countWhere := r.buildCountWhere(filter)
	countQuery := "SELECT COUNT(*) " + countFrom + countWhere
	countArgs := r.buildCountArgs(filter)

	orderClause := " ORDER BY " + buildOrderClause(orderBy, sortOrder, filter.Depo != "")
	selectQuery := selectCols + baseFrom + whereClause + orderClause + " LIMIT ? OFFSET ?"
	dataArgs := append(args, filter.Limit, filter.Offset())

	type countResult struct {
		total int64
		err   error
	}
	ch := make(chan countResult, 1)

	go func() {
		var total int64
		err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
		ch <- countResult{total, err}
	}()

	rows, err := r.db.QueryContext(ctx, selectQuery, dataArgs...)
	if err != nil {
		<-ch
		return nil, 0, err
	}
	defer rows.Close()

	listObat := make([]Obat, 0)
	for rows.Next() {
		var o Obat
		if err := rows.Scan(
			&o.KodeObat,
			&o.NamaObat,
			&o.Komposisi,
			&o.Harga,
			&o.Satuan,
			&o.Stok,
			&o.Kapasitas,
			&o.KodeDepo,
			&o.NamaDepo,
			&o.KodeJenis,
			&o.NamaJenis,
			&o.KodeGolongan,
			&o.NamaGolongan,
			&o.KodeKategori,
			&o.NamaKategori,
		); err != nil {
			<-ch
			return nil, 0, err
		}
		listObat = append(listObat, o)
	}

	if err := rows.Err(); err != nil {
		<-ch
		return nil, 0, err
	}

	cr := <-ch
	if cr.err != nil {
		return nil, 0, cr.err
	}
	if filter.Depo == "" && orderBy != "stok" && len(listObat) > 0 {
		if err := r.hydrateStok(ctx, listObat); err != nil {
			return nil, 0, err
		}
	}

	return listObat, cr.total, nil
}

func (r *repository) buildCountFrom(filter FilterDaftarObat) string {
	from := "FROM databarang dtb "

	if filter.Depo != "" {
		from += "INNER JOIN gudangbarang gdb ON dtb.kode_brng = gdb.kode_brng "
	}
	if filter.Jenis != "" {
		from += "INNER JOIN jenis jn ON dtb.kdjns = jn.kdjns "
	}
	if filter.Golongan != "" {
		from += "INNER JOIN golongan_barang gb ON dtb.kode_golongan = gb.kode "
	}
	if filter.Kategori != "" {
		from += "INNER JOIN kategori_barang kb ON dtb.kode_kategori = kb.kode "
	}

	return from
}

func (r *repository) buildCountWhere(filter FilterDaftarObat) string {
	conditions := []string{"dtb.status = '1'"}

	if filter.Depo != "" {
		conditions = append(conditions, "gdb.kd_bangsal = ?")
	}
	if filter.Jenis != "" {
		conditions = append(conditions, "dtb.kdjns = ?")
	}
	if filter.Golongan != "" {
		conditions = append(conditions, "dtb.kode_golongan = ?")
	}
	if filter.Kategori != "" {
		conditions = append(conditions, "dtb.kode_kategori = ?")
	}
	if filter.Keyword != "" {
		conditions = append(conditions, "(dtb.nama_brng LIKE ? OR dtb.letak_barang LIKE ?)")
	}

	return " WHERE " + strings.Join(conditions, " AND ")
}

func (r *repository) buildCountArgs(filter FilterDaftarObat) []any {
	var args []any

	if filter.Depo != "" {
		args = append(args, filter.Depo)
	}
	if filter.Jenis != "" {
		args = append(args, filter.Jenis)
	}
	if filter.Golongan != "" {
		args = append(args, filter.Golongan)
	}
	if filter.Kategori != "" {
		args = append(args, filter.Kategori)
	}
	if filter.Keyword != "" {
		keywordPattern := filter.Keyword + "%"
		args = append(args, keywordPattern, keywordPattern)
	}

	return args
}

func (r *repository) hydrateStok(ctx context.Context, listObat []Obat) error {
	kodeList := make([]string, len(listObat))
	for i, o := range listObat {
		kodeList[i] = o.KodeObat
	}

	placeholders := shared.CreateInPlaceholders(len(kodeList))
	query := fmt.Sprintf(`SELECT kode_brng, SUM(stok) AS total_stok 
		FROM gudangbarang 
		WHERE kode_brng IN (%s) AND %s 
		GROUP BY kode_brng`, placeholders, shared.InClauseDepoFarmasi("kd_bangsal"))

	args := make([]any, len(kodeList))
	for i, k := range kodeList {
		args[i] = k
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	stokMap := make(map[string]float64)
	for rows.Next() {
		var kode string
		var totalStok float64
		if err := rows.Scan(&kode, &totalStok); err != nil {
			return err
		}
		stokMap[kode] = totalStok
	}

	if err := rows.Err(); err != nil {
		return err
	}

	for i := range listObat {
		if stok, ok := stokMap[listObat[i].KodeObat]; ok {
			listObat[i].Stok = stok
		}
	}

	return nil
}

func (r *repository) DetailObat(ctx context.Context, kodeObat string) (*Obat, error) {
	query := `SELECT 
		dtb.kode_brng AS kode_obat,
		dtb.nama_brng AS nama_obat,
		dtb.letak_barang AS komposisi,
		dtb.ralan AS harga,
		dtb.kode_sat AS satuan,
		COALESCE((SELECT SUM(gdb.stok) FROM gudangbarang gdb WHERE gdb.kode_brng = dtb.kode_brng AND ` + shared.InClauseDepoFarmasi("gdb.kd_bangsal") + `), 0) AS stok,
		dtb.kapasitas,
		dtb.kdjns AS kode_jenis,
		jn.nama AS nama_jenis,
		dtb.kode_golongan AS kode_golongan,
		gb.nama AS nama_golongan,
		dtb.kode_kategori AS kode_kategori,
		kb.nama AS nama_kategori
	FROM databarang dtb
	INNER JOIN kategori_barang kb ON dtb.kode_kategori = kb.kode 
	INNER JOIN golongan_barang gb ON dtb.kode_golongan = gb.kode 
	INNER JOIN jenis jn ON dtb.kdjns = jn.kdjns
	WHERE dtb.kode_brng = ? AND dtb.status = '1'`

	var o Obat
	err := r.db.QueryRowContext(ctx, query, kodeObat).Scan(
		&o.KodeObat,
		&o.NamaObat,
		&o.Komposisi,
		&o.Harga,
		&o.Satuan,
		&o.Stok,
		&o.Kapasitas,
		&o.KodeJenis,
		&o.NamaJenis,
		&o.KodeGolongan,
		&o.NamaGolongan,
		&o.KodeKategori,
		&o.NamaKategori,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	queryStok := fmt.Sprintf(`SELECT 
		gdb.kd_bangsal AS kode_depo,
		bg.nm_bangsal AS nama_depo,
		gdb.stok
	FROM gudangbarang gdb
	INNER JOIN bangsal bg ON gdb.kd_bangsal = bg.kd_bangsal
	WHERE gdb.kode_brng = ? AND %s
	ORDER BY bg.nm_bangsal DESC`, shared.InClauseDepoFarmasi("gdb.kd_bangsal"))

	rows, err := r.db.QueryContext(ctx, queryStok, kodeObat)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	o.StokDepo = make([]StokDepo, 0)
	for rows.Next() {
		var s StokDepo
		if err := rows.Scan(&s.KodeDepo, &s.NamaDepo, &s.Stok); err != nil {
			return nil, err
		}
		o.StokDepo = append(o.StokDepo, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &o, nil
}

func (r *repository) DaftarJenis(ctx context.Context) ([]JenisObat, error) {
	query := `SELECT kdjns AS kode, nama FROM jenis ORDER BY nama ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]JenisObat, 0)
	for rows.Next() {
		var j JenisObat
		if err := rows.Scan(&j.Kode, &j.Nama); err != nil {
			return nil, err
		}
		list = append(list, j)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *repository) DaftarGolongan(ctx context.Context) ([]GolonganObat, error) {
	query := `SELECT kode, nama FROM golongan_barang ORDER BY nama ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]GolonganObat, 0)
	for rows.Next() {
		var g GolonganObat
		if err := rows.Scan(&g.Kode, &g.Nama); err != nil {
			return nil, err
		}
		list = append(list, g)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *repository) DaftarKategori(ctx context.Context) ([]KategoriObat, error) {
	query := `SELECT kode, nama FROM kategori_barang ORDER BY nama ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]KategoriObat, 0)
	for rows.Next() {
		var k KategoriObat
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

func (r *repository) CekKeberadaanObat(ctx context.Context, listKodeObat []string) (map[string]bool, error) {
	if len(listKodeObat) == 0 {
		return map[string]bool{}, nil
	}

	uniqueCodes := make([]string, 0, len(listKodeObat))
	seen := make(map[string]bool)
	for _, code := range listKodeObat {
		trimmed := strings.TrimSpace(code)
		if trimmed != "" && !seen[trimmed] {
			seen[trimmed] = true
			uniqueCodes = append(uniqueCodes, trimmed)
		}
	}

	if len(uniqueCodes) == 0 {
		return map[string]bool{}, nil
	}

	inPlaceholders := shared.CreateInPlaceholders(len(uniqueCodes))
	args := make([]any, len(uniqueCodes))
	for i, c := range uniqueCodes {
		args[i] = c
	}

	query := fmt.Sprintf("SELECT kode_brng FROM databarang WHERE kode_brng IN (%s)", inPlaceholders)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	foundMap := make(map[string]bool)
	for rows.Next() {
		var kode string
		if err := rows.Scan(&kode); err != nil {
			return nil, err
		}
		foundMap[kode] = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return foundMap, nil
}
