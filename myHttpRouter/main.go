package main

import (
	"context"
	"encoding/json"
	desc "github.com/RikiTikiTavee17/productionSite/course/grpc/pkg/dish_v1"
	"github.com/go-chi/chi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Dish struct {
	Id        int32     `json:"id"`
	Info      *DishInfo `json:"info"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type DishInfo struct {
	Name        string `json:"name"`
	Price       int32  `json:"price"`
	Description string `json:"description"`
	Composition string `json:"composition"`
	Author      int32  `json:"author"`
	PhotoUrl    string `json:"photo_url"`
}

type CreatePerson struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Position string `json:"position"`
}

type LogInPerson struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type ChangePersonPosition struct {
	Position string `json:"position"`
}

type UpdateDishInfo struct {
	Name        *string `json:"name,omitempty"`
	Price       *int64  `json:"price,omitempty"`
	Description *string `json:"description,omitempty"`
	Composition *string `json:"composition,omitempty"`
	Author      *int64  `json:"author,omitempty"`
	PhotoUrl    *string `json:"photo_url,omitempty"`
}

const (
	baseUrl       = "localhost:8081"
	grpcUrl       = "localhost:50051"
	createDish    = "/dish"
	getDish       = "/dish/get/{dishId}"
	updateDish    = "/dish/update/{dishId}"
	deleteDish    = "/dish/delete/{dishId}"
	listDishes    = "/dishes/list"
	personsCreate = "/persons/create"
	personsChange = "/persons/change/{personId}"
	personsLogIn  = "/persons/login"
)

func convertTimestampToISO8601(ts *timestamppb.Timestamp) string {
	// Преобразуем timestamp в time.Time
	t := ts.AsTime()
	// Преобразуем time.Time в формат ISO 8601
	return t.Format(time.RFC3339)
}

func getGRPCClient() (desc.DishV1Client, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(grpcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	response := map[string]interface{}{
		"id": grpcRes.GetId(),
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
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
		Id: int32(newId),
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

	id := chi.URLParam(r, "dishId")
	if id == "" {
		http.Error(w, "Missing dishId", http.StatusBadRequest)
		return
	}

	var req UpdateDishInfo
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	client, conn, err := getGRPCClient()
	if err != nil {
		http.Error(w, "Failed to connect to server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	newId, _ := strconv.Atoi(id)

	grpcReq := &desc.UpdateRequest{
		Id:   int32(newId),
		Info: &desc.UpdateDishInfo{},
	}
	if req.Name != nil {
		grpcReq.Info.Name = wrapperspb.String(*req.Name)
	}
	if req.Price != nil {
		grpcReq.Info.Price = wrapperspb.Int64(*req.Price)
	}
	if req.Description != nil {
		grpcReq.Info.Description = wrapperspb.String(*req.Description)
	}
	if req.Composition != nil {
		grpcReq.Info.Composition = wrapperspb.String(*req.Composition)
	}
	if req.Author != nil {
		grpcReq.Info.Author = wrapperspb.Int64(*req.Author)
	}
	if req.PhotoUrl != nil {
		grpcReq.Info.PhotoUrl = wrapperspb.String(*req.PhotoUrl)
	}

	_, err = client.Update(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, "Failed to update dish", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func deleteDishHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	id := chi.URLParam(r, "dishId")
	if id == "" {
		http.Error(w, "Missing dishId", http.StatusBadRequest)
		return
	}
	newId, _ := strconv.Atoi(id)

	client, conn, err := getGRPCClient()
	if err != nil {
		http.Error(w, "Failed to connect to server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	grpcReq := &desc.DeleteRequest{
		Id: int32(newId),
	}
	_, err = client.Delete(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, "Failed to delete dish", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func personsCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info := &CreatePerson{}
	if err := json.NewDecoder(r.Body).Decode(info); err != nil {
		http.Error(w, "Failed to decode person data", http.StatusBadRequest)
		return
	}
	client, conn, err := getGRPCClient()
	if err != nil {
		http.Error(w, "Failed to connect to server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	grpcReq := &desc.CreatePersonReqest{
		Login:    info.Login,
		Password: info.Password,
		Position: info.Position,
	}
	size := proto.Size(grpcReq)
	log.Print(size)
	grpcRes, err := client.CreatePerson(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, "Failed to create person", http.StatusInternalServerError)
		log.Printf("failed to create person: %v", err)
		return
	}

	response := map[string]interface{}{
		"id": grpcRes.GetId(),
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode person data", http.StatusInternalServerError)
		return
	}
}

func personsLogInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info := &LogInPerson{}
	if err := json.NewDecoder(r.Body).Decode(info); err != nil {
		http.Error(w, "Failed to decode person data", http.StatusBadRequest)
		return
	}

	client, conn, err := getGRPCClient()
	if err != nil {
		http.Error(w, "Failed to connect to server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	grpcReq := &desc.LogInPersonRequest{
		Login:    info.Login,
		Password: info.Password,
	}

	grpcRes, err := client.LogInPerson(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, "Failed to create person", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":       grpcRes.GetId(),
		"position": grpcRes.GetPosition(),
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode person data", http.StatusInternalServerError)
		return
	}
}

func personsChangeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := chi.URLParam(r, "personId")
	if id == "" {
		http.Error(w, "Missing personId", http.StatusBadRequest)
		return
	}

	info := &ChangePersonPosition{}
	if err := json.NewDecoder(r.Body).Decode(info); err != nil {
		http.Error(w, "Failed to decode person data", http.StatusBadRequest)
		return
	}

	if info.Position == "" {
		http.Error(w, "Missing position", http.StatusBadRequest)
		return
	}

	client, conn, err := getGRPCClient()
	if err != nil {
		http.Error(w, "Failed to connect to server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	newId, _ := strconv.Atoi(id)

	grpcReq := &desc.ChangePersonPositionRequest{
		Id:       int32(newId),
		Position: info.Position,
	}

	grpcRes, err := client.ChangePersonPosition(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, "Failed to update person", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"position": grpcRes.GetPosition(),
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode person data", http.StatusInternalServerError)
		return
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
