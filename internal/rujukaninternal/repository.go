package rujukaninternal

import (
	"context"
	"database/sql"

	"erm-dokter/internal/pkg/database"
)

type Repository interface {
	DaftarOpsiPoliDokter(ctx context.Context, kodeDokterLogin, keyword string) ([]OpsiPoliDokter, error)
	DaftarRujukanInternalByNoRawat(ctx context.Context, noRawat string) ([]RujukanInternal, error)
	CekRujukanInternalAda(ctx context.Context, noRawat, kdDokter string) (bool, error)
	SimpanRujukanInternal(ctx context.Context, noRawat, kdPoli, kdDokter string) error
	HapusRujukanInternal(ctx context.Context, noRawat, kdDokter string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) DaftarOpsiPoliDokter(ctx context.Context, kodeDokterLogin, keyword string) ([]OpsiPoliDokter, error) {
	query := `
		SELECT 
			j.kd_poli,
			COALESCE(p.nm_poli, '') AS nama_poli,
			j.kd_dokter,
			COALESCE(d.nm_dokter, '') AS nama_dokter
		FROM jadwal j
		INNER JOIN poliklinik p ON j.kd_poli = p.kd_poli
		INNER JOIN dokter d ON j.kd_dokter = d.kd_dokter
		WHERE j.kd_dokter != ? AND p.status = '1' AND d.status = '1'
	`
	args := []any{kodeDokterLogin}

	if keyword != "" {
		query += ` AND (p.nm_poli LIKE ? OR d.nm_dokter LIKE ?)`
		pattern := "%" + keyword + "%"
		args = append(args, pattern, pattern)
	}

	query += `
		GROUP BY j.kd_poli, j.kd_dokter
		ORDER BY p.nm_poli ASC, d.nm_dokter ASC
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]OpsiPoliDokter, 0)
	for rows.Next() {
		var item OpsiPoliDokter
		if err := rows.Scan(
			&item.KodePoli,
			&item.NamaPoli,
			&item.KodeDokter,
			&item.NamaDokter,
		); err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

const selectRujukanInternal = `
	SELECT 
		rip.no_rawat,
		rip.kd_poli,
		COALESCE(p.nm_poli, '') AS nama_poli,
		rip.kd_dokter,
		COALESCE(d.nm_dokter, '') AS nama_dokter
	FROM rujukan_internal_poli rip
	INNER JOIN poliklinik p ON rip.kd_poli = p.kd_poli
	INNER JOIN dokter d ON rip.kd_dokter = d.kd_dokter
	WHERE rip.no_rawat = ?
	ORDER BY p.nm_poli ASC, d.nm_dokter ASC
`

func (r *repository) DaftarRujukanInternalByNoRawat(ctx context.Context, noRawat string) ([]RujukanInternal, error) {
	rows, err := r.db.QueryContext(ctx, selectRujukanInternal, noRawat)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]RujukanInternal, 0)
	for rows.Next() {
		var item RujukanInternal
		if err := rows.Scan(
			&item.NoRawat,
			&item.KodePoli,
			&item.NamaPoli,
			&item.KodeDokter,
			&item.NamaDokter,
		); err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

const countRujukanInternalByDokter = `
	SELECT COUNT(*) 
	FROM rujukan_internal_poli 
	WHERE no_rawat = ? AND kd_dokter = ?
`

func (r *repository) CekRujukanInternalAda(ctx context.Context, noRawat, kdDokter string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, countRujukanInternalByDokter, noRawat, kdDokter).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

const insertRujukanInternal = `
	INSERT INTO rujukan_internal_poli (no_rawat, kd_dokter, kd_poli) 
	VALUES (?, ?, ?)
`

func (r *repository) SimpanRujukanInternal(ctx context.Context, noRawat, kdPoli, kdDokter string) error {
	_, err := r.db.ExecContext(ctx, insertRujukanInternal, noRawat, kdDokter, kdPoli)
	if err != nil {
		return err
	}
	return nil
}

const deleteRujukanInternal = `
	DELETE FROM rujukan_internal_poli 
	WHERE no_rawat = ? AND kd_dokter = ?
`

func (r *repository) HapusRujukanInternal(ctx context.Context, noRawat, kdDokter string) error {
	res, err := r.db.ExecContext(ctx, deleteRujukanInternal, noRawat, kdDokter)
	if err != nil {
		return err
	}
	return database.CheckRowsAffected(res)
}
