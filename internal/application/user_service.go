package application

import (
	"context"
	"errors"

	"backend/internal/domain"
	// import "golang.org/x/crypto/bcrypt" // Uncomment to use bcrypt for password hashing
)

type UserService interface {
	CreateUser(ctx context.Context, username, email, password string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uint) (*domain.User, error)
}

type userService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, username, email, password string) (*domain.User, error) {
	if username == "" || email == "" || password == "" {
		return nil, errors.New("username, email, and password are required")
	}

	// Check if user already exists
	existingUser, _ := s.repo.GetByEmail(ctx, email)
	if existingUser != nil {
		return nil, errors.New("email already in use")
	}

	// Hash password (Example implementation)
	// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	// if err != nil { return nil, err }

	user := &domain.User{
		Username: username,
		Email:    email,
		Password: password, // Replace with string(hashedPassword)
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, id uint) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}
