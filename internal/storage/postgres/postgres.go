package postgres

import (
	"database/sql"
	"denet/internal/lib/models"
	"denet/internal/storage"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgresql.New"
	db, err := sql.Open("postgres", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", op, err)
	}

	stmt, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(100) NOT NULL UNIQUE,
			points INT DEFAULT 0,
			referral_id INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_referral_id ON users(referral_id);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) SaveUser(username, password string, points, referral_id int64) (int64, error) {
	const op = "storage.postgresql.SaveUser"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to hash password: %w", op, err)
	}

	stmt, err := s.db.Prepare("INSERT INTO users (username, password, points, referral_id) VALUES($1, $2, $3, $4) RETURNING id")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var id int64
	err = stmt.QueryRow(username, string(hashedPassword), points, referral_id).Scan(&id)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return 0, storage.ErrUserExists
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (s *Storage) LoginUser(username string, password string) (*models.User, error) {
	const op = "storage.mysql.LoginUser"
	stmt, err := s.db.Prepare("SELECT username, password FROM users WHERE username = $1")
	if err != nil {
		return nil, fmt.Errorf("%s: prepare statement %w", op, err)
	}

	user := &models.User{}
	var hashedPassword string
	err = stmt.QueryRow(username).Scan(&user.Username, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	// Сравнение пароля с хешом из базы
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return nil, fmt.Errorf("%s: invalid password: %w", op, err)
	}

	return user, nil
}

func (s *Storage) GetUSER(id int64) (*models.User, error) {
	const op = "storage.mysql.GetUSER"
	fmt.Println(id)
	stmt, err := s.db.Prepare("select id, username, password, points, referral_id, created_at from users where id = $1")
	if err != nil {
		return nil, fmt.Errorf("%s: prepare statement %w", op, err)
	}

	user := &models.User{}
	err = stmt.QueryRow(id).Scan(&user.Id, &user.Username, &user.Password, &user.Points, &user.Referral_id, &user.Created_at)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	return user, nil
}

func (s *Storage) GetLeaderboard() ([]models.User, error) {
	const op = "storage.mysql.GetLeaderboard"
	query := `SELECT id, username, points, referral_id, created_at FROM users ORDER BY points DESC LIMIT 5`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.Id, &user.Username, &user.Points, &user.Referral_id, &user.Created_at); err != nil {
			return nil, fmt.Errorf("%s: execute statement: %w", op, err)
		}
		users = append(users, user)
	}
	return users, nil
}

func (s *Storage) CompleteTask(userID int64, taskPoints int64) error {
	result, err := s.db.Exec(`UPDATE users SET points = points + $1 WHERE id = $2`, taskPoints, userID)
	if err != nil {
		return fmt.Errorf("failed to update points for user %d: %w", userID, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows count: %w", err)
	}
	if rowsAffected == 0 {
		return storage.ErrUserNotFound
	}
	return nil
}

func (s *Storage) SetReferral(userID int64, referralID int64) error {
	result, err := s.db.Exec(`UPDATE users SET referral_id = $1, points=points+5 WHERE id = $2`, referralID, userID)
	if err != nil {
		return fmt.Errorf("failed to update points for user %d: %w", userID, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows count: %w", err)
	}
	if rowsAffected == 0 {
		return storage.ErrUserNotFound
	}
	return nil
}
