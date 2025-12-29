package repo

import (
	"context"
	"database/sql"
)

type User struct {
	ID      int
	Email   string
	PwdHash string
}

type UserRepo interface {
	Create(ctx context.Context, email, pwdHash string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id int) (*User, error)
	UpdateEmail(ctx context.Context, id int, email string) error
	UpdatePassword(ctx context.Context, id int, pwdHash string) error
	Delete(ctx context.Context, id int) error
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, email, pwdHash string) (*User, error) {
	query := `
		INSERT INTO users (email, pwd_hash)
		VALUES ($1, $2)
		RETURNING id, email, pwd_hash
	`
	var user User
	err := r.db.QueryRowContext(ctx, query, email, pwdHash).Scan(
		&user.ID,
		&user.Email,
		&user.PwdHash,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, pwd_hash
		FROM users
		WHERE email = $1
	`
	var user User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PwdHash,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) FindByID(ctx context.Context, id int) (*User, error) {
	query := `
		SELECT id, email, pwd_hash
		FROM users
		WHERE id = $1
	`
	var user User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PwdHash,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) UpdateEmail(ctx context.Context, id int, email string) error {
	query := `
		UPDATE users
		SET email = $1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, email, id)
	return err
}

func (r *userRepo) UpdatePassword(ctx context.Context, id int, pwdHash string) error {
	query := `
		UPDATE users
		SET pwd_hash = $1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, pwdHash, id)
	return err
}

func (r *userRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
