package db

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() error {
	host := getEnv("DB_HOST", "127.0.0.1")
	port := getEnv("DB_PORT", "5433")
	user := getEnv("POSTGRES_USER", "fireprotect")
	password := getEnv("POSTGRES_PASSWORD", "secret")
	dbname := getEnv("POSTGRES_DB", "fireprotect")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Автоматическое создание/обновление таблиц
	if err := DB.AutoMigrate(&User{}, &Material{}, &Calculation{}); err != nil {
		return fmt.Errorf("auto migration failed: %w", err)
	}

	// Заполнение справочника материалов (если пусто)
	seedMaterials()

	return nil
}

func seedMaterials() {
	var count int64
	DB.Model(&Material{}).Count(&count)
	if count > 0 {
		return
	}

	materials := []Material{
		{
			Title:             "Огнебиощит Эксперт (дерево)",
			DefaultDensity:    1.10,
			Group1Consumption: 0.500,
			Group2Consumption: 0.300,
			BrushLoss:         1.05,
			SprayIndoorLoss:   1.20,
			SprayOutdoorLoss:  1.35,
		},
		{
			Title:             "Терма-Металл (металл)",
			DefaultDensity:    1.35,
			Group1Consumption: 1.250, // для R45
			Group2Consumption: 0.0,   // не используется
			BrushLoss:         1.10,
			SprayIndoorLoss:   1.25,
			SprayOutdoorLoss:  1.40,
		},
		{
			Title:             "Универсальный состав (пример)",
			DefaultDensity:    1.20,
			Group1Consumption: 0.600,
			Group2Consumption: 0.400,
			BrushLoss:         1.05,
			SprayIndoorLoss:   1.20,
			SprayOutdoorLoss:  1.30,
		},
	}

	for _, m := range materials {
		DB.Create(&m)
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
