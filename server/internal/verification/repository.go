package verification

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateVerificationCode(code *VerificationCode) error {
	query := `
		INSERT INTO email_verification_codes (user_id, code, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(query, code.UserID, code.Code, code.ExpiresAt).Scan(&code.ID, &code.CreatedAt)
	return err
}

func (r *Repository) GetActiveVerificationCode(userID int, code string) (*VerificationCode, error) {
	var vc VerificationCode
	query := `
		SELECT id, user_id, code, created_at, expires_at, used
		FROM email_verification_codes
		WHERE user_id = $1 AND code = $2 AND used = false
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := r.db.Get(&vc, query, userID, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCodeNotFound
		}
		return nil, err
	}
	return &vc, nil
}

func (r *Repository) MarkCodeAsUsed(codeID int) error {
	query := `
		UPDATE email_verification_codes
		SET used = true
		WHERE id = $1
	`
	_, err := r.db.Exec(query, codeID)
	return err
}

func (r *Repository) DeleteUserCodes(userID int) error {
	query := `
		DELETE FROM email_verification_codes
		WHERE user_id = $1
	`
	_, err := r.db.Exec(query, userID)
	return err
}
