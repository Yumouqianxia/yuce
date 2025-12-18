package mysql

import (
	"backend-go/internal/core/domain/match"
	"backend-go/internal/core/domain/user"
	"backend-go/internal/shared/password"
	"gorm.io/gorm"
)

// RepositoryFactory MySQL 仓储工厂
type RepositoryFactory struct {
	db              *gorm.DB
	passwordService password.Service
}

// NewRepositoryFactory 创建仓储工厂
func NewRepositoryFactory(db *gorm.DB, passwordService password.Service) *RepositoryFactory {
	return &RepositoryFactory{
		db:              db,
		passwordService: passwordService,
	}
}

// NewUserRepository 创建用户仓储
func (f *RepositoryFactory) NewUserRepository() user.Repository {
	return NewUserRepository(f.db, f.passwordService)
}

// NewMatchRepository 创建比赛仓储
func (f *RepositoryFactory) NewMatchRepository() match.Repository {
	return NewMatchRepository(f.db)
}

// GetDB 获取数据库连接
func (f *RepositoryFactory) GetDB() *gorm.DB {
	return f.db
}

// MigrateAll 迁移所有表
func (f *RepositoryFactory) MigrateAll() error {
	// 使用 AutoMigrate 迁移所有表
	return AutoMigrate(f.db)
}

// ValidateAll 验证所有表结构
func (f *RepositoryFactory) ValidateAll() error {
	// 简单的连接测试
	sqlDB, err := f.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
