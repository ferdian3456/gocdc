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
	query := "INSERT INTO users (id,name,email,profile_picture,password,address,phonenumber,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)"
	_, err := tx.ExecContext(ctx, query, user.Id, user.Name, user.Email, user.Profile_picture, user.Password, user.Address, user.PhoneNumber, user.Created_at, user.Updated_at)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}
}

func (repository *UserRepository) AddRefreshTokenWithTx(ctx context.Context, tx *sql.Tx, refreshtoken domain.RefreshToken) {
	query := "INSERT INTO refresh_tokens (user_id,hashed_refresh_token,created_at,expired_at) VALUES ($1,$2,$3,$4)"
	_, err := tx.ExecContext(ctx, query, refreshtoken.User_id, refreshtoken.Hashed_refresh_token, refreshtoken.Created_at, refreshtoken.Expired_at)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}
}

func (repository *UserRepository) UpdateRefreshToken(ctx context.Context, tx *sql.Tx, tokenStatus string, userUUID string) {
	query := "UPDATE refresh_tokens SET status = $1 WHERE user_id = $2 AND created_at = (SELECT MAX(created_at) FROM refresh_tokens WHERE user_id = $2)"
	_, err := tx.ExecContext(ctx, query, tokenStatus, userUUID)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}
}

func (repository *UserRepository) FindLatestRefreshToken(ctx context.Context, tx *sql.Tx) (string, error) {
	query := "SELECT hashed_refresh_token FROM refresh_tokens ORDER BY created_at DESC LIMIT 1"
	row, err := tx.QueryContext(ctx, query)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer row.Close()

	var hashed_refresh_token string
	if row.Next() {
		err = row.Scan(&hashed_refresh_token)
		return hashed_refresh_token, nil
	} else {
		return "", errors.New("refresh token not found")
	}
}

func (repository *UserRepository) LoginWithTx(ctx context.Context, tx *sql.Tx, email string) (domain.User, error) {
	query := "SELECT id,email,password FROM users WHERE email=$1"
	row, err := tx.QueryContext(ctx, query, email)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer row.Close()

	user := domain.User{}
	if row.Next() {
		err = row.Scan(&user.Id, &user.Email, &user.Password)
		if err != nil {
			respErr := errors.New("failed to scan query result")
			repository.Log.Panic().Err(err).Msg(respErr.Error())
		}
		return user, nil
	} else {
		return user, errors.New("wrong email or password")
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

func (repository *UserRepository) FindUserNameAddress(ctx context.Context, userUUID string) (user.UserNameAddressResponse, error) {
	query := "SELECT name,address FROM users WHERE id=$1"
	row, err := repository.DB.QueryContext(ctx, query, userUUID)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer row.Close()

	user := user.UserNameAddressResponse{}

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

func (repository *UserRepository) FindUserEmail(ctx context.Context, userUUID string) (*string, error) {
	query := "SELECT email FROM users WHERE id=$1"
	row, err := repository.DB.QueryContext(ctx, query, userUUID)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer row.Close()

	var useremail *string
	if row.Next() {
		err = row.Scan(&useremail)
		if err != nil {
			respErr := errors.New("failed to scan query result")
			repository.Log.Panic().Err(err).Msg(respErr.Error())
		}

		return useremail, nil
	} else {
		return useremail, errors.New("user not found")
	}
}

func (repository *UserRepository) CheckUserExistence(ctx context.Context, userUUID string) (string, error) {
	query := "SELECT name FROM users WHERE id=$1"
	row, err := repository.DB.QueryContext(ctx, query, userUUID)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer row.Close()

	if row.Next() {
		return "User exist", nil
	} else {
		return "User not found", errors.New("user not found")
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
	query := "SELECT name,email FROM users WHERE name=$1 OR email=$2"
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

func (repository *UserRepository) FindUserEmailByUUID(ctx context.Context, tx *sql.Tx, userUUID string) (*string, error) {
	query := "SELECT email FROM users WHERE id=$1"
	row, err := tx.QueryContext(ctx, query, userUUID)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer row.Close()

	var email *string
	if row.Next() {
		err = row.Scan(&email)
		if err != nil {
			respErr := errors.New("failed to scan query result")
			repository.Log.Panic().Err(err).Msg(respErr.Error())
		}

		return email, nil
	} else {
		return email, errors.New("email not found")
	}
}
