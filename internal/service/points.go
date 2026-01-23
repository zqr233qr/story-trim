package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zqr233qr/story-trim/internal/model"
	"github.com/zqr233qr/story-trim/internal/repository"
)

const (
	pointsTypeEarn  = "earn"
	pointsTypeSpend = "spend"
)

const (
	pointsReasonRegister = "register_bonus"
	pointsReasonTrimUse  = "trim_use"
	pointsReasonRefund   = "trim_refund"
)

// PointsService 积分服务。
type PointsService struct {
	repo repository.PointsRepositoryInterface
}

// PointsChangeInput 积分变更输入。
type PointsChangeInput struct {
	RefType string
	RefID   string
	Extra   map[string]string
}

// NewPointsService 创建积分服务。
func NewPointsService(repo repository.PointsRepositoryInterface) *PointsService {
	return &PointsService{repo: repo}
}

// GrantRegisterBonus 注册赠送积分。
func (s *PointsService) GrantRegisterBonus(ctx context.Context, userID uint, amount int) error {
	_, err := s.repo.ChangeBalance(ctx, userID, amount, pointsTypeEarn, pointsReasonRegister, "user", fmt.Sprintf("%d", userID), nil)
	return err
}

// SpendForTrim 扣除精简积分。
func (s *PointsService) SpendForTrim(ctx context.Context, userID uint, count int, refType, refID string, extra map[string]string) error {
	if count <= 0 {
		return nil
	}
	_, err := s.repo.ChangeBalance(ctx, userID, -count, pointsTypeSpend, pointsReasonTrimUse, refType, refID, extra)
	return err
}

// SpendForTrimBatch 批量扣除精简积分。
func (s *PointsService) SpendForTrimBatch(ctx context.Context, userID uint, entries []PointsChangeInput) error {
	if len(entries) == 0 {
		return nil
	}
	changes := make([]repository.PointsChange, 0, len(entries))
	for _, entry := range entries {
		changes = append(changes, repository.PointsChange{
			Change:   -1,
			Type:     pointsTypeSpend,
			Reason:   pointsReasonTrimUse,
			RefType:  entry.RefType,
			RefID:    entry.RefID,
			ExtraMap: entry.Extra,
		})
	}
	_, err := s.repo.ChangeBalanceBatch(ctx, userID, changes)
	return err
}

// RefundForTrim 退还精简积分。
func (s *PointsService) RefundForTrim(ctx context.Context, userID uint, count int, refType, refID string, extra map[string]string) error {
	if count <= 0 {
		return nil
	}
	_, err := s.repo.ChangeBalance(ctx, userID, count, pointsTypeEarn, pointsReasonRefund, refType, refID, extra)
	return err
}

// RefundForTrimBatch 批量退还精简积分。
func (s *PointsService) RefundForTrimBatch(ctx context.Context, userID uint, entries []PointsChangeInput) error {
	if len(entries) == 0 {
		return nil
	}
	changes := make([]repository.PointsChange, 0, len(entries))
	for _, entry := range entries {
		changes = append(changes, repository.PointsChange{
			Change:   1,
			Type:     pointsTypeEarn,
			Reason:   pointsReasonRefund,
			RefType:  entry.RefType,
			RefID:    entry.RefID,
			ExtraMap: entry.Extra,
		})
	}
	_, err := s.repo.ChangeBalanceBatch(ctx, userID, changes)
	return err
}

// GetBalance 获取积分余额。
func (s *PointsService) GetBalance(ctx context.Context, userID uint) (int, error) {
	points, err := s.repo.GetUserPoints(ctx, userID)
	if err != nil {
		return 0, err
	}
	if points == nil {
		return 0, nil
	}
	return points.Balance, nil
}

// ListLedger 获取积分流水。
func (s *PointsService) ListLedger(ctx context.Context, userID uint, page, size int) ([]PointsLedgerEntry, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	limit := size
	offset := (page - 1) * size

	ledgers, err := s.repo.ListLedger(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	items := make([]PointsLedgerEntry, 0, len(ledgers))
	for _, ledger := range ledgers {
		items = append(items, BuildPointsLedgerEntry(ledger))
	}
	return items, nil
}

// PointsServiceInterface 积分服务接口。
type PointsServiceInterface interface {
	GrantRegisterBonus(ctx context.Context, userID uint, amount int) error
	SpendForTrim(ctx context.Context, userID uint, count int, refType, refID string, extra map[string]string) error
	SpendForTrimBatch(ctx context.Context, userID uint, entries []PointsChangeInput) error
	RefundForTrim(ctx context.Context, userID uint, count int, refType, refID string, extra map[string]string) error
	RefundForTrimBatch(ctx context.Context, userID uint, entries []PointsChangeInput) error
	GetBalance(ctx context.Context, userID uint) (int, error)
	ListLedger(ctx context.Context, userID uint, page, size int) ([]PointsLedgerEntry, error)
}

// PointsLedgerEntry 积分流水简要信息。
type PointsLedgerEntry struct {
	ID           uint              `json:"id"`
	Change       int               `json:"change"`
	BalanceAfter int               `json:"balance_after"`
	Type         string            `json:"type"`
	Reason       string            `json:"reason"`
	RefType      string            `json:"ref_type"`
	RefID        string            `json:"ref_id"`
	Extra        map[string]string `json:"extra,omitempty"`
	CreatedAt    string            `json:"created_at"`
}

// PointsBalanceResponse 积分余额返回结构。
type PointsBalanceResponse struct {
	Balance int `json:"balance"`
}

// BuildPointsLedgerEntry 构造积分流水返回。
func BuildPointsLedgerEntry(item model.PointsLedger) PointsLedgerEntry {
	extra := map[string]string{}
	if item.Extra != "" {
		if err := json.Unmarshal([]byte(item.Extra), &extra); err != nil {
			extra = map[string]string{}
		}
	}
	return PointsLedgerEntry{
		ID:           item.ID,
		Change:       item.Change,
		BalanceAfter: item.BalanceAfter,
		Type:         item.Type,
		Reason:       item.Reason,
		RefType:      item.RefType,
		RefID:        item.RefID,
		Extra:        extra,
		CreatedAt:    item.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
