package repository

import (
	"context"
	"encoding/json"

	"github.com/zqr233qr/story-trim/internal/errno"
	"github.com/zqr233qr/story-trim/internal/model"
	"gorm.io/gorm"
)

// PointsRepository 积分数据访问层。
type PointsRepository struct {
	db *gorm.DB
}

// NewPointsRepository 创建积分仓库。
func NewPointsRepository(db *gorm.DB) *PointsRepository {
	return &PointsRepository{db: db}
}

// GetUserPoints 获取用户积分余额记录。
func (r *PointsRepository) GetUserPoints(ctx context.Context, userID uint) (*model.UserPoints, error) {
	var points model.UserPoints
	exist, err := FirstRecodeIgnoreError(r.db.WithContext(ctx).Where("user_id = ?", userID), &points)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	return &points, nil
}

// CreateUserPoints 创建用户积分余额记录。
func (r *PointsRepository) CreateUserPoints(ctx context.Context, points *model.UserPoints) error {
	return r.db.WithContext(ctx).Create(points).Error
}

// PointsChange 积分变更请求。
type PointsChange struct {
	Change   int
	Type     string
	Reason   string
	RefType  string
	RefID    string
	ExtraMap map[string]string
}

// ChangeBalance 变更积分余额并写入流水。
func (r *PointsRepository) ChangeBalance(ctx context.Context, userID uint, change int, changeType, reason, refType, refID string, extra map[string]string) (int, error) {
	return r.ChangeBalanceBatch(ctx, userID, []PointsChange{{
		Change:   change,
		Type:     changeType,
		Reason:   reason,
		RefType:  refType,
		RefID:    refID,
		ExtraMap: extra,
	}})
}

// ChangeBalanceBatch 批量变更积分余额并写入流水。
func (r *PointsRepository) ChangeBalanceBatch(ctx context.Context, userID uint, changes []PointsChange) (int, error) {
	if len(changes) == 0 {
		return 0, nil
	}

	var balance int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var points model.UserPoints
		if err := tx.Where("user_id = ?", userID).First(&points).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				points = model.UserPoints{UserID: userID, Balance: 0}
				if err := tx.Create(&points).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		current := points.Balance
		ledgers := make([]model.PointsLedger, 0, len(changes))
		for _, change := range changes {
			nextBalance := current + change.Change
			if nextBalance < 0 {
				return errno.ErrPointsNotEnough
			}
			extraJSON := ""
			if change.ExtraMap != nil {
				if data, err := json.Marshal(change.ExtraMap); err == nil {
					extraJSON = string(data)
				}
			}
			ledgers = append(ledgers, model.PointsLedger{
				UserID:       userID,
				Change:       change.Change,
				BalanceAfter: nextBalance,
				Type:         change.Type,
				Reason:       change.Reason,
				RefType:      change.RefType,
				RefID:        change.RefID,
				Extra:        extraJSON,
			})
			current = nextBalance
		}

		if err := tx.Model(&model.UserPoints{}).Where("user_id = ?", userID).Update("balance", current).Error; err != nil {
			return err
		}
		if err := tx.Create(&ledgers).Error; err != nil {
			return err
		}

		balance = current
		return nil
	})

	if err != nil {
		return 0, err
	}
	return balance, nil
}

// ListLedger 获取积分流水。
func (r *PointsRepository) ListLedger(ctx context.Context, userID uint, limit, offset int) ([]model.PointsLedger, error) {
	var ledgers []model.PointsLedger
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&ledgers).Error; err != nil {
		return nil, err
	}
	return ledgers, nil
}

// PointsRepositoryInterface 积分仓库接口。
type PointsRepositoryInterface interface {
	GetUserPoints(ctx context.Context, userID uint) (*model.UserPoints, error)
	CreateUserPoints(ctx context.Context, points *model.UserPoints) error
	ChangeBalance(ctx context.Context, userID uint, change int, changeType, reason, refType, refID string, extra map[string]string) (int, error)
	ChangeBalanceBatch(ctx context.Context, userID uint, changes []PointsChange) (int, error)
	ListLedger(ctx context.Context, userID uint, limit, offset int) ([]model.PointsLedger, error)
}
