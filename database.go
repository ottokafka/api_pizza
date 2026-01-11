package main

import (
	"database/sql"
	"log"
)

func initDB(db *sql.DB) {

	// 2. Clear old data
	// db.Exec("DELETE FROM products")

	// 1. Updated Schema to include 'type_tag' and 'in_stock'
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		category TEXT,
		name TEXT,
		description TEXT,
		price REAL,
		image_url TEXT,
		type_tag TEXT,
		in_stock BOOLEAN
	)`)

	// 1. Create ORDERS Table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		customer_name TEXT,
		total_amount REAL,
		status TEXT DEFAULT 'Paid', -- Paid, Completed
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)

	// 2. Create ORDER ITEMS Table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS order_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id INTEGER,
		product_name TEXT,
		options TEXT,
		price REAL,
		FOREIGN KEY(order_id) REFERENCES orders(id)
	)`)

	// 4. Create CATEGORIES Table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS categories (
    name TEXT PRIMARY KEY
)`)

	// Seed default categories if table is empty
	var count int
	db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if count == 0 {
		// Insert defaults to keep existing functionality
		db.Exec(`INSERT INTO categories (name) VALUES ('pizza'), ('pasta'), ('drink'), ('dessert')`)
	}
	if err != nil {
		log.Fatal(err)
	}

	// NOTE: Run new_start_data if need Brand new start if database is deleted
	new_start_data()

}
