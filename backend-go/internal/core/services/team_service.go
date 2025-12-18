package services

import (
	"context"
	"fmt"
	"strings"

	"backend-go/internal/core/domain/team"
	"backend-go/internal/core/ports"
)

// TeamService 战队服务实现
type TeamService struct {
	repo ports.TeamRepository
}

// NewTeamService 创建战队服务
func NewTeamService(repo ports.TeamRepository) *TeamService {
	return &TeamService{repo: repo}
}

// CreateTeam 创建战队
func (s *TeamService) CreateTeam(ctx context.Context, req *ports.CreateTeamRequest) (*team.Team, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("战队名称不能为空")
	}
	t := &team.Team{
		Name:      strings.TrimSpace(req.Name),
		ShortName: strings.TrimSpace(req.ShortName),
		LogoURL:   strings.TrimSpace(req.LogoURL),
		IsActive:  req.IsActive,
	}
	return s.repo.Create(ctx, t)
}

// UpdateTeam 更新战队
func (s *TeamService) UpdateTeam(ctx context.Context, id uint, req *ports.UpdateTeamRequest) (*team.Team, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("战队名称不能为空")
	}
	t := &team.Team{
		Name:      strings.TrimSpace(req.Name),
		ShortName: strings.TrimSpace(req.ShortName),
		LogoURL:   strings.TrimSpace(req.LogoURL),
		IsActive:  req.IsActive,
	}
	return s.repo.Update(ctx, id, t)
}

// DeleteTeam 删除战队
func (s *TeamService) DeleteTeam(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

// GetTeam 获取战队
func (s *TeamService) GetTeam(ctx context.Context, id uint) (*team.Team, error) {
	return s.repo.GetByID(ctx, id)
}

// ListTeams 列出战队
func (s *TeamService) ListTeams(ctx context.Context, includeInactive bool) ([]team.Team, error) {
	return s.repo.List(ctx, includeInactive)
}
