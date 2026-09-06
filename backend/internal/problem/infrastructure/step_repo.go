package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/team/pkb/internal/problem/domain"
)

// StepRepository implements domain.StepRepository.
type StepRepository struct {
	db *gorm.DB
}

// NewStepRepository builds the step repository.
func NewStepRepository(db *gorm.DB) *StepRepository { return &StepRepository{db: db} }

func (r *StepRepository) List(ctx context.Context, problemID int64) ([]domain.Step, error) {
	var models []stepModel
	if err := r.db.WithContext(ctx).
		Where("problem_id = ?", problemID).
		Order("step_no ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Step, len(models))
	for i, m := range models {
		out[i] = m.toDomain()
	}
	return out, nil
}

func (r *StepRepository) Append(ctx context.Context, problemID int64, action, result string) (*domain.Step, error) {
	var next int
	if err := r.db.WithContext(ctx).
		Raw(`SELECT coalesce(max(step_no), 0) + 1 FROM problem_steps WHERE problem_id = ?`, problemID).
		Scan(&next).Error; err != nil {
		return nil, err
	}
	m := stepModel{ProblemID: problemID, StepNo: next, Action: action, Result: result}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return nil, err
	}
	s := m.toDomain()
	return &s, nil
}

func (r *StepRepository) Update(ctx context.Context, problemID, stepID int64, action, result string, stepNo *int) (*domain.Step, error) {
	var updated stepModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current stepModel
		if err := tx.Where("id = ? AND problem_id = ?", stepID, problemID).First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrStepNotFound
			}
			return err
		}

		if stepNo != nil && *stepNo != current.StepNo {
			if err := moveStep(tx, problemID, current.StepNo, *stepNo); err != nil {
				return err
			}
		}

		fields := map[string]any{"action": action, "result": result, "updated_at": gorm.Expr("now()")}
		if err := tx.Model(&stepModel{}).Where("id = ?", stepID).Updates(fields).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", stepID).First(&updated).Error
	})
	if err != nil {
		return nil, err
	}
	s := updated.toDomain()
	return &s, nil
}

func (r *StepRepository) Delete(ctx context.Context, problemID, stepID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("id = ? AND problem_id = ?", stepID, problemID).Delete(&stepModel{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domain.ErrStepNotFound
		}
		// Compact the remaining step numbers: 1..N with no gaps.
		return renumber(tx, problemID)
	})
}

func (r *StepRepository) Reorder(ctx context.Context, problemID int64, orderedIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing []int64
		if err := tx.Model(&stepModel{}).
			Where("problem_id = ?", problemID).
			Pluck("id", &existing).Error; err != nil {
			return err
		}
		if !sameSet(existing, orderedIDs) {
			return domain.ErrStepNotFound
		}

		// Two-phase to dodge the UNIQUE(problem_id, step_no) constraint.
		if err := tx.Model(&stepModel{}).
			Where("problem_id = ?", problemID).
			UpdateColumn("step_no", gorm.Expr("step_no + 100000")).Error; err != nil {
			return err
		}
		for i, id := range orderedIDs {
			if err := tx.Model(&stepModel{}).
				Where("id = ? AND problem_id = ?", id, problemID).
				UpdateColumn("step_no", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// moveStep shifts step numbers so a row can move from -> to (both 1-based).
func moveStep(tx *gorm.DB, problemID int64, from, to int) error {
	var count int64
	if err := tx.Model(&stepModel{}).Where("problem_id = ?", problemID).Count(&count).Error; err != nil {
		return err
	}
	if to < 1 {
		to = 1
	}
	if to > int(count) {
		to = int(count)
	}
	if from == to {
		return nil
	}

	// Park the moving row out of the way.
	if err := tx.Model(&stepModel{}).
		Where("problem_id = ? AND step_no = ?", problemID, from).
		UpdateColumn("step_no", 1000000).Error; err != nil {
		return err
	}
	if from < to {
		if err := tx.Model(&stepModel{}).
			Where("problem_id = ? AND step_no > ? AND step_no <= ?", problemID, from, to).
			UpdateColumn("step_no", gorm.Expr("step_no - 1")).Error; err != nil {
			return err
		}
	} else {
		if err := tx.Model(&stepModel{}).
			Where("problem_id = ? AND step_no >= ? AND step_no < ?", problemID, to, from).
			UpdateColumn("step_no", gorm.Expr("step_no + 1")).Error; err != nil {
			return err
		}
	}
	return tx.Model(&stepModel{}).
		Where("problem_id = ? AND step_no = ?", problemID, 1000000).
		UpdateColumn("step_no", to).Error
}

// renumber rewrites step_no to a gapless 1..N sequence ordered by current step_no.
func renumber(tx *gorm.DB, problemID int64) error {
	var ids []int64
	if err := tx.Model(&stepModel{}).
		Where("problem_id = ?", problemID).
		Order("step_no ASC").
		Pluck("id", &ids).Error; err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	if err := tx.Model(&stepModel{}).
		Where("problem_id = ?", problemID).
		UpdateColumn("step_no", gorm.Expr("step_no + 100000")).Error; err != nil {
		return err
	}
	for i, id := range ids {
		if err := tx.Model(&stepModel{}).Where("id = ?", id).UpdateColumn("step_no", i+1).Error; err != nil {
			return err
		}
	}
	return nil
}

func sameSet(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[int64]int, len(a))
	for _, v := range a {
		m[v]++
	}
	for _, v := range b {
		m[v]--
		if m[v] < 0 {
			return false
		}
	}
	return true
}
