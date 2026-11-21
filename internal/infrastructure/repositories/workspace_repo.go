package repositories

import (
	"context"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"

	"gorm.io/gorm"
)

type workspaceRepo struct {
	db *gorm.DB
}

func NewWorkspaceRepository(db *gorm.DB) domain.WorkspaceRepository {
	return &workspaceRepo{db: db}
}

func (r *workspaceRepo) Create(ctx context.Context, workspace *domain.Workspace) error {
	return r.db.WithContext(ctx).Create(workspace).Error
}

func (r *workspaceRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Workspace, error) {
	var workspace domain.Workspace
	err := r.db.WithContext(ctx).
		Table("workspaces w").
		Select("w.*").
		Joins("LEFT JOIN workspace_members m ON m.workspace_id = w.id").
		Where("w.id = ? AND (w.owner_id = ? OR m.user_id = ?)", id, userID, userID).
		First(&workspace).Error
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

func (r *workspaceRepo) List(ctx context.Context, userID uuid.UUID) ([]domain.Workspace, error) {
	var workspaces []domain.Workspace
	err := r.db.WithContext(ctx).
		Table("workspaces w").
		Distinct("w.id", "w.name", "w.owner_id", "w.created_at", "w.updated_at").
		Joins("LEFT JOIN workspace_members m ON m.workspace_id = w.id").
		Where("w.owner_id = ? OR m.user_id = ?", userID, userID).
		Order("w.created_at ASC").
		Find(&workspaces).Error
	return workspaces, err
}

func (r *workspaceRepo) Update(ctx context.Context, workspace *domain.Workspace) error {
	return r.db.WithContext(ctx).Save(workspace).Error
}

func (r *workspaceRepo) Delete(ctx context.Context, id, ownerID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&domain.Workspace{}).Where("id = ? AND owner_id = ?", id, ownerID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Where("flow_id IN (?)", tx.Model(&domain.Flow{}).Select("id").Where("workspace_id = ?", id)).Delete(&domain.Request{}).Error; err != nil {
			return err
		}
		if err := tx.Where("workspace_id = ?", id).Delete(&domain.Flow{}).Error; err != nil {
			return err
		}
		if err := tx.Where("workspace_id = ?", id).Delete(&domain.Environment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("workspace_id = ?", id).Delete(&domain.Collection{}).Error; err != nil {
			return err
		}
		if err := tx.Where("workspace_id = ?", id).Delete(&domain.RequestHistory{}).Error; err != nil {
			return err
		}
		if err := tx.Where("workspace_id = ?", id).Delete(&domain.WorkspaceMember{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND owner_id = ?", id, ownerID).Delete(&domain.Workspace{}).Error
	})
}

type workspaceMemberRepo struct {
	db *gorm.DB
}

func NewWorkspaceMemberRepository(db *gorm.DB) domain.WorkspaceMemberRepository {
	return &workspaceMemberRepo{db: db}
}

func (r *workspaceMemberRepo) Add(ctx context.Context, member *domain.WorkspaceMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *workspaceMemberRepo) Get(ctx context.Context, workspaceID, userID uuid.UUID) (*domain.WorkspaceMember, error) {
	var member domain.WorkspaceMember
	err := r.db.WithContext(ctx).Where("workspace_id = ? AND user_id = ?", workspaceID, userID).First(&member).Error
	return &member, err
}

func (r *workspaceMemberRepo) List(ctx context.Context, workspaceID uuid.UUID) ([]domain.WorkspaceMember, error) {
	var members []domain.WorkspaceMember
	err := r.db.WithContext(ctx).
		Table("workspace_members wm").
		Select("wm.*, u.username, u.email").
		Joins("JOIN users u ON u.id = wm.user_id").
		Where("wm.workspace_id = ?", workspaceID).
		Order("wm.created_at ASC").
		Find(&members).Error
	return members, err
}

func (r *workspaceMemberRepo) Update(ctx context.Context, member *domain.WorkspaceMember) error {
	return r.db.WithContext(ctx).Model(member).Where("workspace_id = ? AND user_id = ?", member.WorkspaceID, member.UserID).Update("role", member.Role).Error
}

func (r *workspaceMemberRepo) Delete(ctx context.Context, workspaceID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("workspace_id = ? AND user_id = ?", workspaceID, userID).Delete(&domain.WorkspaceMember{}).Error
}
