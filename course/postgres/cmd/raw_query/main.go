package main

import (
	"context"
	"database/sql"
	"github.com/brianvoe/gofakeit"
	"github.com/jackc/pgx/v4"
	"log"
	"time"
)

const (
	dbDSN = "host=localhost port=54321 dbname=note user=note-user password=note-password sslmode=disable"
)

func main() {
	ctx := context.Background()
	con, err := pgx.Connect(ctx, dbDSN)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer func(con *pgx.Conn, ctx context.Context) {
		err := con.Close(ctx)
		if err != nil {

		}
	}(con, ctx)
	res, err := con.Exec(ctx, "INSERT INTO note (title, content, id, status) VALUES ($1, $2, 1, false)", gofakeit.BeerBlg(), gofakeit.BeerHop())
	if err != nil {
		log.Fatal("failed to insert note", err)
	}

	log.Printf("inserted %d rows", res.RowsAffected())

	rows, err := con.Query(ctx, "SELECT id, title, content, created_at, updated_at, status, dead_line, author FROM note")
	if err != nil {
		log.Fatal("failed to select notes: ", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, author int64
		var title, content string
		var status bool
		var dead_line, updated_at sql.NullTime
		var created_at time.Time

		err = rows.Scan(&id, &title, &content, &created_at, &updated_at, &status, &dead_line, &author)
		if err != nil {
			log.Fatal("failed to scan note:", err)
		}
		log.Printf("id: %d", id)
	}
}
