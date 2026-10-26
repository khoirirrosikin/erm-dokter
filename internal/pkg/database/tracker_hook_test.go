package database

import (
	"context"
	"strings"
	"testing"

	"erm-dokter/internal/middleware"
	"erm-dokter/internal/pkg/token"
)

func TestFormatTrackerSQL_Insert(t *testing.T) {
	query := `
		INSERT INTO pemeriksaan_ranap (
			no_rawat, tgl_perawatan, jam_rawat, suhu_tubuh, tensi, nadi, respirasi,
			tinggi, berat, spo2, gcs, kesadaran, keluhan, pemeriksaan, alergi,
			rtl, penilaian, instruksi, evaluasi, nip
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	args := []any{
		"2026/10/12/099101", "2026-10-26", "12:00:53", "", "", "", "",
		"159", "60", "", "425", "Delirium", "Pasien belum dapat di ajak komunikasi",
		"K/U LEMAH, TERPASANG NC 4 LPM\nterpasang O2 NC 3 lpm", "-",
		"BERSIHAN JALAN NAFAS TIDAK EFEKTIF", "1. Monitor pola napas",
		"implementasi\n1. Memonitor pola napas", "DIHARAPKAN SETELAH DILAKUKAN ASKEP", "VARIHATUL",
	}

	result, isCUD := FormatTrackerSQL(query, args)
	if !isCUD {
		t.Fatalf("expected isCUD=true, got false")
	}

	expectedPrefix := "insert into pemeriksaan_ranap values(|2026/10/12/099101|2026-10-26|12:00:53|||||159|60||425|Delirium|Pasien belum dapat di ajak komunikasi|"
	expectedSuffix := "|VARIHATUL)"

	if !strings.HasPrefix(result, expectedPrefix) {
		t.Errorf("expected prefix:\n%s\ngot:\n%s", expectedPrefix, result)
	}
	if !strings.HasSuffix(result, expectedSuffix) {
		t.Errorf("expected suffix:\n%s\ngot:\n%s", expectedSuffix, result)
	}
}

func TestFormatTrackerSQL_Update(t *testing.T) {
	query := `
		UPDATE pemeriksaan_ranap SET
			tgl_perawatan = ?, jam_rawat = ?,
			suhu_tubuh = ?, tensi = ?
		WHERE no_rawat = ? AND tgl_perawatan = ?
	`
	args := []any{
		"2026-10-26", "12:00:53", "36.5", "120/80", "2026/10/12/099101", "2026-10-26",
	}

	result, isCUD := FormatTrackerSQL(query, args)
	if !isCUD {
		t.Fatalf("expected isCUD=true, got false")
	}

	expected := "update pemeriksaan_ranap set tgl_perawatan = |2026-10-26|, jam_rawat = |12:00:53|, suhu_tubuh = |36.5|, tensi = |120/80| where no_rawat = |2026/10/12/099101| and tgl_perawatan = |2026-10-26|"
	if result != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestFormatTrackerSQL_Delete(t *testing.T) {
	query := "DELETE FROM pemeriksaan_ranap WHERE no_rawat = ? AND tgl_perawatan = ? AND jam_rawat = ?"
	args := []any{
		"2026/10/12/099101", "2026-10-26", "12:00:53",
	}

	result, isCUD := FormatTrackerSQL(query, args)
	if !isCUD {
		t.Fatalf("expected isCUD=true, got false")
	}

	expected := "delete from pemeriksaan_ranap where no_rawat = |2026/10/12/099101| and tgl_perawatan = |2026-10-26| and jam_rawat = |12:00:53|"
	if result != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestFormatTrackerSQL_SelectIgnored(t *testing.T) {
	query := "SELECT * FROM pemeriksaan_ranap WHERE no_rawat = ?"
	args := []any{"2026/10/12/099101"}

	_, isCUD := FormatTrackerSQL(query, args)
	if isCUD {
		t.Errorf("expected isCUD=false for SELECT query, got true")
	}
}

func TestFormatTrackerSQL_TrackerSQLIgnored(t *testing.T) {
	query := "INSERT INTO trackersql (tanggal, sqle, usere) VALUES (NOW(), ?, ?)"
	args := []any{"insert into test values(1)", "D001"}

	_, isCUD := FormatTrackerSQL(query, args)
	if isCUD {
		t.Errorf("expected isCUD=false for query touching trackersql, got true")
	}
}

func TestTrackerHookIntegration(t *testing.T) {
	hook := NewTrackerHook(nil)
	ctx := context.WithValue(context.Background(), middleware.UserClaimKey, &token.Claims{
		KodeDokter: "D0001",
		NamaUser:   "dr. Budi",
	})

	var capturedQuery, capturedUser string
	hook.SetOnTrack(func(sqle, usere string) {
		capturedQuery = sqle
		capturedUser = usere
	})

	query := "UPDATE pemeriksaan_ranap SET keluhan = ? WHERE no_rawat = ?"
	args := []any{"Pusing berputar", "2026/10/26/000001"}

	_, err := hook.After(ctx, query, args...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedUser != "D0001" {
		t.Errorf("expected user D0001, got %s", capturedUser)
	}
	expectedSQL := "update pemeriksaan_ranap set keluhan = |Pusing berputar| where no_rawat = |2026/10/26/000001|"
	if capturedQuery != expectedSQL {
		t.Errorf("expected SQL:\n%s\ngot:\n%s", expectedSQL, capturedQuery)
	}
}
