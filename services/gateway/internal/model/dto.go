package model

// CalcRequest – новый формат запроса на расчёт
type CalcRequest struct {
	Area              float64  `json:"area"`
	AreaType          string   `json:"area_type"` // "projection" или "slope"
	SlopeAngle        float64  `json:"slope_angle"`
	TargetGroup       string   `json:"target_group"`       // "1_group" или "2_group"
	ApplicationMethod string   `json:"application_method"` // "brush", "spray_indoor", "spray_outdoor"
	MaterialID        *uint    `json:"material_id,omitempty"`
	NormativeRate     *float64 `json:"normative_rate,omitempty"`
	Density           *float64 `json:"density,omitempty"`
}

// CalcResponse – результат расчёта
type CalcResponse struct {
	TotalMass   float64 `json:"total_mass"`
	TotalVolume float64 `json:"total_volume"`
}

// RegisterRequest – регистрация
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest – аутентификация
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenResponse – JWT
type TokenResponse struct {
	Token string `json:"token"`
}

// CalcEnriched – расширенная информация о расчёте с вычисленными полями
type CalcEnriched struct {
	ID                uint    `json:"id"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	DeletedAt         *string `json:"deleted_at"`
	UserID            uint    `json:"user_id"`
	MaterialID        *uint   `json:"material_id"`
	Area              float64 `json:"area"`
	AreaType          string  `json:"area_type"`
	SlopeAngle        float64 `json:"slope_angle"`
	TargetGroup       string  `json:"target_group"`
	ApplicationMethod string  `json:"application_method"`
	LossFactor        float64 `json:"loss_factor"`
	Layers            int32   `json:"layers"`
	UsedNormativeRate float64 `json:"used_normative_rate"`
	UsedDensity       float64 `json:"used_density"`
	User              any     `json:"user,omitempty"`
	Material          any     `json:"material,omitempty"`
	TotalMass         float64 `json:"total_mass"`
	TotalVolume       float64 `json:"total_volume"`
}
