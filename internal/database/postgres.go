package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"os"
	"test-em/internal/config"
	"test-em/internal/logger"
)

func LoadDB(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		cfg.DbHost, cfg.DbUser, cfg.DbPassword, cfg.DbName, cfg.DbPort)

	logger.Info("Attempting to connect to database", "dsn", fmt.Sprintf("host=%s user=%s dbname=%s port=%s", cfg.DbHost, cfg.DbUser, cfg.DbName, cfg.DbPort))

	var gormLogLevel gormlogger.LogLevel
	logLevel := getEnv("LOG_LEVEL", "info")
	switch logLevel {
	case "debug":
		gormLogLevel = gormlogger.Info
	case "info":
		gormLogLevel = gormlogger.Warn
	case "warn", "error":
		gormLogLevel = gormlogger.Error
	default:
		gormLogLevel = gormlogger.Warn
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		logger.WithError(err).Fatal("Failed to get database instance")
	}

	if err := sqlDB.Ping(); err != nil {
		logger.WithError(err).Fatal("Failed to ping database")
	}

	logger.Info("Database connection successful", "host", cfg.DbHost, "port", cfg.DbPort, "database", cfg.DbName)
	return db
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
