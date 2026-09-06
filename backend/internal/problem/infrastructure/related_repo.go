package infrastructure

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/team/pkb/internal/problem/domain"
)

// RelatedRepository implements domain.RelatedRepository.
type RelatedRepository struct {
	db *gorm.DB
}

// NewRelatedRepository builds the related-links repository.
func NewRelatedRepository(db *gorm.DB) *RelatedRepository { return &RelatedRepository{db: db} }

func (r *RelatedRepository) List(ctx context.Context, problemID int64) ([]domain.RelatedLink, error) {
	var models []relatedModel
	if err := r.db.WithContext(ctx).
		Where("problem_id = ?", problemID).
		Order("created_at ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	if len(models) == 0 {
		return []domain.RelatedLink{}, nil
	}

	ids := make([]int64, len(models))
	for i, m := range models {
		ids[i] = m.RelatedProblemID
	}

	var rows []joinedListRow
	_ = r.db.WithContext(ctx).
		Table("problems AS p").
		Joins("JOIN problem_categories c ON c.id = p.category_id").
		Select(selectColumns+", 0 AS rank, '' AS headline").
		Where("p.id IN ?", ids).
		Scan(&rows).Error

	snapshot := make(map[int64]domain.ListItem, len(rows))
	for _, row := range rows {
		snapshot[row.ID] = row.toListItem()
	}

	out := make([]domain.RelatedLink, len(models))
	for i, m := range models {
		link := domain.RelatedLink{
			ID:           m.ID,
			ProblemID:    m.ProblemID,
			RelatedID:    m.RelatedProblemID,
			RelationType: m.RelationType,
			CreatedAt:    m.CreatedAt,
		}
		if snap, ok := snapshot[m.RelatedProblemID]; ok {
			s := snap
			link.Related = &s
		}
		out[i] = link
	}
	return out, nil
}

func (r *RelatedRepository) Add(ctx context.Context, problemID, relatedID int64, relationType string) (*domain.RelatedLink, error) {
	m := relatedModel{ProblemID: problemID, RelatedProblemID: relatedID, RelationType: relationType}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		msg := err.Error()
		if strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint") {
			return nil, domain.ErrRelationExists
		}
		if strings.Contains(msg, "problem_id <> related_problem_id") || strings.Contains(msg, "check constraint") {
			return nil, domain.ErrSelfRelation
		}
		return nil, err
	}
	link := domain.RelatedLink{
		ID:           m.ID,
		ProblemID:    m.ProblemID,
		RelatedID:    m.RelatedProblemID,
		RelationType: m.RelationType,
		CreatedAt:    m.CreatedAt,
	}
	return &link, nil
}

func (r *RelatedRepository) Delete(ctx context.Context, problemID, relatedID int64) error {
	res := r.db.WithContext(ctx).
		Where("problem_id = ? AND related_problem_id = ?", problemID, relatedID).
		Delete(&relatedModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrRelatedNotFound
	}
	return nil
}
