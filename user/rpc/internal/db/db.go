package db

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Init(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		dsn = os.Getenv("POSTGRES_DSN")
	}

	if dsn == "" {
		return nil, fmt.Errorf("postgres dsn is empty")
	}

	gormDB, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{
			TranslateError: true,
			Logger:         logger.Default.LogMode(logLevel()),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("get postgres sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	return gormDB, nil
}

func logLevel() logger.LogLevel {
	switch strings.ToLower(os.Getenv("DB_LOG_LEVEL")) {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "info", "debug":
		return logger.Info
	default:
		return logger.Warn
	}
}
