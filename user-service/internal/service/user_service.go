package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hero/microservice/user-service/internal/model"
	"github.com/hero/microservice/user-service/internal/rabbitmq"
	"github.com/hero/microservice/user-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const sessionTTL = 24 * time.Hour

type UserService interface {
	Register(input model.RegisterInput) (*model.User, error)
	Login(input model.LoginInput) (*model.LoginResponse, error)
	Logout(token string) error
	ValidateSession(token string) (*model.User, error)
	GetProfile(id uuid.UUID) (*model.User, error)
	UpdateProfile(id uuid.UUID, input model.UpdateInput) (*model.User, error)
	DeleteUser(id uuid.UUID) error
}

type userService struct {
	repo      repository.UserRepository
	publisher *rabbitmq.Publisher
	rdb       *redis.Client
}

func NewUserService(repo repository.UserRepository, publisher *rabbitmq.Publisher, rdb *redis.Client) UserService {
	return &userService{repo: repo, publisher: publisher, rdb: rdb}
}

func (s *userService) Register(input model.RegisterInput) (*model.User, error) {
	user := &model.User{
		ID:           uuid.New(),
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: input.Password,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, errors.New("failed to create user: " + err.Error())
	}

	s.publisher.Publish("user.registered", map[string]interface{}{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
	})

	return user, nil
}

func (s *userService) Login(input model.LoginInput) (*model.LoginResponse, error) {
	user, err := s.repo.GetByEmail(input.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if user.PasswordHash != input.Password {
		return nil, errors.New("invalid email or password")
	}

	// Create session token
	token := uuid.New().String()
	userJSON, _ := json.Marshal(user)
	s.rdb.Set(context.Background(), "session:"+token, userJSON, sessionTTL)

	return &model.LoginResponse{User: user, Token: token}, nil
}

func (s *userService) Logout(token string) error {
	result := s.rdb.Del(context.Background(), "session:"+token)
	if result.Err() != nil {
		return errors.New("failed to logout")
	}
	return nil
}

func (s *userService) ValidateSession(token string) (*model.User, error) {
	val, err := s.rdb.Get(context.Background(), "session:"+token).Result()
	if err == redis.Nil {
		return nil, errors.New("invalid or expired session")
	}
	if err != nil {
		return nil, errors.New("session validation failed")
	}

	var user model.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, errors.New("failed to parse session data")
	}

	return &user, nil
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

	s.publisher.Publish("user.deleted", map[string]interface{}{
		"user_id": id.String(),
	})

	return nil
}
