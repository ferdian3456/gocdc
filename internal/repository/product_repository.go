package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"gocdc/internal/model/domain"
	"gocdc/internal/model/web/product"
)

type ProductRepository struct {
	Log *zerolog.Logger
	DB  *sql.DB
}

func NewProductRepository(zerolog *zerolog.Logger, db *sql.DB) *ProductRepository {
	return &ProductRepository{
		Log: zerolog,
		DB:  db,
	}
}

func (repository *ProductRepository) CreateWithTx(ctx context.Context, tx *sql.Tx, product domain.Product) {
	query := "INSERT INTO products (seller_id,name,quantity,price,weight,size,description,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)"
	_, err := tx.ExecContext(ctx, query, product.Seller_id, product.Name, product.Quantity, product.Price, product.Weight, product.Size, product.Description, product.Created_at, product.Updated_at)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}
}

func (repository *ProductRepository) UpdateWithTx(ctx context.Context, tx *sql.Tx, product domain.Product) {
	query := "UPDATE products SET "
	args := []interface{}{}
	argCounter := 1

	if product.Name != "" {
		query += fmt.Sprintf("name = $%d, ", argCounter)
		args = append(args, product.Name)
		argCounter++
	}
	if product.Quantity != 0 {
		query += fmt.Sprintf("quantity = $%d, ", argCounter)
		args = append(args, product.Quantity)
		argCounter++
	}
	if product.Price != 0 {
		query += fmt.Sprintf("price = $%d, ", argCounter)
		args = append(args, product.Price)
		argCounter++
	}
	if product.Weight != 0 {
		query += fmt.Sprintf("weight = $%d, ", argCounter)
		args = append(args, product.Weight)
		argCounter++
	}
	if product.Size != "" {
		query += fmt.Sprintf("size = $%d, ", argCounter)
		args = append(args, product.Size)
		argCounter++
	}
	if product.Description != "" {
		query += fmt.Sprintf("description = $%d, ", argCounter)
		args = append(args, product.Description)
		argCounter++
	}

	query += fmt.Sprintf("updated_at = $%d ", argCounter)
	args = append(args, product.Updated_at)
	argCounter++

	query += fmt.Sprintf("WHERE id = $%d", argCounter)
	args = append(args, product.Id)

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}
}

func (repository *ProductRepository) CheckOwnershipWithTx(ctx context.Context, tx *sql.Tx, userUUID string, productID int) error {
	query := "SELECT id FROM products WHERE id=$1 AND seller_id=$2"
	row, err := tx.QueryContext(ctx, query, productID, userUUID)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer row.Close()

	if row.Next() {
		return nil
	} else {
		return errors.New("product not found")
	}
}

func (repository *ProductRepository) CheckOwnership(ctx context.Context, userUUID string, productID int) error {
	query := "SELECT id FROM products WHERE id=$1 AND seller_id=$2"
	row, err := repository.DB.QueryContext(ctx, query, productID, userUUID)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer row.Close()

	if row.Next() {
		return nil
	} else {
		return errors.New("product not found")
	}
}

func (repository *ProductRepository) Delete(ctx context.Context, productID int) {
	query := "DELETE FROM products WHERE id=$1"
	_, err := repository.DB.ExecContext(ctx, query, productID)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}
}

func (repository *ProductRepository) FindProductInfo(ctx context.Context, productID int) (product.ProductResponse, error) {
	query := "SELECT id,seller_id,name,quantity,price,weight,size,status,description,created_at,updated_at FROM products WHERE id=$1"
	row, err := repository.DB.QueryContext(ctx, query, productID)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}

	defer row.Close()

	product := product.ProductResponse{}

	if row.Next() {
		err = row.Scan(&product.Id, &product.Seller_id, &product.Name, &product.Quantity, &product.Price, &product.Weight, &product.Size, &product.Status, &product.Description, &product.Created_at, &product.Updated_at)
		if err != nil {
			respErr := errors.New("failed to scan query result")
			repository.Log.Panic().Err(err).Msg(respErr.Error())
		}

		return product, nil
	} else {
		return product, errors.New("product not found")
	}
}

func (repository *ProductRepository) FindAllProduct(ctx context.Context) ([]product.ProductResponse, error) {
	query := "SELECT id,seller_id,name,quantity,price,weight,size,status,description,created_at,updated_at FROM products"
	row, err := repository.DB.QueryContext(ctx, query)
	if err != nil {
		respErr := errors.New("failed to query into database")
		repository.Log.Panic().Err(err).Msg(respErr.Error())
	}

	hasData := false

	defer row.Close()

	products := []product.ProductResponse{}

	for row.Next() {
		product := product.ProductResponse{}
		err = row.Scan(&product.Id, &product.Seller_id, &product.Name, &product.Quantity, &product.Price, &product.Weight, &product.Size, &product.Status, &product.Description, &product.Created_at, &product.Updated_at)
		if err != nil {
			respErr := errors.New("failed to scan query result")
			repository.Log.Panic().Err(err).Msg(respErr.Error())
		}

		products = append(products, product)

		hasData = true
	}

	if hasData == false {
		return products, errors.New("product not found")
	}

	return products, nil
}
