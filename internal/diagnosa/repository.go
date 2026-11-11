package diagnosa

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"erm-dokter/internal/pkg/database"
	"erm-dokter/internal/shared"
)

type Repository interface {
	GetDiagnosaByNoRawat(ctx context.Context, noRawat string, status shared.StatusLanjut) ([]DiagnosaPasien, error)
	GetProsedurByNoRawat(ctx context.Context, noRawat string, status shared.StatusLanjut) ([]ProsedurPasien, error)
	GetRiwayatDiagnosaByListNoRawat(ctx context.Context, listNoRawat []string, status shared.StatusLanjut) ([]DiagnosaPasien, error)
	GetRiwayatProsedurByListNoRawat(ctx context.Context, listNoRawat []string, status shared.StatusLanjut) ([]ProsedurPasien, error)
	CekRiwayatPenyakit(ctx context.Context, listPastNoRawat []string, kodePenyakit string, status shared.StatusLanjut) (bool, error)
	GetMaxPrioritasDiagnosa(ctx context.Context, noRawat string, status shared.StatusLanjut) (int, error)
	GetMaxPrioritasProsedur(ctx context.Context, noRawat string, status shared.StatusLanjut) (int, error)
	CekDuplikasiDiagnosa(ctx context.Context, noRawat string, kodePenyakit string, status shared.StatusLanjut) (bool, error)
	CekDuplikasiProsedur(ctx context.Context, noRawat string, kodeProsedur string, status shared.StatusLanjut) (bool, error)
	TambahDiagnosa(ctx context.Context, diagnosa DiagnosaPasien) error
	TambahProsedur(ctx context.Context, prosedur ProsedurPasien) error
	UpdateDiagnosa(ctx context.Context, noRawat, kode string, status shared.StatusLanjut, prioritas *int, statusPenyakit string) error
	UpdateProsedur(ctx context.Context, noRawat, kode string, status shared.StatusLanjut, prioritas int) error
	HapusDiagnosa(ctx context.Context, noRawat, kode string, status shared.StatusLanjut) error
	HapusProsedur(ctx context.Context, noRawat, kode string, status shared.StatusLanjut) error
	ReorderDiagnosa(ctx context.Context, noRawat string, status shared.StatusLanjut, items []ReorderItemRequest) error
	ReorderProsedur(ctx context.Context, noRawat string, status shared.StatusLanjut, items []ReorderItemRequest) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetDiagnosaByNoRawat(ctx context.Context, noRawat string, status shared.StatusLanjut) ([]DiagnosaPasien, error) {
	query := `
		SELECT dp.no_rawat, dp.kd_penyakit, COALESCE(p.nm_penyakit, '') AS nama,
		       dp.status, dp.prioritas, dp.status_penyakit
		FROM diagnosa_pasien dp
		INNER JOIN penyakit p ON dp.kd_penyakit = p.kd_penyakit
		WHERE dp.no_rawat = ?
	`
	args := []any{noRawat}
	if status != "" && !strings.EqualFold(string(status), "semua") {
		query += " AND dp.status = ?"
		args = append(args, string(status))
	}
	query += " ORDER BY dp.prioritas ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DiagnosaPasien
	for rows.Next() {
		var item DiagnosaPasien
		var rawStatus string
		if err := rows.Scan(
			&item.NoRawat,
			&item.Kode,
			&item.Nama,
			&rawStatus,
			&item.Prioritas,
			&item.StatusPenyakit,
		); err != nil {
			return nil, err
		}
		item.Status = shared.StatusLanjut(rawStatus)
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]DiagnosaPasien, 0)
	}
	return result, nil
}

func (r *repository) GetProsedurByNoRawat(ctx context.Context, noRawat string, status shared.StatusLanjut) ([]ProsedurPasien, error) {
	query := `
		SELECT pp.no_rawat, pp.kode, COALESCE(i.deskripsi_panjang, '') AS nama,
		       pp.status, pp.prioritas
		FROM prosedur_pasien pp
		INNER JOIN icd9 i ON pp.kode = i.kode
		WHERE pp.no_rawat = ?
	`
	args := []any{noRawat}
	if status != "" && !strings.EqualFold(string(status), "semua") {
		query += " AND pp.status = ?"
		args = append(args, string(status))
	}
	query += " ORDER BY pp.prioritas ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ProsedurPasien
	for rows.Next() {
		var item ProsedurPasien
		var rawStatus string
		if err := rows.Scan(
			&item.NoRawat,
			&item.Kode,
			&item.Nama,
			&rawStatus,
			&item.Prioritas,
		); err != nil {
			return nil, err
		}
		item.Status = shared.StatusLanjut(rawStatus)
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]ProsedurPasien, 0)
	}
	return result, nil
}

func (r *repository) GetRiwayatDiagnosaByListNoRawat(ctx context.Context, listNoRawat []string, status shared.StatusLanjut) ([]DiagnosaPasien, error) {
	if len(listNoRawat) == 0 {
		return make([]DiagnosaPasien, 0), nil
	}

	placeholders := make([]string, len(listNoRawat))
	args := make([]any, len(listNoRawat))
	for i, nr := range listNoRawat {
		placeholders[i] = "?"
		args[i] = nr
	}

	query := fmt.Sprintf(`
		SELECT dp.no_rawat, dp.kd_penyakit, COALESCE(p.nm_penyakit, '') AS nama,
		       dp.status, dp.prioritas, dp.status_penyakit
		FROM diagnosa_pasien dp
		INNER JOIN penyakit p ON dp.kd_penyakit = p.kd_penyakit
		WHERE dp.no_rawat IN (%s)
	`, strings.Join(placeholders, ","))

	if status != "" && !strings.EqualFold(string(status), "semua") {
		query += " AND dp.status = ?"
		args = append(args, string(status))
	}

	query += " ORDER BY dp.no_rawat DESC, dp.prioritas ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DiagnosaPasien
	for rows.Next() {
		var item DiagnosaPasien
		var rawStatus string
		if err := rows.Scan(
			&item.NoRawat,
			&item.Kode,
			&item.Nama,
			&rawStatus,
			&item.Prioritas,
			&item.StatusPenyakit,
		); err != nil {
			return nil, err
		}
		item.Status = shared.StatusLanjut(rawStatus)
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]DiagnosaPasien, 0)
	}
	return result, nil
}

func (r *repository) GetRiwayatProsedurByListNoRawat(ctx context.Context, listNoRawat []string, status shared.StatusLanjut) ([]ProsedurPasien, error) {
	if len(listNoRawat) == 0 {
		return make([]ProsedurPasien, 0), nil
	}

	placeholders := make([]string, len(listNoRawat))
	args := make([]any, len(listNoRawat))
	for i, nr := range listNoRawat {
		placeholders[i] = "?"
		args[i] = nr
	}

	query := fmt.Sprintf(`
		SELECT pp.no_rawat, pp.kode, COALESCE(i.deskripsi_panjang, '') AS nama,
		       pp.status, pp.prioritas
		FROM prosedur_pasien pp
		INNER JOIN icd9 i ON pp.kode = i.kode
		WHERE pp.no_rawat IN (%s)
	`, strings.Join(placeholders, ","))

	if status != "" && !strings.EqualFold(string(status), "semua") {
		query += " AND pp.status = ?"
		args = append(args, string(status))
	}

	query += " ORDER BY pp.no_rawat DESC, pp.prioritas ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ProsedurPasien
	for rows.Next() {
		var item ProsedurPasien
		var rawStatus string
		if err := rows.Scan(
			&item.NoRawat,
			&item.Kode,
			&item.Nama,
			&rawStatus,
			&item.Prioritas,
		); err != nil {
			return nil, err
		}
		item.Status = shared.StatusLanjut(rawStatus)
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]ProsedurPasien, 0)
	}
	return result, nil
}

func (r *repository) CekRiwayatPenyakit(ctx context.Context, listPastNoRawat []string, kodePenyakit string, status shared.StatusLanjut) (bool, error) {
	if len(listPastNoRawat) == 0 {
		return false, nil
	}

	placeholders := make([]string, len(listPastNoRawat))
	args := make([]any, len(listPastNoRawat)+2)
	for i, nr := range listPastNoRawat {
		placeholders[i] = "?"
		args[i] = nr
	}
	args[len(listPastNoRawat)] = kodePenyakit
	args[len(listPastNoRawat)+1] = string(status)

	query := fmt.Sprintf(`
		SELECT 1 FROM diagnosa_pasien
		WHERE no_rawat IN (%s) AND kd_penyakit = ? AND status = ?
		LIMIT 1
	`, strings.Join(placeholders, ","))

	var exists int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *repository) GetMaxPrioritasDiagnosa(ctx context.Context, noRawat string, status shared.StatusLanjut) (int, error) {
	query := "SELECT COALESCE(MAX(prioritas), 0) FROM diagnosa_pasien WHERE no_rawat = ? AND status = ?"
	var maxPrio int
	err := r.db.QueryRowContext(ctx, query, noRawat, string(status)).Scan(&maxPrio)
	if err != nil {
		return 0, err
	}
	return maxPrio, nil
}

func (r *repository) GetMaxPrioritasProsedur(ctx context.Context, noRawat string, status shared.StatusLanjut) (int, error) {
	query := "SELECT COALESCE(MAX(prioritas), 0) FROM prosedur_pasien WHERE no_rawat = ? AND status = ?"
	var maxPrio int
	err := r.db.QueryRowContext(ctx, query, noRawat, string(status)).Scan(&maxPrio)
	if err != nil {
		return 0, err
	}
	return maxPrio, nil
}

func (r *repository) CekDuplikasiDiagnosa(ctx context.Context, noRawat string, kodePenyakit string, status shared.StatusLanjut) (bool, error) {
	query := "SELECT 1 FROM diagnosa_pasien WHERE no_rawat = ? AND kd_penyakit = ? AND status = ? LIMIT 1"
	var exists int
	err := r.db.QueryRowContext(ctx, query, noRawat, kodePenyakit, string(status)).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *repository) CekDuplikasiProsedur(ctx context.Context, noRawat string, kodeProsedur string, status shared.StatusLanjut) (bool, error) {
	query := "SELECT 1 FROM prosedur_pasien WHERE no_rawat = ? AND kode = ? AND status = ? LIMIT 1"
	var exists int
	err := r.db.QueryRowContext(ctx, query, noRawat, kodeProsedur, string(status)).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *repository) TambahDiagnosa(ctx context.Context, diagnosa DiagnosaPasien) error {
	query := `
		INSERT INTO diagnosa_pasien (no_rawat, kd_penyakit, status, prioritas, status_penyakit)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, diagnosa.NoRawat, diagnosa.Kode, string(diagnosa.Status), diagnosa.Prioritas, diagnosa.StatusPenyakit)
	return err
}

func (r *repository) TambahProsedur(ctx context.Context, prosedur ProsedurPasien) error {
	query := `
		INSERT INTO prosedur_pasien (no_rawat, kode, status, prioritas)
		VALUES (?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, prosedur.NoRawat, prosedur.Kode, string(prosedur.Status), prosedur.Prioritas)
	return err
}

func (r *repository) UpdateDiagnosa(ctx context.Context, noRawat, kode string, status shared.StatusLanjut, prioritas *int, statusPenyakit string) error {
	var sets []string
	var args []any

	if prioritas != nil {
		sets = append(sets, "prioritas = ?")
		args = append(args, *prioritas)
	}
	if statusPenyakit != "" {
		sets = append(sets, "status_penyakit = ?")
		args = append(args, statusPenyakit)
	}

	if len(sets) == 0 {
		return nil
	}

	query := fmt.Sprintf("UPDATE diagnosa_pasien SET %s WHERE no_rawat = ? AND kd_penyakit = ? AND status = ?", strings.Join(sets, ", "))
	args = append(args, noRawat, kode, string(status))

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	return database.CheckRowsAffected(result)
}

func (r *repository) UpdateProsedur(ctx context.Context, noRawat, kode string, status shared.StatusLanjut, prioritas int) error {
	query := "UPDATE prosedur_pasien SET prioritas = ? WHERE no_rawat = ? AND kode = ? AND status = ?"
	result, err := r.db.ExecContext(ctx, query, prioritas, noRawat, kode, string(status))
	if err != nil {
		return err
	}
	return database.CheckRowsAffected(result)
}

func (r *repository) HapusDiagnosa(ctx context.Context, noRawat, kode string, status shared.StatusLanjut) error {
	query := "DELETE FROM diagnosa_pasien WHERE no_rawat = ? AND kd_penyakit = ? AND status = ?"
	result, err := r.db.ExecContext(ctx, query, noRawat, kode, string(status))
	if err != nil {
		return err
	}
	return database.CheckRowsAffected(result)
}

func (r *repository) HapusProsedur(ctx context.Context, noRawat, kode string, status shared.StatusLanjut) error {
	query := "DELETE FROM prosedur_pasien WHERE no_rawat = ? AND kode = ? AND status = ?"
	result, err := r.db.ExecContext(ctx, query, noRawat, kode, string(status))
	if err != nil {
		return err
	}
	return database.CheckRowsAffected(result)
}

func (r *repository) ReorderDiagnosa(ctx context.Context, noRawat string, status shared.StatusLanjut, items []ReorderItemRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "UPDATE diagnosa_pasien SET prioritas = ? WHERE no_rawat = ? AND kd_penyakit = ? AND status = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		res, err := stmt.ExecContext(ctx, item.Prioritas, noRawat, item.Kode, string(status))
		if err != nil {
			return err
		}
		if err := database.CheckRowsAffected(res); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *repository) ReorderProsedur(ctx context.Context, noRawat string, status shared.StatusLanjut, items []ReorderItemRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "UPDATE prosedur_pasien SET prioritas = ? WHERE no_rawat = ? AND kode = ? AND status = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		res, err := stmt.ExecContext(ctx, item.Prioritas, noRawat, item.Kode, string(status))
		if err != nil {
			return err
		}
		if err := database.CheckRowsAffected(res); err != nil {
			return err
		}
	}

	return tx.Commit()
}
