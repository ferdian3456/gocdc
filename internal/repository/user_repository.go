package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"gocdc/internal/model/domain"
	"gocdc/internal/model/web/user"
)

type UserRepository struct {
	Log *zerolog.Logger
	DB  *sql.DB
}

func NewUserRepository(zerolog *zerolog.Logger, db *sql.DB) *UserRepository {
	return &UserRepository{
		Log: zerolog,
		DB:  db,
	}
}

func (repository *UserRepository) RegisterWithTx(ctx context.Context, tx *sql.Tx, user domain.User) {
	query := "INSERT INTO users (id,name,email,password,address,phonenumber,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)"
	_, err := tx.ExecContext(ctx, query, user.Id, user.Name, user.Email, user.Password, user.Address, user.PhoneNumber, user.Created_at, user.Updated_at)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}
}

func (repository *UserRepository) Login(ctx context.Context, email string) (domain.User, error) {
	query := "SELECT email,password FROM users WHERE email=$1"
	row, err := repository.DB.QueryContext(ctx, query, email)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer row.Close()

	user := domain.User{}
	if row.Next() {
		err = row.Scan(&user.Email, &user.Password)
		if err != nil {
			respErr := errors.New("failed to scan query result")
			repository.Log.Panic().Err(err).Msg(respErr.Error())
		}
		return user, nil
	} else {
		return user, errors.New("user not found")
	}
}

func (repository *UserRepository) Update(ctx context.Context, user domain.User) {
	query := "UPDATE users SET "
	args := []interface{}{}
	argCounter := 1

	if user.Name != "" {
		query += fmt.Sprintf("name = $%d, ", argCounter)
		args = append(args, user.Name)
		argCounter++
	}
	if user.Email != "" {
		query += fmt.Sprintf("email = $%d, ", argCounter)
		args = append(args, user.Email)
		argCounter++
	}
	if user.Password != "" {
		query += fmt.Sprintf("password = $%d, ", argCounter)
		args = append(args, user.Password)
		argCounter++
	}
	if user.Address != "" {
		query += fmt.Sprintf("address = $%d, ", argCounter)
		args = append(args, user.Address)
		argCounter++
	}
	if user.PhoneNumber != "" {
		query += fmt.Sprintf("phonenumber = $%d, ", argCounter)
		args = append(args, user.PhoneNumber)
		argCounter++
	}

	query += fmt.Sprintf("updated_at = $%d ", argCounter)
	args = append(args, user.Updated_at)
	argCounter++

	query += fmt.Sprintf("WHERE id = $%d", argCounter)
	args = append(args, user.Id)

	_, err := repository.DB.ExecContext(ctx, query, args...)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}
}

func (repository *UserRepository) Delete(ctx context.Context, userUUID string) {
	query := "DELETE FROM users WHERE id=$1"
	_, err := repository.DB.ExecContext(ctx, query, userUUID)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}
}

func (repository *UserRepository) FindUserInfo(ctx context.Context, userUUID string) (user.UserResponse, error) {
	query := "SELECT id,name,email,address,phonenumber,created_at,updated_at FROM users WHERE id=$1"
	row, err := repository.DB.QueryContext(ctx, query, userUUID)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer row.Close()

	user := user.UserResponse{}

	if row.Next() {
		err = row.Scan(&user.Id, &user.Name, &user.Email, &user.Address, &user.PhoneNumber, &user.Created_at, &user.Updated_at)
		if err != nil {
			respErr := errors.New("failed to scan query result")
			repository.Log.Panic().Err(err).Msg(respErr.Error())
		}

		return user, nil
	} else {
		return user, errors.New("user not found")
	}
}

func (repository *UserRepository) FindUserInfoWithTx(ctx context.Context, tx *sql.Tx, userUUID string) (user.UserResponse, error) {
	query := "SELECT name,address FROM users WHERE id=$1"
	row, err := tx.QueryContext(ctx, query, userUUID)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer row.Close()

	user := user.UserResponse{}

	if row.Next() {
		err = row.Scan(&user.Name, &user.Address)
		if err != nil {
			respErr := errors.New("failed to scan query result")
			repository.Log.Panic().Err(err).Msg(respErr.Error())
		}

		return user, nil
	} else {
		return user, errors.New("user not found")
	}
}

func (repository *UserRepository) CheckUserExistence(ctx context.Context, userUUID string) error {
	query := "SELECT name FROM users WHERE id=$1"
	row, err := repository.DB.QueryContext(ctx, query, userUUID)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer row.Close()

	if row.Next() {
		return nil
	} else {
		return errors.New("user not found")
	}
}

func (repository *UserRepository) CheckUserExistenceWithTx(ctx context.Context, tx *sql.Tx, userUUID string) error {
	query := "SELECT name FROM users WHERE id=$1"
	row, err := tx.QueryContext(ctx, query, userUUID)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer row.Close()

	if row.Next() {
		return nil
	} else {
		return errors.New("user not found")
	}
}

func (repository *UserRepository) CheckCredentialUniqueWithTx(ctx context.Context, tx *sql.Tx, user domain.User) error {
	query := "SELECT name,email FROM users WHERE name=$1 AND email=$2"
	row, err := tx.QueryContext(ctx, query, user.Name, user.Email)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer row.Close()

	if row.Next() {
		return errors.New("name or email are already exist")
	} else {
		return nil
	}
}
