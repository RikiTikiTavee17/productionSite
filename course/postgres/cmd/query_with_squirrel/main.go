package main

import (
	"context"
	"database/sql"
	"github.com/Masterminds/squirrel"
	"github.com/brianvoe/gofakeit"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"time"
)

const (
	dbDSN = "host=localhost port=54321 dbname=note user=note-user password=note-password sslmode=disable"
)

func main() {
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer pool.Close()

	builderInsert := squirrel.Insert("note").
		PlaceholderFormat(squirrel.Dollar).
		Columns("title", "content", "id").
		Values(gofakeit.City(), gofakeit.City(), gofakeit.Int64()).
		Suffix("RETURNING id")

	query, args, err := builderInsert.ToSql()
	if err != nil {
		log.Fatal("failed to build query", err)
	}

	var noteID int
	err = pool.QueryRow(ctx, query, args...).Scan(&noteID)
	if err != nil {
		log.Fatal("failed to insert note:", err)
	}

	log.Printf("inserted note with id: %d", noteID)

	builderSelect := squirrel.Select("id", "title", "content", "created_at", "updated_at").
		From("note").
		PlaceholderFormat(squirrel.Dollar).
		OrderBy("id ASC").
		Limit(10)

	query, args, err = builderSelect.ToSql()
	if err != nil {
		log.Fatalf("failed to build query: %v", err)
	}

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		log.Fatalf("failed to select notes: %v", err)
	}

	var id int
	var title, content string
	var createdAt time.Time
	var updatedAt sql.NullTime

	for rows.Next() {
		err = rows.Scan(&id, &title, &content, &createdAt, &updatedAt)
		if err != nil {
			log.Fatalf("failed to scan note: %v", err)
		}

		log.Printf("id: %d, title: %s, content: %s, created_at: %v, updated_at: %v\n", id, title, content, createdAt, updatedAt)
	}

	// Делаем запрос на обновление записи в таблице note
	builderUpdate := squirrel.Update("note").
		PlaceholderFormat(squirrel.Dollar).
		Set("title", gofakeit.City()).
		Set("content", gofakeit.Address().Street).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": noteID})

	query, args, err = builderUpdate.ToSql()
	if err != nil {
		log.Fatalf("failed to build query: %v", err)
	}

	res, err := pool.Exec(ctx, query, args...)
	if err != nil {
		log.Fatalf("failed to update note: %v", err)
	}

	log.Printf("updated %d rows", res.RowsAffected())

	// Делаем запрос на получение измененной записи из таблицы note
	builderSelectOne := squirrel.Select("id", "title", "content", "created_at", "updated_at").
		From("note").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": noteID}).
		Limit(1)

	query, args, err = builderSelectOne.ToSql()
	if err != nil {
		log.Fatalf("failed to build query: %v", err)
	}

	err = pool.QueryRow(ctx, query, args...).Scan(&id, &title, &content, &createdAt, &updatedAt)
	if err != nil {
		log.Fatalf("failed to select notes: %v", err)
	}

	log.Printf("id: %d, title: %s, content: %s, created_at: %v, updated_at: %v\n", id, title, content, createdAt, updatedAt)
}
