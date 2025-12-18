package ports

import (
	"context"

	"backend-go/internal/core/domain/team"
)

// TeamRepository 战队仓储接口
type TeamRepository interface {
	Create(ctx context.Context, t *team.Team) (*team.Team, error)
	Update(ctx context.Context, id uint, t *team.Team) (*team.Team, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*team.Team, error)
	List(ctx context.Context, includeInactive bool) ([]team.Team, error)
}

// TeamService 战队服务接口
type TeamService interface {
	CreateTeam(ctx context.Context, req *CreateTeamRequest) (*team.Team, error)
	UpdateTeam(ctx context.Context, id uint, req *UpdateTeamRequest) (*team.Team, error)
	DeleteTeam(ctx context.Context, id uint) error
	GetTeam(ctx context.Context, id uint) (*team.Team, error)
	ListTeams(ctx context.Context, includeInactive bool) ([]team.Team, error)
}

// CreateTeamRequest 创建战队请求
type CreateTeamRequest struct {
	Name      string
	ShortName string
	LogoURL   string
	IsActive  bool
}

// UpdateTeamRequest 更新战队请求
type UpdateTeamRequest struct {
	Name      string
	ShortName string
	LogoURL   string
	IsActive  bool
}
