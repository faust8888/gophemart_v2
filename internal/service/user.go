// Package service содержит бизнес-логику накопительной системы лояльности.
package service

import (
	"context"
	"errors"
	"strings"

	"github.com/faust8888/gophemart_v2/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidInput возвращается, если логин или пароль не заданы.
var ErrInvalidInput = errors.New("invalid login or password")

// ErrInvalidCredentials возвращается, если пара логин/пароль неверна.
var ErrInvalidCredentials = errors.New("invalid login or password")

// dummyPasswordHash — валидный bcrypt-хеш, используется только для выравнивания
// времени ответа, когда пользователь с указанным логином не найден.
const dummyPasswordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

// UserRepository описывает операции сохранения и чтения пользователей.
type UserRepository interface {
	// Create сохраняет нового пользователя с уникальным логином.
	Create(ctx context.Context, login, passwordHash string) (*model.User, error)
	// GetByLogin возвращает пользователя по логину.
	GetByLogin(ctx context.Context, login string) (*model.User, error)
}

// TokenIssuer выпускает токен аутентификации для пользователя.
type TokenIssuer interface {
	// IssueToken выпускает подписанный токен для пользователя с идентификатором userID.
	IssueToken(userID, login string) (string, error)
}

// UserService реализует сценарии работы с пользователями.
type UserService struct {
	repo     UserRepository
	tokens   TokenIssuer
	hashCost int
}

// NewUserService создаёт сервис пользователей с bcrypt-хешированием паролей
// и выпуском токена после успешной регистрации или аутентификации.
func NewUserService(repo UserRepository, tokens TokenIssuer) *UserService {
	return &UserService{
		repo:     repo,
		tokens:   tokens,
		hashCost: bcrypt.DefaultCost,
	}
}

// Register регистрирует пользователя по паре логин/пароль и возвращает
// выпущенный токен аутентификации. Каждый логин должен быть уникальным.
func (s *UserService) Register(ctx context.Context, login, password string) (*model.User, string, error) {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return nil, "", ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.hashCost)
	if err != nil {
		return nil, "", err
	}

	user, err := s.repo.Create(ctx, login, string(hash))
	if err != nil {
		if errors.Is(err, model.ErrLoginTaken) {
			return nil, "", model.ErrLoginTaken
		}
		return nil, "", err
	}

	token, err := s.tokens.IssueToken(user.ID, user.Login)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

// Login проверяет пару логин/пароль и возвращает токен аутентификации.
func (s *UserService) Login(ctx context.Context, login, password string) (*model.User, string, error) {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return nil, "", ErrInvalidInput
	}

	user, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(password))
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.tokens.IssueToken(user.ID, user.Login)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}
