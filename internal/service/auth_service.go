package service

import (
    "context"
    "errors"

    "github.com/go-chi/jwtauth"
    "golang.org/x/crypto/bcrypt"
    "github.com/GuilhAndrad/point-api/internal/model"
    "github.com/GuilhAndrad/point-api/internal/repository"
)

var (
    ErrInvalidCredentials = errors.New("email ou senha inválidos")
    ErrEmailAlreadyExists = errors.New("email já cadastrado")
)

type AuthService interface {
    Register(ctx context.Context, name, email, password string, role model.Role) (*model.User, error)
    Login(ctx context.Context, email, password string) (string, *model.User, error)
}

type authService struct {
    repo      repository.UserRepository
    tokenAuth *jwtauth.JWTAuth
}

func NewAuthService(repo repository.UserRepository, secret string) AuthService {
    return &authService{
        repo:      repo,
        tokenAuth: jwtauth.New("HS256", []byte(secret), nil),
    }
}

func (s *authService) Register(ctx context.Context, name, email, password string, role model.Role) (*model.User, error) {
    // Verifica se e-mail já existe
    if _, err := s.repo.FindByEmail(ctx, email); err == nil {
        return nil, ErrEmailAlreadyExists
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    user := &model.User{
        Name:         name,
        Email:        email,
        PasswordHash: string(hash),
        Role:         role,
    }

    if err := s.repo.Create(ctx, user); err != nil {
        return nil, err
    }
    return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, *model.User, error) {
    user, err := s.repo.FindByEmail(ctx, email)
    if err != nil {
        return "", nil, ErrInvalidCredentials
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
        return "", nil, ErrInvalidCredentials
    }

    // Gera JWT com id e role do usuário
    _, token, _ := s.tokenAuth.Encode(map[string]interface{}{
        "user_id": user.ID,
        "role":    user.Role,
    })

    return token, user, nil
}