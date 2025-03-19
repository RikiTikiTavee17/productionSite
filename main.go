package main

import (
	"context"
	"encoding/json"
	desc "github.com/RikiTikiTavee17/productionSite/course/grpc/pkg/dish_v1"
	"github.com/go-chi/chi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Dish struct {
	Id        int64     `json:"id"`
	Info      *DishInfo `json:"info"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type DishInfo struct {
	Name        string `json:"name"`
	Price       int64  `json:"price"`
	Description string `json:"description"`
	Composition string `json:"composition"`
	Author      int64  `json:"author"`
	PhotoUrl    string `json:"photo_url"`
}

type Person struct {
	Id       int64  `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Position string `json:"position"`
}

const (
	baseUrl       = "localhost:8081"
	createDish    = "/dish"
	getDish       = "/dish/get/{dishId}"
	updateDish    = "/dish/update/{dishId}"
	deleteDish    = "/dish/delete/{dishId}"
	listDishes    = "/dishes/list"
	personsCreate = "/persons/create"
	personsChange = "/persons/change"
	personsLogIn  = "/persons/login"
)

func convertTimestampToISO8601(ts *timestamppb.Timestamp) string {
	// Преобразуем timestamp в time.Time
	t := ts.AsTime()
	// Преобразуем time.Time в формат ISO 8601
	return t.Format(time.RFC3339)
}

func getGRPCClient() (desc.DishV1Client, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(baseUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to server %v", err)
	}
	c := desc.NewDishV1Client(conn)
	return c, conn, nil
}

func createDishHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info := &DishInfo{}
	if err := json.NewDecoder(r.Body).Decode(info); err != nil {
		http.Error(w, "Failed to decode dish data", http.StatusBadRequest)
		return
	}
	client, conn, err := getGRPCClient()
	if err != nil {
		http.Error(w, "Failed to connect to server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	grpcReq := &desc.CreateRequest{
		Info: &desc.DishInfo{
			Name:        info.Name,
			Price:       info.Price,
			Description: info.Description,
			Composition: info.Composition,
			Author:      info.Author,
			PhotoUrl:    info.PhotoUrl,
		},
	}

	grpcRes, err := client.Create(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, "Failed to create dish", http.StatusInternalServerError)
		return
	}

	responce := map[string]interface{}{
		"id": grpcRes.GetId(),
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(responce); err != nil {
		http.Error(w, "Failed to encode dish data", http.StatusInternalServerError)
		return
	}
}

func getDishHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := chi.URLParam(r, "dishId")
	if id == "" {
		http.Error(w, "Missing dishId", http.StatusBadRequest)
		return
	}

	client, conn, err := getGRPCClient()
	if err != nil {
		http.Error(w, "Failed to connect to server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	newId, _ := strconv.Atoi(id)
	grpcReq := &desc.GetRequest{
		Id: int64(newId),
	}
	grpcRes, err := client.Get(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, "Failed to get dish", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":          grpcRes.GetNote().GetId(),
		"name":        grpcRes.GetNote().GetInfo().GetName(),
		"price":       grpcRes.GetNote().GetInfo().GetPrice(),
		"description": grpcRes.GetNote().GetInfo().GetDescription(),
		"composition": grpcRes.GetNote().GetInfo().GetComposition(),
		"author":      grpcRes.GetNote().GetInfo().GetAuthor(),
		"photo_url":   grpcRes.GetNote().GetInfo().GetPhotoUrl(),
		"created_at":  grpcRes.GetNote().GetCreatedAt().AsTime().Format(time.RFC3339),
		"updated_at":  grpcRes.GetNote().GetUpdatedAt().AsTime().Format(time.RFC3339),
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode dish data", http.StatusInternalServerError)
		return
	}
}

func listDishesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	client, conn, err := getGRPCClient()
	if err != nil {
		http.Error(w, "Failed to connect to server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	grpcReq := &desc.ListRequest{}
	grpcRes, err := client.List(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, "Failed to list dishes", http.StatusInternalServerError)
		return
	}
	Dishes := make([]Dish, 0)
	for _, dish := range grpcRes.GetDishes() {
		Dishes = append(Dishes, Dish{
			Id:        dish.GetId(),
			CreatedAt: convertTimestampToISO8601(dish.GetCreatedAt()),
			UpdatedAt: convertTimestampToISO8601(dish.GetUpdatedAt()),
			Info: &DishInfo{
				Name:        dish.GetInfo().GetName(),
				Price:       dish.GetInfo().GetPrice(),
				Description: dish.GetInfo().GetDescription(),
				Composition: dish.GetInfo().GetComposition(),
				Author:      dish.GetInfo().GetAuthor(),
				PhotoUrl:    dish.GetInfo().GetPhotoUrl(),
			},
		})
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(Dishes); err != nil {
		http.Error(w, "Failed to encode dishes", http.StatusInternalServerError)
		return
	}
}

func updateDishHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	client, conn, err := getGRPCClient()
	if err != nil {
		http.Error(w, "Failed to connect to server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	id := chi.URLParam(r, "dishId")
	if id == "" {
		http.Error(w, "Missing dishId", http.StatusBadRequest)
		return
	}
	newId, _ := strconv.Atoi(id)
	grpcReq := &desc.UpdateRequest{
		Id: int64(newId),
	}
}

func main() {
	r := chi.NewRouter()
	r.Post(createDish, createDishHandler)
	r.Get(getDish, getDishHandler)
	r.Get(listDishes, listDishesHandler)
	r.Patch(updateDish, updateDishHandler)
	r.Delete(deleteDish, deleteDishHandler)
	r.Post(personsCreate, personsCreateHandler)
	r.Post(personsLogIn, personsLogInHandler)
	r.Patch(personsChange, personsChangeHandler)

	err := http.ListenAndServe(baseUrl, r)
	if err != nil {
		log.Fatal(err)
	}

}
