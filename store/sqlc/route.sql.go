// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: route.sql

package sqlc

import (
	"context"
	"time"
)

const createRoute = `-- name: CreateRoute :one
INSERT INTO routes (
  distance, eta, polyline
) VALUES (
  $1, $2, ST_GeographyFromText($3)
)
RETURNING id, distance, polyline, eta, created_at, updated_at
`

type CreateRouteParams struct {
	Distance string      `json:"distance"`
	Eta      time.Time   `json:"eta"`
	Polyline interface{} `json:"polyline"`
}

func (q *Queries) CreateRoute(ctx context.Context, arg CreateRouteParams) (Route, error) {
	row := q.db.QueryRowContext(ctx, createRoute, arg.Distance, arg.Eta, arg.Polyline)
	var i Route
	err := row.Scan(
		&i.ID,
		&i.Distance,
		&i.Polyline,
		&i.Eta,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}