package database

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/qustavo/sqlhooks/v2"

	"erm-dokter/internal/pkg/logger"
)

// DriverNameWithTracker adalah nama driver yang dibungkus oleh hook trackersql
const DriverNameWithTracker = "mysql-tracker"

var (
	trackerHook  *TrackerHook
	registerOnce sync.Once
)

func initDriver(log *logger.Logger) {
	registerOnce.Do(func() {
		trackerHook = NewTrackerHook(log)
		sql.Register(DriverNameWithTracker, sqlhooks.Wrap(&mysqlDriver.MySQLDriver{}, trackerHook))
	})
	if log != nil && trackerHook != nil {
		trackerHook.SetLogger(log)
	}
}

// GetTrackerHook mengembalikan instance TrackerHook global
func GetTrackerHook() *TrackerHook {
	return trackerHook
}

func InitMySQL(host, port, user, password, dbname string, loggers ...*logger.Logger) (*sql.DB, error) {
	var log *logger.Logger
	if len(loggers) > 0 {
		log = loggers[0]
	}

	initDriver(log)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=5s",
		user, password, host, port, dbname)

	db, err := sql.Open(DriverNameWithTracker, dsn)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi database: %w", err)
	}

	if trackerHook != nil {
		trackerHook.SetDB(db)
	}

	db.SetMaxOpenConns(40)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("gagal ping ke database MySQL: %w", err)
	}

	return db, nil
}
