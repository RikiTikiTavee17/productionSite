package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	desc "github.com/RikiTikiTavee17/course/grpc/pkg/note_v1"
	"github.com/brianvoe/gofakeit"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net"
	"time"
)

const (
	dbDSN    = "host=course-db-1 port=5432 dbname=note user=note-user password=note-password sslmode=disable"
	grpcPort = 50051
)

/*
	type SyncMap struct {
		elems map[int64]*desc.Note
		num   int64
	}

	type LoginMap struct {
		checks  map[string]string
		idLogin map[int64]string
		loginId map[string]int64
		num     int64
	}

var notes = &SyncMap{elems: make(map[int64]*desc.Note), num: 1}
var users = &LoginMap{checks: make(map[string]string), idLogin: make(map[int64]string), loginId: make(map[string]int64), num: 1}
*/
type server struct {
	desc.UnimplementedNoteV1Server
}

func (s *server) Get(ctx context.Context, req *desc.GetRequest) (*desc.GetResponse, error) {
	log.Printf("Note id: %d", req.GetId())

	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		log.Printf("failed to connect to database: %d", err)
		return nil, errors.New("failed to connect to database")

	}
	defer pool.Close()

	builderSelectOne := squirrel.Select("id", "title", "content", "author", "dead_line", "status", "created_at", "updated_at").
		From("note").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": req.GetId()}).
		Limit(1)

	query, args, err := builderSelectOne.ToSql()
	if err != nil {
		log.Printf("failed to build query: %d", err)
		return nil, errors.New("failed to build query")
	}

	var id, author int64
	var status bool
	var title, content string
	var createdAt time.Time
	var updatedAt, dead_line time.Time

	err = pool.QueryRow(ctx, query, args...).Scan(&id, &title, &content, &author, &dead_line, &status, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("no note with such id in system: %d", err)
			return nil, errors.New("no note with such id in system")
		} else {
			log.Printf("failed to select notes: %d", err)
			return nil, errors.New("failed to select notes")
		}
	}
	n := &desc.Note{
		Id: id,
		Info: &desc.NoteInfo{
			Title:    title,
			Content:  content,
			Author:   author,
			DeadLine: timestamppb.New(dead_line),
			Status:   false,
		},
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
	}
	return &desc.GetResponse{
		Note: n,
	}, nil
}

func (s *server) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	var id int64
	a := req.GetInfo().GetAuthor()
	var flag bool = true

	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		log.Printf("failed to connect to database: %d", err)
		return nil, errors.New("failed to connect to database")
	}
	defer pool.Close()

	for flag {
		id = int64(gofakeit.Uint32())
		builderSelect := squirrel.Select("id").
			From("note").
			PlaceholderFormat(squirrel.Dollar).
			Where(squirrel.Eq{"id": id})

		query, args, err := builderSelect.ToSql()
		if err != nil {
			log.Printf("failed to build query: %d", err)
			return nil, errors.New("failed to build query")
		}

		err = pool.QueryRow(ctx, query, args...).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			flag = false
		}
	}

	builderSelect := squirrel.Select("id").
		From("persons").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": a})

	query, args, err := builderSelect.ToSql()
	if err != nil {
		log.Printf("failed to build query: %d", err)
		return nil, errors.New("failed to build query")
	}

	err = pool.QueryRow(ctx, query, args...).Scan(&a)
	if err != nil {
		log.Printf("error to use this user id: %d", err)
		return nil, errors.New("error to use this user id")
	}
	builderInsert := squirrel.Insert("note").
		PlaceholderFormat(squirrel.Dollar).
		Columns("id", "title", "content", "author", "dead_line", "status", "created_at", "updated_at").
		Values(id, req.GetInfo().GetTitle(), req.GetInfo().GetContent(), a, req.GetInfo().GetDeadLine().AsTime().Format("2006-01-02 15:04:05.999999"), req.GetInfo().GetStatus(), time.Now().Format("2006-01-02 15:04:05.999999"), time.Now().Format("2006-01-02 15:04:05.999999"))

	query, args, err = builderInsert.ToSql()
	if err != nil {
		log.Printf("failed to build query: %d", err)
		return nil, errors.New("failed to build query")
	}

	var noteID int
	err = pool.QueryRow(ctx, query, args...).Scan(&noteID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Printf("failed to insert note: %d", err)
		return nil, errors.New("failed to insert note")
	}

	return &desc.CreateResponse{Id: id}, nil
}

func (s *server) Update(ctx context.Context, req *desc.UpdateRequest) (*emptypb.Empty, error) {
	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		log.Printf("failed to connect to database: %d", err)
		return nil, errors.New("failed to connect to database")
	}
	defer pool.Close()

	reqId := req.GetId()

	builderSelect := squirrel.Select("id", "title", "content", "author", "dead_line", "status").
		From("note").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": reqId})

	query, args, err := builderSelect.ToSql()
	if err != nil {
		log.Printf("failed to build query: %d", err)
		return nil, errors.New("failed to build query")
	}
	var id, author int64
	var status bool
	var title, content string
	var dead_line time.Time

	err = pool.QueryRow(ctx, query, args...).Scan(&id, &title, &content, &author, &dead_line, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Printf("there is no note with such id in system: %d", err)
		return nil, errors.New("there is no note with such id in system")
	}

	if req.GetInfo().GetAuthor() != nil {
		auth := req.GetInfo().GetAuthor().GetValue()

		builderSelect = squirrel.Select("id").
			From("persons").
			PlaceholderFormat(squirrel.Dollar).
			Where(squirrel.Eq{"id": auth})

		query, args, err = builderSelect.ToSql()
		if err != nil {
			log.Printf("failed to build query: %d", err)
			return nil, errors.New("failed to build query")
		}
		err = pool.QueryRow(ctx, query, args...).Scan(&auth)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				log.Printf("there is no person with such id in system: %d", err)
				return nil, errors.New("there is no person with such id in system")
			} else {
				log.Printf("failed to find user with this id: %d", err)
				return nil, errors.New("failed to find user with this id")
			}
		}
	}
	var newAuthor int64
	if req.GetInfo().GetAuthor() == nil {
		newAuthor = author
	} else {
		newAuthor = req.GetInfo().GetAuthor().GetValue()
	}
	var newTitle string
	if req.GetInfo().GetTitle() == nil {
		newTitle = title
	} else {
		newTitle = req.GetInfo().GetTitle().GetValue()
	}
	var newContent string
	if req.GetInfo().GetContent() == nil {
		newContent = content
	} else {
		newContent = req.GetInfo().GetContent().GetValue()
	}
	var newStatus bool
	if req.GetInfo().GetStatus() == nil {
		newStatus = status
	} else {
		newStatus = req.GetInfo().GetStatus().GetValue()
	}
	var newDeadLine string
	if req.GetInfo().GetDeadLine() == nil {
		newDeadLine = dead_line.Format("2006-01-02 15:04:05.999999")
	} else {
		newDeadLine = req.GetInfo().GetDeadLine().AsTime().Format("2006-01-02 15:04:05.999999")
	}
	builderUpdate := squirrel.Update("note").
		PlaceholderFormat(squirrel.Dollar).
		Set("author", newAuthor).
		Set("title", newTitle).
		Set("content", newContent).
		Set("status", newStatus).
		Set("dead_line", newDeadLine).
		Set("updated_at", time.Now().Format("2006-01-02 15:04:05.999999")).
		Where(squirrel.Eq{"id": reqId})

	query, args, err = builderUpdate.ToSql()
	if err != nil {
		log.Printf("failed to build query: %d", err)
		return nil, errors.New("failed to build query")
	}

	res, err := pool.Exec(ctx, query, args...)
	log.Print(res)
	if err != nil {
		log.Printf("failed to update note information: %d", err)
		return nil, errors.New("failed to update note information")
	}
	return nil, nil
}

func (s *server) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		log.Printf("failed to connect to database: %d", err)
		return nil, errors.New("failed to connect to database")
	}
	defer pool.Close()

	reqId := req.GetId()

	builderSelect := squirrel.Select("id").
		From("note").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": reqId})

	query, args, err := builderSelect.ToSql()
	if err != nil {
		log.Printf("failed to build query: %d", err)
		return nil, errors.New("failed to build query")
	}
	var id int64
	err = pool.QueryRow(ctx, query, args...).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Printf("there is no note with such id in system: %d", err)
		return nil, errors.New("there is no note with such id in system")
	}

	builderDelete := squirrel.Delete("id").
		From("note").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": reqId})

	query, args, err = builderDelete.ToSql()
	if err != nil {
		log.Printf("failed to build query: %d", err)
		return nil, errors.New("failed to build query")
	}

	err = pool.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Printf("failed to delete note: %d", err)
		return nil, errors.New("failed to delete note")
	}
	return nil, nil

}

func (s *server) List(ctx context.Context, req *desc.ListRequest) (*desc.ListResponse, error) {
	curr := make([]*desc.Note, 0)

	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		log.Printf("failed to connect to database: %d", err)
		return nil, errors.New("failed to connect to database")
	}
	defer pool.Close()

	author := req.GetPersonId()

	builderSelect := squirrel.Select("id").
		From("persons").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": author})

	query, args, err := builderSelect.ToSql()
	if err != nil {
		log.Printf("failed to build query: %d", err)
		return nil, errors.New("failed to build query")
	}
	var id int64
	err = pool.QueryRow(ctx, query, args...).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Printf("there is no person with such id in system: %d", err)
		return nil, errors.New("there is no person with such id in system")
	} else if err != nil {
		log.Printf("failed to find personId: %d", err)
		return nil, errors.New("failed to find personId")
	}

	builderSelect = squirrel.Select("id", "title", "content", "author", "dead_line", "status", "created_at", "updated_at").
		From("note").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"author": author})

	query, args, err = builderSelect.ToSql()
	if err != nil {
		log.Printf("failed to build query: %d", err)
		return nil, errors.New("failed to build query")
	}

	var status bool
	var title, content string
	var createdAt time.Time
	var updatedAt, dead_line time.Time

	rows, err := pool.Query(ctx, query, args...)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Printf("there is no notes made by this person: %d", err)
		return nil, errors.New("there is no notes made by this person")
	}

	for rows.Next() {
		err = rows.Scan(&id, &title, &content, &author, &dead_line, &status, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("failed to scan note: %d", err)
			return nil, errors.New("failed to scan note")
		}
		n := &desc.Note{
			Id: id,
			Info: &desc.NoteInfo{
				Title:    title,
				Content:  content,
				Author:   author,
				DeadLine: timestamppb.New(dead_line),
				Status:   false,
			},
			CreatedAt: timestamppb.New(createdAt),
			UpdatedAt: timestamppb.New(updatedAt),
		}
		curr = append(curr, n)
	}
	return &desc.ListResponse{Notes: curr}, nil
}

func (s *server) CreatePerson(ctx context.Context, req *desc.CreatePersonReqest) (*desc.CreatePersonResponse, error) {
	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer pool.Close()

	author := gofakeit.Int64()
	flag := true
	for flag {
		builderSelect := squirrel.Select("id").
			From("persons").
			PlaceholderFormat(squirrel.Dollar).
			Where(squirrel.Eq{"id": author})

		query, args, err := builderSelect.ToSql()
		if err != nil {
			log.Fatalf("failed to build query: %v", err)
		}
		var id int64
		err = pool.QueryRow(ctx, query, args...).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			author = int64(gofakeit.Uint32())
			flag = false
		}
	}

	login := req.GetLogin()

	builderSelect := squirrel.Select("id").
		From("persons").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"login": login})

	query, args, err := builderSelect.ToSql()
	if err != nil {
		log.Printf("failed to build query: %d", err)
		return nil, errors.New("failed to build query")
	}
	var id int64
	err = pool.QueryRow(ctx, query, args...).Scan(&id)
	if !(errors.Is(err, pgx.ErrNoRows)) {
		log.Printf("there is a person with such login in system: %d", err)
		return nil, errors.New("there is a person with such login in system")
	}

	builderInsert := squirrel.Insert("persons").
		PlaceholderFormat(squirrel.Dollar).
		Columns("id", "login", "password").
		Values(author, req.GetLogin(), req.GetPassword())

	query, args, err = builderInsert.ToSql()
	if !errors.Is(err, pgx.ErrNoRows) && (err != nil) {
		log.Printf("failed to build query: %d", err)
		return nil, errors.New("failed to build query")
	}

	err = pool.QueryRow(ctx, query, args...).Scan(&id)
	if !errors.Is(err, pgx.ErrNoRows) && err != nil {
		log.Printf("failed to insert note: %d", err)
		return nil, errors.New("failed to insert note")
	}
	return &desc.CreatePersonResponse{Id: author}, nil
}

func (s *server) LogInPerson(ctx context.Context, req *desc.LogInPersonRequest) (*desc.LogInPersonResponce, error) {
	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer pool.Close()

	login := req.GetLogin()

	builderSelect := squirrel.Select("id", "login", "password").
		From("persons").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"login": login})

	query, args, err := builderSelect.ToSql()
	if err != nil {
		log.Fatalf("failed to build query: %v", err)
	}
	var id int64
	var password string
	err = pool.QueryRow(ctx, query, args...).Scan(&id, &login, &password)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Printf("there is no person with such login in system: %d", err)
		return nil, errors.New("there is no person with such login in system")
	} else if err != nil {
		log.Printf("failed to found person: %d", err)
		return nil, errors.New("failed to found person")
	}

	if password != req.GetPassword() {
		return nil, errors.New("incorrect password")
	} else {
		return &desc.LogInPersonResponce{Id: id}, nil
	}
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterNoteV1Server(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
