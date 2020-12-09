package store

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v4"
	"math/rand"
	"time"
)

type User struct {
	Id        *int   `json:"user_id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Salt      string `json:"salt"`
	Role      string `json:"role"`
	RelatedId *int   `json:"related_id"`
}

func (s *Store) CreateUser(ctx context.Context, username string, password string, role string) error {
	var user User
	user.Username = username
	user.Salt = generateSalt(7)
	user.Password = generateChecksum(password, user.Salt)
	user.Role = role

	sql, _, err := goqu.Insert("users").
		Rows(goqu.Record{
			"username": user.Username,
			"password": user.Password,
			"salt":     user.Salt,
			"role":     user.Role,
		}).ToSQL()
	if err != nil {
		return fmt.Errorf("CreateRecord(): sql query build failed: %v", err)
	}
	if _, err := s.connPool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("CreateRecord(): execute a query failed: %v", err)
	}
	return nil
}

func (s *Store) AddRelatedIdToUser(ctx context.Context, username string, relatedId int) error {
	sql, _, err := goqu.Update("users").
		Set(goqu.Record{
			"id_related": relatedId,
		}).
		Where(goqu.C("username").Eq(username)).
		ToSQL()
	if err != nil {
		return fmt.Errorf("AddRelatedIdToUser(): sql query build failed: %v", err)
	}
	if _, err := s.connPool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("AddRelatedIdToUser(): execute a query failed: %v", err)
	}
	return nil
}

func (s *Store) IsRelatedIdSet(ctx context.Context, relatedId int) (*bool, error) {
	sql, _, err := goqu.Select("id", "username", "password", "salt", "role", "id_related").
		From("users").
		Where(goqu.C("id_related").Eq(relatedId)).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("IsRelatedIdSet(): sql query build failed: %v", err)
	}

	rows, err := s.connPool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("IsRelatedIdSet(): execute a query failed: %v", err)
	}
	defer rows.Close()

	var users []*User

	for rows.Next() {
		user, err := readUser(rows)
		if err != nil {
			return nil, fmt.Errorf("IsRelatedIdSet(): converting failed: %v", err)
		}
		users = append(users, user)
	}

	var cond = false
	if len(users) != 0 {
		cond = true
		return &cond, nil
	}
	return &cond, nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	sql, _, err := goqu.Select("id", "username", "password", "salt", "role", "id_related").
		From("users").
		Where(goqu.C("username").Eq(username)).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("GetUserByUsername(): sql query build failed: %v", err)
	}

	rows, err := s.connPool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("GetUserByUsername(): execute a query failed: %v", err)
	}
	defer rows.Close()

	var users []*User

	for rows.Next() {
		user, err := readUser(rows)
		if err != nil {
			return nil, fmt.Errorf("GetUserByUsername(): converting failed: %v", err)
		}
		users = append(users, user)
	}

	if len(users) == 0 {
		return nil, nil
	}
	return users[0], nil
}

func (s *Store) IsPasswordCorrect(ctx context.Context, username string, password string) (*bool, error) {
	user, err := s.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("IsPasswordCorrect(): %v", err)
	}

	var cond = false

	if generateChecksum(password, user.Salt) == user.Password {
		cond = true
		return &cond, nil
	}
	return &cond, nil
}

func generateSalt(n int) string {
	rand.Seed(time.Now().UnixNano())
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func generateChecksum(password string, salt string) string {
	var passwordChecksum = md5.Sum([]byte(password))
	var saltChecksum = md5.Sum([]byte(salt))
	var strChecksums = hex.EncodeToString(passwordChecksum[:]) + hex.EncodeToString(saltChecksum[:])
	var fullChecksum = md5.Sum([]byte(strChecksums))
	return hex.EncodeToString(fullChecksum[:])
}

func readUser(row pgx.Row) (*User, error) {
	var u User

	err := row.Scan(&u.Id, &u.Username, &u.Password, &u.Salt, &u.Role, &u.RelatedId)
	if err != nil {
		return nil, err
	}

	return &u, nil
}
