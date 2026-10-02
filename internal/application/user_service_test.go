package application

import (
	"context"
	"errors"
	"testing"

	"api.request.app.backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userRepoStub struct {
	user *domain.User
}

func (r *userRepoStub) Create(_ context.Context, user *domain.User) error {
	user.ID = uuid.New()
	r.user = user
	return nil
}

func (r *userRepoStub) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	if r.user == nil || r.user.ID != id {
		return nil, errors.New("user not found")
	}
	return r.user, nil
}

func (r *userRepoStub) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	if r.user == nil || r.user.Email != email {
		return nil, errors.New("user not found")
	}
	return r.user, nil
}

func (r *userRepoStub) Update(_ context.Context, user *domain.User) error {
	r.user = user
	return nil
}

func TestUserServiceHashesPasswordAndLogsIn(t *testing.T) {
	repo := &userRepoStub{}
	const secret = "test-secret"
	service := NewUserService(repo, secret)

	user, err := service.CreateUser(context.Background(), "alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if user.Password == "password123" {
		t.Fatal("password was stored in plain text")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123")); err != nil {
		t.Fatalf("stored password is not a valid hash: %v", err)
	}

	tokenString, loggedInUser, err := service.Login(context.Background(), user.Email, "password123")
	if err != nil {
		t.Fatal(err)
	}
	if loggedInUser.ID != user.ID {
		t.Fatal("login returned a different user")
	}
	token, err := jwt.Parse(tokenString, func(_ *jwt.Token) (any, error) { return []byte(secret), nil })
	if err != nil || !token.Valid {
		t.Fatalf("login returned an invalid token: %v", err)
	}

	if _, _, err := service.Login(context.Background(), user.Email, "wrong-password"); err == nil {
		t.Fatal("login accepted an invalid password")
	}
}

func TestUserServiceChangePassword(t *testing.T) {
	repo := &userRepoStub{}
	service := NewUserService(repo, "test-secret")
	user, err := service.CreateUser(context.Background(), "alice", "alice@example.com", "old-password")
	if err != nil {
		t.Fatal(err)
	}

	if err := service.ChangePassword(context.Background(), user.ID, "wrong-password", "new-password", "new-password"); !errors.Is(err, ErrInvalidCurrentPassword) {
		t.Fatalf("wrong current password returned %v", err)
	}
	if err := service.ChangePassword(context.Background(), user.ID, "old-password", "new-password", "different"); !errors.Is(err, ErrPasswordConfirmation) {
		t.Fatalf("mismatched confirmation returned %v", err)
	}
	if err := service.ChangePassword(context.Background(), user.ID, "old-password", "new-password", "new-password"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.Login(context.Background(), user.Email, "old-password"); err == nil {
		t.Fatal("old password still works")
	}
	if _, _, err := service.Login(context.Background(), user.Email, "new-password"); err != nil {
		t.Fatalf("new password does not work: %v", err)
	}
}
