package database

import (
	"database/sql"
	"fmt"
	"net"
	"sync"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/qustavo/sqlhooks/v2"

	"erm-dokter/internal/pkg/logger"
)

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

func GetTrackerHook() *TrackerHook {
	return trackerHook
}

func buildDSN(host, port, user, password, dbname string) string {
	cfg := mysqlDriver.NewConfig()
	cfg.User = user
	cfg.Passwd = password
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(host, port)
	cfg.DBName = dbname
	cfg.Timeout = 5 * time.Second
	cfg.ClientFoundRows = true
	cfg.ParseTime = false
	cfg.Loc = time.Local
	return cfg.FormatDSN()
}

func InitMySQL(host, port, user, password, dbname string, loggers ...*logger.Logger) (*sql.DB, error) {
	var log *logger.Logger
	if len(loggers) > 0 {
		log = loggers[0]
	}

	initDriver(log)

	dsn := buildDSN(host, port, user, password, dbname)

	db, err := sql.Open(DriverNameWithTracker, dsn)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi database: %w", err)
	}

	db.SetMaxOpenConns(40)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("gagal ping ke database MySQL: %w", err)
	}

	if trackerHook != nil {
		trackerHook.SetDB(db)
	}

	return db, nil
}

func CheckRowsAffected(res sql.Result) error {
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
