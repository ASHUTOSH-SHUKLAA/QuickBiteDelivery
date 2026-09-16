package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

// Restaurant struct
type Restaurant struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// MenuItem struct
type MenuItem struct {
	ID           int     `json:"id"`
	RestaurantID int     `json:"restaurant_id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
}

// Order struct
type Order struct {
	ID           int     `json:"id"`
	RestaurantID int     `json:"restaurant_id"`
	CustomerName string  `json:"customer_name"`
	Status       string  `json:"status"`
	TotalAmount  float64 `json:"total_amount"`
}

func main() {
	initDB()
	defer db.Close()

	r := mux.NewRouter()

	// Routes
	r.HandleFunc("/restaurants", getRestaurants).Methods("GET")
	r.HandleFunc("/restaurants/{id}/menu", getMenu).Methods("GET")
	r.HandleFunc("/orders", createOrder).Methods("POST")
	r.HandleFunc("/orders/{id}", getOrder).Methods("GET")
	r.HandleFunc("/orders/{id}/status", updateOrderStatus).Methods("PUT")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server is running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func initDB() {
	var err error
	
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Fallback for local development
		connStr = "user=postgres password=postgres dbname=quickbite sslmode=disable"
	}
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Println("Could not connect to database, ensure PostgreSQL is running and credentials are correct. Error:", err)
		return
	}
	fmt.Println("Connected to PostgreSQL!")

	// Create tables if they don't exist
	createTables()
}

func createTables() {
	query := `
	CREATE TABLE IF NOT EXISTS restaurants (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL
	);
	CREATE TABLE IF NOT EXISTS menu_items (
		id SERIAL PRIMARY KEY,
		restaurant_id INT REFERENCES restaurants(id),
		name VARCHAR(100) NOT NULL,
		price DECIMAL(10, 2) NOT NULL
	);
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		restaurant_id INT REFERENCES restaurants(id),
		customer_name VARCHAR(100) NOT NULL,
		status VARCHAR(50) DEFAULT 'Pending',
		total_amount DECIMAL(10, 2) NOT NULL
	);
	
	-- Insert some dummy data if empty
	INSERT INTO restaurants (name) 
	SELECT 'Burger King' WHERE NOT EXISTS (SELECT 1 FROM restaurants WHERE name = 'Burger King');
	
	INSERT INTO menu_items (restaurant_id, name, price)
	SELECT 1, 'Whopper', 5.99 WHERE NOT EXISTS (SELECT 1 FROM menu_items WHERE name = 'Whopper');
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Println("Error creating tables:", err)
	}
}

// Handlers
func getRestaurants(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name FROM restaurants")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var restaurants []Restaurant
	for rows.Next() {
		var rest Restaurant
		if err := rows.Scan(&rest.ID, &rest.Name); err != nil {
			continue
		}
		restaurants = append(restaurants, rest)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurants)
}

func getMenu(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	restaurantID := vars["id"]

	rows, err := db.Query("SELECT id, restaurant_id, name, price FROM menu_items WHERE restaurant_id = $1", restaurantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var menu []MenuItem
	for rows.Next() {
		var item MenuItem
		if err := rows.Scan(&item.ID, &item.RestaurantID, &item.Name, &item.Price); err != nil {
			continue
		}
		menu = append(menu, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menu)
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Simplified order logic: just insert it
	err := db.QueryRow(
		"INSERT INTO orders (restaurant_id, customer_name, total_amount, status) VALUES ($1, $2, $3, 'Pending') RETURNING id",
		order.RestaurantID, order.CustomerName, order.TotalAmount,
	).Scan(&order.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	order.Status = "Pending"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	var order Order
	err := db.QueryRow("SELECT id, restaurant_id, customer_name, status, total_amount FROM orders WHERE id = $1", orderID).
		Scan(&order.ID, &order.RestaurantID, &order.CustomerName, &order.Status, &order.TotalAmount)
	
	if err == sql.ErrNoRows {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	var update struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE orders SET status = $1 WHERE id = $2", update.Status, orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message":"Order status updated successfully"}`)
}
