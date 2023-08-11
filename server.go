package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Product struct {
	Name string `json:"name"`
	Price float64 `json:"price"`
}

type Products []Product

type productHandler struct {
	sync.Mutex
	products Products
}

func (ph *productHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
	switch r.Method {
	case "GET":
		ph.get(w, r)
	case "POST":
		ph.post(w, r)
	case "PUT", "PATCH":
		ph.put(w, r)
	case "DELETE":
		ph.delete(w, r)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid Method")
	}
}

func (ph *productHandler) get(w http.ResponseWriter, r *http.Request){
	defer ph.Unlock()
	ph.Lock()

	id, err := idFromUrl(r)
	if err != nil {
		respondWithJSON(w, http.StatusOK, ph.products)
		return
	}

	if id >= len(ph.products) || id < 0 {
		respondWithError(w, http.StatusNotFound, "not found")
		return
	}

	respondWithJSON(w, http.StatusOK, ph.products[id])
}

func (ph *productHandler) post(w http.ResponseWriter, r *http.Request){
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ct := r.Header.Get("content-type")
	if ct != "application/json"{
		respondWithError(w, http.StatusUnsupportedMediaType, "content type 'application/json' required")
		return
	}

	var product Product
	err = json.Unmarshal(body, &product)
	if err != nil{
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer ph.Unlock()
	ph.Lock()

	ph.products = append(ph.products, product)
	respondWithJSON(w, http.StatusCreated, product)
}

func (ph *productHandler) put(w http.ResponseWriter, r *http.Request){
	defer r.Body.Close()

	id, err := idFromUrl(r)
	if err != nil {
		respondWithJSON(w, http.StatusOK, ph.products)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ct := r.Header.Get("content-type")
	if ct != "application/json"{
		respondWithError(w, http.StatusUnsupportedMediaType, "content type 'application/json' required")
		return
	}

	var product Product
	err = json.Unmarshal(body, &product)
	if err != nil{
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer ph.Unlock()
	ph.Lock()

	if id >= len(ph.products) || id < 0 {
		respondWithError(w, http.StatusNotFound, "not found")
		return
	}

	if product.Name != "" {
		ph.products[id].Name = product.Name
	}
	if product.Price != 0.0 {
		ph.products[id].Price = product.Price
	}

	respondWithJSON(w, http.StatusCreated, ph.products[id])
}

func (ph *productHandler) delete(w http.ResponseWriter, r *http.Request){
	id, err := idFromUrl(r)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "not found")
		return
	}

	defer ph.Unlock()
	ph.Lock()

	if id < 0 || id >= len(ph.products) {
		respondWithError(w, http.StatusNotFound, "not found")
		return
	}

	ph.products = append(ph.products[ : id], ph.products[id + 1 : ]...)

	respondWithJSON(w, http.StatusNoContent, "")
}


func respondWithError(w http.ResponseWriter, code int, msg string){
	respondWithJSON(w, code, map[string]string{"error": msg})
}

func respondWithJSON(w http.ResponseWriter, code int, data interface{}){
	response, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func idFromUrl(r *http.Request) (int, error) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3{
		return 0, errors.New("not found")
	}

	id, err := strconv.Atoi(parts[len(parts) - 1])
	if err != nil{
		return 0, errors.New("not found")
	}

	return id, nil
}

func newProductHandler() *productHandler {
	return &productHandler{
		products: Products{
			Product{"Shoes", 25.00},
			Product{"WebCam", 50.00},
			Product{"Mic", 20.00},
		},
	}
}


func main() {
	port := ":8080"

	ph := newProductHandler()
	http.Handle("/products", ph)
	http.Handle("/products/", ph)


	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World")
	})

	log.Fatal(http.ListenAndServe(port, nil))
}
