package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"api.request.app.backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(ctx context.Context, username, email, password string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, *domain.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, username string) (*domain.User, error)
	RefreshToken(ctx context.Context, id uuid.UUID) (string, error)
	ChangePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword, confirmPassword string) error
}

var (
	ErrInvalidCurrentPassword = errors.New("current password is incorrect")
	ErrPasswordConfirmation   = errors.New("new password and confirmation do not match")
	ErrEmailAlreadyInUse      = errors.New("email already in use")
	ErrUserNotFound           = errors.New("user not found")
	ErrInvalidCredentials     = errors.New("invalid email or password")
)

type userService struct {
	repo      domain.UserRepository
	jwtSecret []byte
}

func NewUserService(repo domain.UserRepository, jwtSecret string) UserService {
	return &userService{repo: repo, jwtSecret: []byte(jwtSecret)}
}

func (s *userService) CreateUser(ctx context.Context, username, email, password string) (*domain.User, error) {
	username = strings.TrimSpace(username)
	email = strings.ToLower(strings.TrimSpace(email))

	if username == "" || email == "" || password == "" {
		return nil, errors.New("username, email, and password are required")
	}
	if len(password) < 6 {
		return nil, errors.New("password must contain at least 6 characters")
	}

	// Check if user already exists
	existingUser, _ := s.repo.GetByEmail(ctx, email)
	if existingUser != nil {
		return nil, ErrEmailAlreadyInUse
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil || user == nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return "", nil, ErrInvalidCredentials
	}

	token, err := s.issueToken(user.ID)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, username string) (*domain.User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, errors.New("username is required")
	}
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user.Username = username
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) RefreshToken(ctx context.Context, id uuid.UUID) (string, error) {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return "", err
	}
	return s.issueToken(id)
}

func (s *userService) ChangePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword, confirmPassword string) error {
	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		return errors.New("current password, new password, and confirmation are required")
	}
	if len(newPassword) < 8 {
		return errors.New("new password must contain at least 8 characters")
	}
	if newPassword != confirmPassword {
		return ErrPasswordConfirmation
	}
	if currentPassword == newPassword {
		return errors.New("new password must be different from the current password")
	}
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)) != nil {
		return ErrInvalidCurrentPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	return s.repo.Update(ctx, user)
}

func (s *userService) issueToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"jti": uuid.NewString(),
		"iat": now.Unix(),
		"exp": now.Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}
