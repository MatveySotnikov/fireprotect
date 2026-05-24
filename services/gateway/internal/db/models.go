package db

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `gorm:"not null"`
	Email    string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
}

type Material struct {
	gorm.Model
	Title             string  `gorm:"not null"`
	DefaultDensity    float64 `gorm:"not null"` // кг/л
	Group1Consumption float64 `gorm:"not null"` // кг/м² для I группы
	Group2Consumption float64 `gorm:"not null"` // кг/м² для II группы
	BrushLoss         float64 `gorm:"not null;default:1.05"`
	SprayIndoorLoss   float64 `gorm:"not null;default:1.20"`
	SprayOutdoorLoss  float64 `gorm:"not null;default:1.35"`
}

type Calculation struct {
	gorm.Model
	UserID            uint     `gorm:"index;not null"`
	MaterialID        *uint    `gorm:"index"` // nullable, если состав введён вручную
	Area              float64  `gorm:"not null"`
	AreaType          string   `gorm:"not null;default:'slope'"` // "projection" или "slope"
	SlopeAngle        float64  `gorm:"not null"`
	TargetGroup       string   `gorm:"not null;default:'1_group'"` // "1_group" или "2_group"
	ApplicationMethod string   `gorm:"not null;default:'brush'"`   // "brush", "spray_indoor", "spray_outdoor"
	LossFactor        float64  `gorm:"not null"`                   // коэффициент потерь (1.05, 1.20 и т.д.)
	Layers            int32    `gorm:"not null;default:1"`         // всегда 1 в новой логике
	UsedNormativeRate float64  `gorm:"not null"`                   // итоговый расход кг/м² (base * loss)
	UsedDensity       float64  `gorm:"not null"`                   // плотность из справочника
	User              User     `gorm:"foreignKey:UserID"`
	Material          Material `gorm:"foreignKey:MaterialID"`
}
