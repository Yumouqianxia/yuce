package user

import (
	"time"
)

// User 用户实体
type User struct {
	ID                 uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	Username           string     `json:"username" gorm:"uniqueIndex:idx_username;size:50;not null"`
	Email              string     `json:"email" gorm:"uniqueIndex:idx_email;size:100;not null"`
	Nickname           string     `json:"nickname" gorm:"size:50"`
	Password           string     `json:"-" gorm:"size:255;not null"`
	Avatar             string     `json:"avatar" gorm:"size:255"`
	Points             int        `json:"points" gorm:"default:0;index:idx_points"`
	Role               UserRole   `json:"role" gorm:"default:user;size:20;not null"`
	CreatedAt          time.Time  `json:"createdAt" gorm:"column:createdAt;autoCreateTime;index:idx_created_at"`
	UpdatedAt          time.Time  `json:"updatedAt" gorm:"column:updatedAt;autoUpdateTime"`
	LastPasswordChange *time.Time `json:"lastPasswordChange,omitempty" gorm:"column:lastPasswordChange;type:datetime"`
}

// UserRole 用户角色枚举
type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// IsAdmin 检查用户是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// GetDisplayName 获取显示名称
func (u *User) GetDisplayName() string {
	if u.Nickname != "" {
		return u.Nickname
	}
	return u.Username
}

// AddPoints 增加积分（业务方法）
func (u *User) AddPoints(points int) {
	u.Points += points
	if u.Points < 0 {
		u.Points = 0
	}
}

// CanModifyResource 检查用户是否可以修改资源
func (u *User) CanModifyResource(resourceUserID uint) bool {
	return u.ID == resourceUserID || u.IsAdmin()
}

// IsValidRole 检查角色是否有效
func IsValidRole(role string) bool {
	return role == string(UserRoleUser) || role == string(UserRoleAdmin)
}

// LeaderboardEntry 排行榜条目
type LeaderboardEntry struct {
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	Points     int    `json:"points"`
	Rank       int    `json:"rank"`
	Tournament string `json:"tournament"`
}
