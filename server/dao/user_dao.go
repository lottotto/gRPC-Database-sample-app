package dao

import (
	"context"
	"math/rand"

	"github.com/jmoiron/sqlx"
	"github.com/lottotto/stdgrpc/model"
)

type UserDao struct {
	Conn *sqlx.DB
}

func (dao *UserDao) Update(ctx context.Context, name string) error {
	query := `INSERT INTO example (NAME, NUMBER) VALUES ($1, $2)`
	_, err := dao.Conn.ExecContext(ctx, query, name, rand.Intn(100))
	if err != nil {
		return err
	}
	return nil
}

func (dao *UserDao) FindByName(ctx context.Context, name string) ([]model.User, error) {
	query := `SELECT * FROM example where NAME=$1`

	rows, err := dao.Conn.QueryxContext(ctx, query, name)
	if err != nil {
		return nil, err
	}
	var users []model.User
	for rows.Next() {
		var user model.User
		err := rows.StructScan(&user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
