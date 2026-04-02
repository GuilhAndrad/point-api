package repository

import (
	"context"

	"github.com/GuilhAndrad/point-api/internal/model"
	"github.com/jmoiron/sqlx"
)

// Interface — o Service só conhece isso, não o banco
type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    FindByEmail(ctx context.Context, email string) (*model.User, error)
    FindByID(ctx context.Context, id int64) (*model.User, error)
    FindAll(ctx context.Context) ([]model.User, error)
}

type userRepository struct {
    db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
    query := `
        INSERT INTO users (name, email, password_hash, role)
        VALUES (:name, :email, :password_hash, :role)
        RETURNING id, created_at
    `
    rows, err := r.db.NamedQueryContext(ctx, query, user)
    if err != nil {
        return err
    }
    defer rows.Close()
    if rows.Next() {
        return rows.Scan(&user.ID, &user.CreatedAt)
    }
    return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
    var user model.User
    err := r.db.GetContext(ctx, &user,
        "SELECT * FROM users WHERE email = $1", email)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
    var user model.User
    err := r.db.GetContext(ctx, &user,
        "SELECT * FROM users WHERE id = $1", id)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) FindAll(ctx context.Context) ([]model.User, error) {
    var users []model.User
    err := r.db.SelectContext(ctx, &users,
        "SELECT id, name, email, role, created_at FROM users ORDER BY name")
    return users, err
}