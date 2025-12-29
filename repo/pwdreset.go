package repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/yookibooki/auth/auth"
)

type PwdResetToken struct {
	ID        int
	TokenHash string
	UserID    int
	ExpiresAt time.Time
	UsedAt    sql.NullTime
}

type PwdResetTokenRepo interface {
	Create(ctx context.Context, token *auth.PwdResetToken) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*PwdResetToken, error)
	MarkUsed(ctx context.Context, id int) error
	CleanupExpired(ctx context.Context) error
}

type pwdResetTokenRepo struct {
	db *sql.DB
}

func NewPwdResetTokenRepo(db *sql.DB) PwdResetTokenRepo {
	return &pwdResetTokenRepo{db: db}
}

func (r *pwdResetTokenRepo) Create(ctx context.Context, token *auth.PwdResetToken) error {
	query := `
		INSERT INTO pwd_reset_tokens (token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var id int
	err := r.db.QueryRowContext(ctx, query,
		token.TokenHash,
		token.UserID,
		token.ExpiresAt,
	).Scan(&id)
	return err
}

func (r *pwdResetTokenRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*PwdResetToken, error) {
	query := `
		SELECT id, token_hash, user_id, expires_at, used_at
		FROM pwd_reset_tokens
		WHERE token_hash = $1
	`
	var token PwdResetToken
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.TokenHash,
		&token.UserID,
		&token.ExpiresAt,
		&token.UsedAt,
	)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *pwdResetTokenRepo) MarkUsed(ctx context.Context, id int) error {
	query := `
		UPDATE pwd_reset_tokens
		SET used_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *pwdResetTokenRepo) CleanupExpired(ctx context.Context) error {
	query := `
		DELETE FROM pwd_reset_tokens
		WHERE expires_at < NOW()
	`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
