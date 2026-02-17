package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/hero/microservice/user-service/internal/model"
	"github.com/hero/microservice/user-service/internal/rabbitmq"
	"github.com/hero/microservice/user-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	Register(input model.RegisterInput) (*model.User, error)
	Login(input model.LoginInput) (*model.User, error)
	GetProfile(id uuid.UUID) (*model.User, error)
	UpdateProfile(id uuid.UUID, input model.UpdateInput) (*model.User, error)
	DeleteUser(id uuid.UUID) error
}

type userService struct {
	repo      repository.UserRepository
	publisher *rabbitmq.Publisher
}

func NewUserService(repo repository.UserRepository, publisher *rabbitmq.Publisher) UserService {
	return &userService{repo: repo, publisher: publisher}
}

func (s *userService) Register(input model.RegisterInput) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &model.User{
		ID:           uuid.New(),
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(hash),
	}

	if err := s.repo.Create(user); err != nil {
		return nil, errors.New("failed to create user: " + err.Error())
	}

	// Publish user.registered event
	s.publisher.Publish("user.registered", map[string]interface{}{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
	})

	return user, nil
}

func (s *userService) Login(input model.LoginInput) (*model.User, error) {
	user, err := s.repo.GetByEmail(input.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}

func (s *userService) GetProfile(id uuid.UUID) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (s *userService) UpdateProfile(id uuid.UUID, input model.UpdateInput) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if input.Username != "" {
		user.Username = input.Username
	}
	if input.Email != "" {
		user.Email = input.Email
	}

	if err := s.repo.Update(user); err != nil {
		return nil, errors.New("failed to update user: " + err.Error())
	}

	// Publish user.updated event
	s.publisher.Publish("user.updated", map[string]interface{}{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
	})

	return user, nil
}

func (s *userService) DeleteUser(id uuid.UUID) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	if err := s.repo.Delete(id); err != nil {
		return errors.New("failed to delete user: " + err.Error())
	}

	// Publish user.deleted event
	s.publisher.Publish("user.deleted", map[string]interface{}{
		"user_id": id.String(),
	})

	return nil
}
