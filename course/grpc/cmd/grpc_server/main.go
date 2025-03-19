package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	desc "github.com/RikiTikiTavee17/productionSite/course/grpc/pkg/dish_v1"
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
	desc.UnimplementedDishV1Server
}

func (s *server) Get(ctx context.Context, req *desc.GetRequest) (*desc.GetResponse, error) {
	log.Printf("Note id: %d", req.GetId())

	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		log.Printf("failed to connect to database: %d", err)
		return nil, errors.New("failed to connect to database")

	}
	defer pool.Close()

	builderSelectOne := squirrel.Select("id", "name", "price", "description", "composition", "author", "photo_url", "created_at", "updated_at").
		From("note").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": req.GetId()}).
		Limit(1)

	query, args, err := builderSelectOne.ToSql()
	if err != nil {
		log.Printf("failed to build query: %d", err)
		return nil, errors.New("failed to build query")
	}

	var id, author, price int64
	var description, composition, photo_url, name string
	var createdAt time.Time
	var updatedAt time.Time

	err = pool.QueryRow(ctx, query, args...).Scan(&id, &name, &price, &description, &composition, &author, &photo_url, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("no note with such id in system: %d", err)
			return nil, errors.New("no note with such id in system")
		} else {
			log.Printf("failed to select notes: %d", err)
			return nil, errors.New("failed to select notes")
		}
	}
	n := &desc.Dish{
		Id: id,
		Info: &desc.DishInfo{
			Name:        name,
			Price:       price,
			Description: description,
			Composition: composition,
			Author:      author,
			PhotoUrl:    photo_url,
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
		Columns("id", "name", "price", "description", "composition", "author", "photo_url", "created_at", "updated_at").
		Values(id, req.GetInfo().GetName(), req.GetInfo().GetPrice(), a, req.GetInfo().GetDescription(), req.GetInfo().GetComposition(), req.GetInfo().GetAuthor(), req.GetInfo().GetPhotoUrl(), time.Now().Format("2006-01-02 15:04:05.999999"), time.Now().Format("2006-01-02 15:04:05.999999"))

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

	builderSelect := squirrel.Select("id", "name", "price", "description", "composition", "author", "photo_url").
		From("note").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": reqId})

	query, args, err := builderSelect.ToSql()
	if err != nil {
		log.Printf("failed to build query: %d", err)
		return nil, errors.New("failed to build query")
	}
	var id, author, price int64
	var description, composition, photo_url, name string

	err = pool.QueryRow(ctx, query, args...).Scan(&id, &name, &price, &description, &composition, &author, &photo_url)
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
	var newName string
	if req.GetInfo().GetName() != nil {
		newName = req.GetInfo().GetName().GetValue()
	} else {
		newName = name
	}

	var newPrice int64
	if req.GetInfo().GetPrice() != nil {
		newPrice = req.GetInfo().GetPrice().GetValue()
	} else {
		newPrice = price
	}

	var newDescription string
	if req.GetInfo().GetDescription() != nil {
		newDescription = req.GetInfo().GetDescription().GetValue()
	} else {
		newDescription = description
	}

	var newComposition string
	if req.GetInfo().GetComposition() != nil {
		newComposition = req.GetInfo().GetComposition().GetValue()
	} else {
		newComposition = composition
	}

	var newAuthor int64
	if req.GetInfo().GetAuthor() != nil {
		newAuthor = req.GetInfo().GetAuthor().GetValue()
	} else {
		newAuthor = author
	}

	var newPhotoUrl string
	if req.GetInfo().GetPhotoUrl() != nil {
		newPhotoUrl = req.GetInfo().GetPhotoUrl().GetValue()
	} else {
		newPhotoUrl = photo_url
	}

	builderUpdate := squirrel.Update("note").
		PlaceholderFormat(squirrel.Dollar).
		Set("name", newName).
		Set("price", newPrice).
		Set("description", newDescription).
		Set("composition", newComposition).
		Set("author", newAuthor).
		Set("phoyo_url", newPhotoUrl).
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
	curr := make([]*desc.Dish, 0)

	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		log.Printf("failed to connect to database: %d", err)
		return nil, errors.New("failed to connect to database")
	}
	defer pool.Close()

	builderSelect := squirrel.Select("id", "name", "price", "description", "composition", "author", "photo_url", "created_at", "updated_at").
		From("note").
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := builderSelect.ToSql()
	if err != nil {
		log.Printf("failed to build query: %d", err)
		return nil, errors.New("failed to build query")
	}

	var id, author, price int64
	var description, composition, photo_url, name string
	var createdAt time.Time
	var updatedAt time.Time

	rows, err := pool.Query(ctx, query, args...)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Printf("there is no notes made by this person: %d", err)
		return nil, errors.New("there is no notes made by this person")
	}

	for rows.Next() {
		err = rows.Scan(&id, &name, &price, &description, &composition, &author, &photo_url, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("failed to scan note: %d", err)
			return nil, errors.New("failed to scan note")
		}
		n := &desc.Dish{
			Id: id,
			Info: &desc.DishInfo{
				Name:        name,
				Price:       price,
				Description: description,
				Composition: composition,
				Author:      author,
				PhotoUrl:    photo_url,
			},
			CreatedAt: timestamppb.New(createdAt),
			UpdatedAt: timestamppb.New(updatedAt),
		}
		curr = append(curr, n)
	}
	return &desc.ListResponse{Dishes: curr}, nil
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
		Columns("id", "login", "password", "position").
		Values(author, req.GetLogin(), req.GetPassword(), req.GetPosition())

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

	builderSelect := squirrel.Select("id", "login", "password", "position").
		From("persons").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"login": login})

	query, args, err := builderSelect.ToSql()
	if err != nil {
		log.Fatalf("failed to build query: %v", err)
	}
	var id int64
	var password, position string
	err = pool.QueryRow(ctx, query, args...).Scan(&id, &login, &password, &position)
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
		return &desc.LogInPersonResponce{Id: id, Position: position}, nil
	}
}

func (s *server) ChangePersonPosition(ctx context.Context, req *desc.ChangePersonPositionRequest) (*desc.ChangePersonPositionResponse, error) {
	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		log.Printf("failed to connect to database: %d", err)
		return nil, errors.New("failed to connect to database")
	}
	defer pool.Close()

	auth := req.GetId()

	builderSelect := squirrel.Select("id").
		From("persons").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": auth})

	query, args, err := builderSelect.ToSql()
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

	builderUpdate := squirrel.Update("persons").
		PlaceholderFormat(squirrel.Dollar).
		Set("position", req.GetPosition()).
		Where(squirrel.Eq{"id": auth})

	query, args, err = builderUpdate.ToSql()
	if err != nil {
		log.Printf("failed to build query: %d", err)
		return nil, errors.New("failed to build query")
	}

	res, err := pool.Exec(ctx, query, args...)
	log.Print(res)
	if err != nil {
		log.Printf("failed to update person information: %d", err)
		return nil, errors.New("failed to update person information")
	}
	return &desc.ChangePersonPositionResponse{Position: req.GetPosition()}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterDishV1Server(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
