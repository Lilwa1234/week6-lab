package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Menu struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Type  string  `json:"type"`
	Stock int     `json:"stock"`
}

var ErrNotFound = errors.New("menu not found")

type MenuStore interface {
	List(ctx context.Context, menuType string) ([]Menu, error)
	Get(ctx context.Context, id int) (Menu, error)
	Add(ctx context.Context, m Menu) (Menu, error)
	Delete(ctx context.Context, id int) error
}

type PostgresStore struct {
	Pool *pgxpool.Pool
}

func (s *PostgresStore) List(ctx context.Context, menuType string) ([]Menu, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id, name, price, type, stock FROM menus
		 WHERE ($1 = '' OR type = $1)
		 ORDER BY id`, menuType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	menus := []Menu{}
	for rows.Next() {
		var m Menu
		if err := rows.Scan(&m.ID, &m.Name, &m.Price, &m.Type, &m.Stock); err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}

	return menus, rows.Err()
}

func (s *PostgresStore) Get(ctx context.Context, id int) (Menu, error) {
	var m Menu
	err := s.Pool.QueryRow(ctx,
		`SELECT id, name, price, type, stock FROM menus WHERE id = $1`, id).
		Scan(&m.ID, &m.Name, &m.Price, &m.Type, &m.Stock)

	if errors.Is(err, pgx.ErrNoRows) {
		return Menu{}, ErrNotFound
	}

	return m, err
}

func (s *PostgresStore) Add(ctx context.Context, m Menu) (Menu, error) {
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO menus (name, price, type, stock)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`, m.Name, m.Price, m.Type, m.Stock).Scan(&m.ID)

	return m, err
}

func (s *PostgresStore) Delete(ctx context.Context, id int) error {
	tag, err := s.Pool.Exec(ctx, `DELETE FROM menus WHERE id = $1`, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}