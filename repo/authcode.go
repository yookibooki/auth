package repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/yookibooki/auth/auth"
)

type AuthCode struct {
	ID          int
	CodeHash    string
	UserID      int
	ClientID    string
	RedirectURI string
	State       string
	ExpiresAt   time.Time
	UsedAt      sql.NullTime
}

type AuthCodeRepo interface {
	Create(ctx context.Context, code *auth.AuthCode) error
	FindByCodeHash(ctx context.Context, codeHash string) (*AuthCode, error)
	MarkUsed(ctx context.Context, id int) error
	CleanupExpired(ctx context.Context) error
}

type authCodeRepo struct {
	db *sql.DB
}

func NewAuthCodeRepo(db *sql.DB) AuthCodeRepo {
	return &authCodeRepo{db: db}
}

func (r *authCodeRepo) Create(ctx context.Context, code *auth.AuthCode) error {
	query := `
		INSERT INTO auth_codes (code_hash, user_id, client_id, redirect_uri, state, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	var id int
	err := r.db.QueryRowContext(ctx, query,
		code.CodeHash,
		code.UserID,
		code.ClientID,
		code.RedirectURI,
		code.State,
		code.ExpiresAt,
	).Scan(&id)
	return err
}

func (r *authCodeRepo) FindByCodeHash(ctx context.Context, codeHash string) (*AuthCode, error) {
	query := `
		SELECT id, code_hash, user_id, client_id, redirect_uri, state, expires_at, used_at
		FROM auth_codes
		WHERE code_hash = $1
	`
	var code AuthCode
	err := r.db.QueryRowContext(ctx, query, codeHash).Scan(
		&code.ID,
		&code.CodeHash,
		&code.UserID,
		&code.ClientID,
		&code.RedirectURI,
		&code.State,
		&code.ExpiresAt,
		&code.UsedAt,
	)
	if err != nil {
		return nil, err
	}
	return &code, nil
}

func (r *authCodeRepo) MarkUsed(ctx context.Context, id int) error {
	query := `
		UPDATE auth_codes
		SET used_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *authCodeRepo) CleanupExpired(ctx context.Context) error {
	query := `
		DELETE FROM auth_codes
		WHERE expires_at < NOW()
	`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
