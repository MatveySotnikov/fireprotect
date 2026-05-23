package db

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `gorm:"not null"`
	Email    string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
}

type Calculation struct {
	gorm.Model
	UserID        uint    `gorm:"index;not null"`
	Area          float64 `gorm:"not null"`
	NormativeRate float64 `gorm:"not null"`
	Layers        int32   `gorm:"not null"`
	SlopeAngle    float64 `gorm:"not null"`
	LossFactor    float64 `gorm:"not null"`
	Density       float64 `gorm:"not null"`
	TotalMass     float64 `gorm:"not null"`
	TotalVolume   float64 `gorm:"not null"`
	User          User    `gorm:"foreignKey:UserID"`
}
