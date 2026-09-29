package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/users-api-go/internal/models"
)

var ErrNotFound = errors.New("user not found")
var ErrDuplicateEmail = errors.New("email already exists")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(name, email, passwordHash string) (*models.User, error) {
	now := time.Now().UTC()
	result, err := r.db.Exec(
		`INSERT INTO users (name, email, password, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		name, email, passwordHash, now, now,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return nil, ErrDuplicateEmail
		}
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:        id,
		Name:      name,
		Email:     email,
		Password:  passwordHash,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (r *UserRepository) GetByID(id int64) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(
		`SELECT id, name, email, password, created_at, updated_at FROM users WHERE id = ?`,
		id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) List(page, limit int) ([]models.User, int64, error) {
	var total int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	rows, err := r.db.Query(
		`SELECT id, name, email, password, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) Update(id int64, name, email *string, passwordHash *string) (*models.User, error) {
	user, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		user.Name = *name
	}
	if email != nil {
		user.Email = *email
	}
	if passwordHash != nil {
		user.Password = *passwordHash
	}
	user.UpdatedAt = time.Now().UTC()

	_, err = r.db.Exec(
		`UPDATE users SET name = ?, email = ?, password = ?, updated_at = ? WHERE id = ?`,
		user.Name, user.Email, user.Password, user.UpdatedAt, id,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return nil, ErrDuplicateEmail
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) Delete(id int64) error {
	result, err := r.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func isUniqueConstraintError(err error) bool {
	return err != nil && (
		contains(err.Error(), "UNIQUE constraint failed") ||
		contains(err.Error(), "constraint failed: UNIQUE"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
