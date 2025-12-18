package team

import "time"

// Team 战队实体
type Team struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	ShortName string    `json:"shortName"`
	LogoURL   string    `json:"logoUrl"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
