package main

import (
	"log"
	"net/http"
	"regexp"

	"github.com/go-redis/redis/v8"
)

var products map[string]*Product

func init() {
	rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	GenerateProduct()
}

func main() {
	log.Printf("Starting on port 1234")
	http.HandleFunc("/product", handleProduct)
	log.Fatal(http.ListenAndServe("127.0.0.1:1234", nil))
}

func handleProduct(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if match, _ := regexp.MatchString(`^[A-Z]{2}[0-9]{2}$`, code); !match {
		http.Error(w, "code is invalid", http.StatusBadRequest)
		return
	}

	result := findProduct(products, code)

	if result.Code == "" {
		http.Error(w, "data not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<font size="10">Product Code : ` + result.Code + ` Name :` + result.Name + `</font>`))
}

func findProduct(Products map[string]*Product, code string) Product {
	if v, ok := Products[code]; ok {
		return *v
	}

	return Product{}
}
