package models

import (
	"context"
	"database/sql"

	"github.com/redis/go-redis/v9"
)

func InitModel(db *sql.DB, rd *redis.Client) *BookModel {
	ctx, cancel := context.WithCancel(context.Background())
	return &BookModel{db: db, redis: rd, ctx: ctx, cancel: cancel}
}
