// Package Classification Product API
// Documentation for Product API
//  Schemes: http
//  BasePath: /
//  Version: 1.0.0
//
//  Consumes:
//  - application/json
//
//  Produces:
//  - application/json
//
// swagger:meta

package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"simpleMSs/data"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/jackc/pgx/v5"
)

type Products struct {
	l  *log.Logger
	db *pgx.Conn
}

func NewProducts(l *log.Logger, db *pgx.Conn) *Products {
	return &Products{l, db}
}

// swagger:route GET /products products listProducts
// Returns a list of products
// responses:
//
//	200: productsResponse
//	500: errorResponse
func (p *Products) GetProducts(w http.ResponseWriter, r *http.Request) {
	p.l.Println("Handle GET request")

	vars := mux.Vars(r)
	if id, ok := vars["id"]; ok {
		p.getProduct(id, w) // now returning just json, no database row
		return
	}

	rows, err := p.db.Query(context.Background(), "select id, name, description, price, sku from products")
	if err != nil {
		p.l.Println("[ERROR] querying products", err)
		http.Error(w, "Unable to query products", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	listOfProducts := data.Products{}
	for rows.Next() {
		prod := data.Product{}
		err := rows.Scan(&prod.ID, &prod.Name, &prod.Description, &prod.Price, &prod.SKU)
		if err != nil {
			p.l.Println("[ERROR] scanning product", err)
			http.Error(w, "Unable to scan product", http.StatusInternalServerError)
			return
		}
		listOfProducts = append(listOfProducts, &prod)
	}

	err = listOfProducts.ToJSON(w)
	if err != nil {
		p.l.Println("[ERROR]Error JSON encoding", err)
		http.Error(w, "Enable to marshal json", http.StatusInternalServerError)
	}
}

func (p *Products) getProduct(idStr string, w http.ResponseWriter) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		p.l.Println("[ERROR] converting id", err)
		http.Error(w, "Unable to convert id", http.StatusBadRequest)
		return
	}

	p.l.Printf("Get single product with id: %d", id)

	prod := data.Product{}
	err = p.db.QueryRow(context.Background(),
		"select id, name, description, price, sku from products where id = $1", id).Scan(&prod.ID, &prod.Name, &prod.Description, &prod.Price, &prod.SKU)

	if err != nil {
		p.l.Println("[ERROR] querying product", err)
		http.Error(w, "Unable to query product", http.StatusInternalServerError)
		return
	}

	err = prod.ToJSON(w)
	if err != nil {
		p.l.Println("[ERROR] serializing product", err)
		http.Error(w, "Error serializing product", http.StatusInternalServerError)
	}
}

func (p *Products) AddProduct(w http.ResponseWriter, r *http.Request) {
	p.l.Println("Handle POST request")
	prod := r.Context().Value(KeyProduct{}).(data.Product)

	_, err := p.db.Exec(context.Background(),
		"insert into products (name, description, price, sku) values ($1, $2, $3, $4)",
		prod.Name, prod.Description, prod.Price, prod.SKU)

	if err != nil {
		p.l.Println("[ERROR] inserting product", err)
		http.Error(w, "Error creating product", http.StatusInternalServerError)
		return
	}
}

func (p *Products) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		p.l.Fatal()
		http.Error(w, "Unable to convert id", http.StatusBadRequest)
		return
	}
	p.l.Println("Handle DELETE request with id ", id)

	_, err = p.db.Exec(context.Background(),
		"delete from products where id = $1", id)

	if err != nil {
		p.l.Println("[ERROR] deleting product", err)
		http.Error(w, "Error deleting product", http.StatusInternalServerError)
		return
	}
}

func (p *Products) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		p.l.Fatal()
		http.Error(w, "Unable to convert id", http.StatusBadRequest)
		return
	}

	p.l.Println("Handle PUT request", id)

	prod := r.Context().Value(KeyProduct{}).(data.Product)

	_, err = p.db.Exec(context.Background(),
		"update products set name = $1, description = $2, price = $3, sku = $4 where id = $5",
		prod.Name, prod.Description, prod.Price, prod.SKU, id)

	if err != nil {
		p.l.Println("[ERROR] updating product", err)
		http.Error(w, "Error updating product", http.StatusInternalServerError)
		return
	}
}

type KeyProduct struct{}

func (p Products) MiddlewareValidateProduct(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		prod := data.Product{}

		err := prod.FromJSON(r.Body)
		if err != nil {
			p.l.Println("[ERROR] deserializing product", err)
			http.Error(rw, "Error reading product", http.StatusBadRequest)
			return
		}

		err = prod.Validate()
		if err != nil {
			p.l.Println("[ERROR] validating product", err)
			http.Error(rw, fmt.Sprintf("Error validating product: %s", err), http.StatusBadRequest)
			return
		}

		// add the product to the context
		ctx := context.WithValue(r.Context(), KeyProduct{}, prod)
		r = r.WithContext(ctx)

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(rw, r)
	})
}
