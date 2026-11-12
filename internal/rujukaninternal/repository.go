package rujukaninternal

import (
	"context"
	"database/sql"

	"erm-dokter/internal/pkg/database"
)

type Repository interface {
	DaftarOpsiPoliDokter(ctx context.Context, kodeDokterLogin, keyword string) ([]OpsiPoliDokter, error)
	DaftarRujukanInternalByNoRawat(ctx context.Context, noRawat string) ([]RujukanInternal, error)
	CekOpsiPoliDokterAktif(ctx context.Context, kdPoli, kdDokter string) (bool, error)
	CekRujukanInternalAda(ctx context.Context, noRawat, kdDokter string) (bool, error)
	CekRujukanInternalPoliAda(ctx context.Context, noRawat, kdPoli, kdDokter string) (bool, error)
	CekAktivitasDokterKonsulen(ctx context.Context, noRawat, kdDokter string) (bool, error)
	SimpanRujukanInternal(ctx context.Context, noRawat, kdPoli, kdDokter string) error
	HapusRujukanInternal(ctx context.Context, noRawat, kdPoli, kdDokter string) error
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
		COALESCE(rip.kd_poli, '') AS kd_poli,
		COALESCE(p.nm_poli, '') AS nama_poli,
		rip.kd_dokter,
		COALESCE(d.nm_dokter, '') AS nama_dokter
	FROM rujukan_internal_poli rip
	LEFT JOIN poliklinik p ON rip.kd_poli = p.kd_poli
	LEFT JOIN dokter d ON rip.kd_dokter = d.kd_dokter
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

const existsOpsiPoliDokterAktif = `
	SELECT EXISTS(
		SELECT 1
		FROM jadwal j
		INNER JOIN poliklinik p ON j.kd_poli = p.kd_poli
		INNER JOIN dokter d ON j.kd_dokter = d.kd_dokter
		WHERE j.kd_poli = ? AND j.kd_dokter = ? AND p.status = '1' AND d.status = '1'
		LIMIT 1
	)
`

func (r *repository) CekOpsiPoliDokterAktif(ctx context.Context, kdPoli, kdDokter string) (bool, error) {
	var exists int
	if err := r.db.QueryRowContext(ctx, existsOpsiPoliDokterAktif, kdPoli, kdDokter).Scan(&exists); err != nil {
		return false, err
	}
	return exists > 0, nil
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

const countRujukanInternalByPoliDokter = `
	SELECT COUNT(*)
	FROM rujukan_internal_poli
	WHERE no_rawat = ? AND kd_poli = ? AND kd_dokter = ?
`

func (r *repository) CekRujukanInternalPoliAda(ctx context.Context, noRawat, kdPoli, kdDokter string) (bool, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, countRujukanInternalByPoliDokter, noRawat, kdPoli, kdDokter).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

const existsAktivitasDokterKonsulen = `
	SELECT (
		EXISTS(SELECT 1 FROM pemeriksaan_ralan WHERE no_rawat = ? AND nip = ? LIMIT 1)
		OR EXISTS(SELECT 1 FROM penilaian_medis_ralan WHERE no_rawat = ? AND kd_dokter = ? LIMIT 1)
		OR EXISTS(SELECT 1 FROM penilaian_medis_ralan_kandungan WHERE no_rawat = ? AND kd_dokter = ? LIMIT 1)
		OR EXISTS(SELECT 1 FROM penilaian_medis_igd WHERE no_rawat = ? AND kd_dokter = ? LIMIT 1)
		OR EXISTS(SELECT 1 FROM resep_obat WHERE no_rawat = ? AND kd_dokter = ? LIMIT 1)
	)
`

func (r *repository) CekAktivitasDokterKonsulen(ctx context.Context, noRawat, kdDokter string) (bool, error) {
	var active int
	err := r.db.QueryRowContext(
		ctx,
		existsAktivitasDokterKonsulen,
		noRawat, kdDokter,
		noRawat, kdDokter,
		noRawat, kdDokter,
		noRawat, kdDokter,
		noRawat, kdDokter,
	).Scan(&active)
	if err != nil {
		return false, err
	}
	return active > 0, nil
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
	WHERE no_rawat = ? AND kd_poli = ? AND kd_dokter = ?
`

func (r *repository) HapusRujukanInternal(ctx context.Context, noRawat, kdPoli, kdDokter string) error {
	res, err := r.db.ExecContext(ctx, deleteRujukanInternal, noRawat, kdPoli, kdDokter)
	if err != nil {
		return err
	}
	return database.CheckRowsAffected(res)
}
