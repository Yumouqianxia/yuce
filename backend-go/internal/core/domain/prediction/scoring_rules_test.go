package prediction

import (
	"testing"
	"time"

	"backend-go/internal/core/domain"
)

func TestScoringRule_CalculatePoints(t *testing.T) {
	// 创建测试用的积分规则
	rule := &ScoringRule{
		ID:                      1,
		Name:                    "测试规则",
		CorrectTeamCorrectScore: 50,
		CorrectTeamWrongScore:   20,
		WrongTeamCorrectScore:   10,
		WrongTeamWrongScore:     5,
		IsActive:                true,
	}

	// 创建测试用的比赛
	match := &domain.Match{
		ID:     1,
		TeamA:  "Team A",
		TeamB:  "Team B",
		Status: domain.MatchStatusFinished,
		Winner: "A",
		ScoreA: 2,
		ScoreB: 1,
	}

	tests := []struct {
		name           string
		prediction     *Prediction
		expectedPoints int
		description    string
	}{
		{
			name: "预测正确队伍和比分",
			prediction: &Prediction{
				ID:              1,
				PredictedWinner: "A",
				PredictedScoreA: 2,
				PredictedScoreB: 1,
				Match:           match,
			},
			expectedPoints: 50,
			description:    "队伍和比分都正确，应该获得最高分",
		},
		{
			name: "预测正确队伍错误比分",
			prediction: &Prediction{
				ID:              2,
				PredictedWinner: "A",
				PredictedScoreA: 3,
				PredictedScoreB: 0,
				Match:           match,
			},
			expectedPoints: 20,
			description:    "队伍正确但比分错误，应该获得中等分数",
		},
		{
			name: "预测错误队伍正确比分",
			prediction: &Prediction{
				ID:              3,
				PredictedWinner: "B",
				PredictedScoreA: 2,
				PredictedScoreB: 1,
				Match:           match,
			},
			expectedPoints: 10,
			description:    "队伍错误但比分正确，应该获得较低分数",
		},
		{
			name: "预测错误队伍错误比分",
			prediction: &Prediction{
				ID:              4,
				PredictedWinner: "B",
				PredictedScoreA: 1,
				PredictedScoreB: 3,
				Match:           match,
			},
			expectedPoints: 5,
			description:    "队伍和比分都错误，应该获得最低分数",
		},
		{
			name: "比赛未结束",
			prediction: &Prediction{
				ID:              5,
				PredictedWinner: "A",
				PredictedScoreA: 2,
				PredictedScoreB: 1,
				Match: &domain.Match{
					ID:     2,
					Status: domain.MatchStatusLive,
				},
			},
			expectedPoints: 0,
			description:    "比赛未结束，不应该计算积分",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			points := rule.CalculatePoints(tt.prediction)
			if points != tt.expectedPoints {
				t.Errorf("CalculatePoints() = %d, want %d. %s", points, tt.expectedPoints, tt.description)
			}
		})
	}
}

func TestPrediction_CalculatePointsWithRule(t *testing.T) {
	// 创建测试用的积分规则
	rule := &ScoringRule{
		ID:                      1,
		Name:                    "测试规则",
		CorrectTeamCorrectScore: 100,
		CorrectTeamWrongScore:   30,
		WrongTeamCorrectScore:   15,
		WrongTeamWrongScore:     0,
		IsActive:                true,
	}

	// 创建测试用的比赛
	match := &domain.Match{
		ID:     1,
		TeamA:  "Team A",
		TeamB:  "Team B",
		Status: domain.MatchStatusFinished,
		Winner: "A",
		ScoreA: 3,
		ScoreB: 1,
	}

	tests := []struct {
		name            string
		prediction      *Prediction
		voteCount       int
		expectedPoints  int
		expectedCorrect bool
		description     string
	}{
		{
			name: "完全正确预测，有热门奖励",
			prediction: &Prediction{
				ID:              1,
				PredictedWinner: "A",
				PredictedScoreA: 3,
				PredictedScoreB: 1,
				VoteCount:       15,
				Match:           match,
			},
			expectedPoints:  105, // 100 + 5 (热门奖励)
			expectedCorrect: true,
			description:     "完全正确且有热门奖励",
		},
		{
			name: "队伍正确比分错误，无热门奖励",
			prediction: &Prediction{
				ID:              2,
				PredictedWinner: "A",
				PredictedScoreA: 2,
				PredictedScoreB: 0,
				VoteCount:       5,
				Match:           match,
			},
			expectedPoints:  30, // 30 + 0 (无热门奖励)
			expectedCorrect: true,
			description:     "队伍正确但比分错误，无热门奖励",
		},
		{
			name: "队伍错误比分正确，有热门奖励",
			prediction: &Prediction{
				ID:              3,
				PredictedWinner: "B",
				PredictedScoreA: 3,
				PredictedScoreB: 1,
				VoteCount:       12,
				Match:           match,
			},
			expectedPoints:  20,   // 15 + 5 (热门奖励)
			expectedCorrect: true, // 比分正确也算正确
			description:     "队伍错误但比分正确，有热门奖励",
		},
		{
			name: "完全错误预测",
			prediction: &Prediction{
				ID:              4,
				PredictedWinner: "B",
				PredictedScoreA: 1,
				PredictedScoreB: 2,
				VoteCount:       3,
				Match:           match,
			},
			expectedPoints:  0, // 0 + 0 (无热门奖励)
			expectedCorrect: false,
			description:     "完全错误预测",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			points := tt.prediction.CalculatePointsWithRule(rule)
			if points != tt.expectedPoints {
				t.Errorf("CalculatePointsWithRule() = %d, want %d. %s", points, tt.expectedPoints, tt.description)
			}
			if tt.prediction.IsCorrect != tt.expectedCorrect {
				t.Errorf("IsCorrect = %v, want %v. %s", tt.prediction.IsCorrect, tt.expectedCorrect, tt.description)
			}
		})
	}
}

func TestPrediction_CalculatePoints_BackwardCompatibility(t *testing.T) {
	// 测试向后兼容性：确保原有的积分计算逻辑仍然有效
	match := &domain.Match{
		ID:     1,
		TeamA:  "Team A",
		TeamB:  "Team B",
		Status: domain.MatchStatusFinished,
		Winner: "A",
		ScoreA: 2,
		ScoreB: 1,
	}

	tests := []struct {
		name            string
		prediction      *Prediction
		expectedPoints  int
		expectedCorrect bool
	}{
		{
			name: "原有逻辑：完全正确预测",
			prediction: &Prediction{
				ID:              1,
				PredictedWinner: "A",
				PredictedScoreA: 2,
				PredictedScoreB: 1,
				VoteCount:       15,
				Match:           match,
			},
			expectedPoints:  35, // 10 + 20 + 5
			expectedCorrect: true,
		},
		{
			name: "原有逻辑：队伍正确比分错误",
			prediction: &Prediction{
				ID:              2,
				PredictedWinner: "A",
				PredictedScoreA: 3,
				PredictedScoreB: 0,
				VoteCount:       5,
				Match:           match,
			},
			expectedPoints:  10, // 10 + 0 + 0
			expectedCorrect: true,
		},
		{
			name: "原有逻辑：完全错误预测",
			prediction: &Prediction{
				ID:              3,
				PredictedWinner: "B",
				PredictedScoreA: 1,
				PredictedScoreB: 3,
				VoteCount:       12,
				Match:           match,
			},
			expectedPoints:  5, // 0 + 0 + 5
			expectedCorrect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置IsCorrect状态
			tt.prediction.IsCorrect = false

			points := tt.prediction.CalculatePoints()
			if points != tt.expectedPoints {
				t.Errorf("CalculatePoints() = %d, want %d", points, tt.expectedPoints)
			}
			if tt.prediction.IsCorrect != tt.expectedCorrect {
				t.Errorf("IsCorrect = %v, want %v", tt.prediction.IsCorrect, tt.expectedCorrect)
			}
		})
	}
}

func TestScoringRule_TableName(t *testing.T) {
	rule := &ScoringRule{}
	expected := "scoring_rules"
	if got := rule.TableName(); got != expected {
		t.Errorf("TableName() = %v, want %v", got, expected)
	}
}

func TestScoringRule_Creation(t *testing.T) {
	rule := &ScoringRule{
		Name:                    "测试规则",
		Description:             "这是一个测试规则",
		CorrectTeamCorrectScore: 50,
		CorrectTeamWrongScore:   20,
		WrongTeamCorrectScore:   10,
		WrongTeamWrongScore:     5,
		IsActive:                true,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}

	if rule.Name != "测试规则" {
		t.Errorf("Name = %v, want %v", rule.Name, "测试规则")
	}
	if rule.CorrectTeamCorrectScore != 50 {
		t.Errorf("CorrectTeamCorrectScore = %v, want %v", rule.CorrectTeamCorrectScore, 50)
	}
	if !rule.IsActive {
		t.Errorf("IsActive = %v, want %v", rule.IsActive, true)
	}
}
