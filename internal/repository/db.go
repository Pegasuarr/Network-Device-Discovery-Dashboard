package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/user/network-monitoring/configs"
	"github.com/user/network-monitoring/internal/model"
)

var (
	DB    *gorm.DB
	RDB   *redis.Client
	CtxDb = context.Background()
)

func InitDB(cfg *configs.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)

	var db *gorm.DB
	var err error

	// Retry database connection a few times for docker startup
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Auto-Migrate Tables
	err = db.AutoMigrate(
		&model.Organization{},
		&model.Role{},
		&model.User{},
		&model.RefreshToken{},
		&model.DeviceGroup{},
		&model.Device{},
		&model.AlertRule{},
		&model.MonitoringResult{},
		&model.Alert{},
		&model.NotificationLog{},
		&model.AuditLog{},
		&model.LoginLog{},
		&model.Setting{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate: %v", err)
	}

	DB = db
	return db, nil
}

func InitRedis(cfg *configs.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %v", err)
	}

	RDB = rdb
	return rdb, nil
}
