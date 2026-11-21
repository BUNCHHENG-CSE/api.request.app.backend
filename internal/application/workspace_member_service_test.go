package application

import (
	"context"
	"errors"
	"testing"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type workspaceMemberRepoMemory struct {
	members map[string]*domain.WorkspaceMember
}

func newWorkspaceMemberRepoMemory() *workspaceMemberRepoMemory {
	return &workspaceMemberRepoMemory{members: make(map[string]*domain.WorkspaceMember)}
}

func (r *workspaceMemberRepoMemory) key(wsID, uID uuid.UUID) string {
	return wsID.String() + ":" + uID.String()
}

func (r *workspaceMemberRepoMemory) Add(_ context.Context, m *domain.WorkspaceMember) error {
	k := r.key(m.WorkspaceID, m.UserID)
	copy := *m
	r.members[k] = &copy
	return nil
}

func (r *workspaceMemberRepoMemory) Get(_ context.Context, wsID, uID uuid.UUID) (*domain.WorkspaceMember, error) {
	k := r.key(wsID, uID)
	m, ok := r.members[k]
	if !ok || m == nil {
		return nil, errors.New("member not found")
	}
	return m, nil
}

func (r *workspaceMemberRepoMemory) List(_ context.Context, wsID uuid.UUID) ([]domain.WorkspaceMember, error) {
	var list []domain.WorkspaceMember
	for _, m := range r.members {
		if m.WorkspaceID == wsID {
			list = append(list, *m)
		}
	}
	return list, nil
}

func (r *workspaceMemberRepoMemory) Update(_ context.Context, m *domain.WorkspaceMember) error {
	k := r.key(m.WorkspaceID, m.UserID)
	copy := *m
	r.members[k] = &copy
	return nil
}

func (r *workspaceMemberRepoMemory) Delete(_ context.Context, wsID, uID uuid.UUID) error {
	k := r.key(wsID, uID)
	delete(r.members, k)
	return nil
}

type userRepoMem struct {
	users map[uuid.UUID]*domain.User
}

func (r *userRepoMem) Create(_ context.Context, user *domain.User) error {
	user.ID = uuid.New()
	r.users[user.ID] = user
	return nil
}

func (r *userRepoMem) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (r *userRepoMem) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *userRepoMem) GetByUsernameOrEmail(_ context.Context, idStr string) (*domain.User, error) {
	for _, u := range r.users {
		if u.Email == idStr || u.Username == idStr {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *userRepoMem) Update(context.Context, *domain.User) error {
	return nil
}

func TestWorkspaceMemberService(t *testing.T) {
	ownerID := uuid.New()
	wsID := uuid.New()

	wsRepo := &workspaceRepoMemory{workspace: &domain.Workspace{ID: wsID, OwnerID: ownerID, Name: "Test WS"}}
	memberRepo := newWorkspaceMemberRepoMemory()
	uRepo := &userRepoMem{users: make(map[uuid.UUID]*domain.User)}

	collabUser := &domain.User{ID: uuid.New(), Username: "collab", Email: "collab@example.com"}
	uRepo.users[collabUser.ID] = collabUser

	svc := NewWorkspaceMemberService(wsRepo, memberRepo, uRepo)

	t.Run("owner can add member", func(t *testing.T) {
		member, err := svc.AddMember(context.Background(), ownerID, wsID, "collab@example.com", domain.RoleEditor)
		require.NoError(t, err)
		assert.Equal(t, collabUser.ID, member.UserID)
		assert.Equal(t, domain.RoleEditor, member.Role)
	})

	t.Run("cannot add already added member", func(t *testing.T) {
		_, err := svc.AddMember(context.Background(), ownerID, wsID, "collab", domain.RoleEditor)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already a member")
	})

	t.Run("list members returns added member", func(t *testing.T) {
		members, err := svc.ListMembers(context.Background(), ownerID, wsID)
		require.NoError(t, err)
		assert.Len(t, members, 1)
		assert.Equal(t, collabUser.ID, members[0].UserID)
	})

	t.Run("owner can update role", func(t *testing.T) {
		updated, err := svc.UpdateRole(context.Background(), ownerID, wsID, collabUser.ID, domain.RoleAdmin)
		require.NoError(t, err)
		assert.Equal(t, domain.RoleAdmin, updated.Role)
	})

	t.Run("non-admin cannot add member", func(t *testing.T) {
		strangerID := uuid.New()
		_, err := svc.AddMember(context.Background(), strangerID, wsID, "collab@example.com", domain.RoleViewer)
		require.Error(t, err)
	})

	t.Run("owner can remove member", func(t *testing.T) {
		err := svc.RemoveMember(context.Background(), ownerID, wsID, collabUser.ID)
		require.NoError(t, err)

		members, err := svc.ListMembers(context.Background(), ownerID, wsID)
		require.NoError(t, err)
		assert.Len(t, members, 0)
	})
}
