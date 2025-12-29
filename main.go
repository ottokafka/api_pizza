package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./pizza.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	initDB(db)

	// --- 1. LANDING PAGE SERVER (Port 9002) ---
	landingMux := http.NewServeMux()
	landingMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("landing_page.html"))
		tmpl.Execute(w, nil)
	})
	landingMux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))

	// --- 2. ORDERING APP SERVER (Port 9001) ---
	orderMux := http.NewServeMux()

	// Customer Routes
	orderMux.HandleFunc("/", handleIndex)
	orderMux.HandleFunc("/menu", handleGetMenu)
	orderMux.HandleFunc("/cart/add", handleAddToCart)
	orderMux.HandleFunc("/cart/clear", handleClearCart)
	orderMux.HandleFunc("/checkout", handleCheckout)

	// Success Page
	orderMux.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		cart = []CartItem{}
		fmt.Fprint(w, `
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0">
				<script src="https://cdn.tailwindcss.com"></script>
			</head>
			<body class="bg-gray-50 flex items-center justify-center h-screen">
				<div class="bg-white p-8 rounded-xl shadow-lg text-center max-w-md">
					<div class="text-6xl mb-4">🎉</div>
					<h1 class="text-2xl font-bold text-gray-800 mb-2">Payment Successful!</h1>
					<p class="text-gray-600 mb-6">Your order has been sent to the kitchen. Pickup in ~20 mins.</p>
					<a href='/' class="inline-block bg-orange-600 text-white px-6 py-2 rounded-lg font-medium hover:bg-orange-700 transition">Order More</a>
				</div>
			</body>
			</html>
		`)
	})

	// Static Assets
	orderMux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))

	// Admin Routes
	orderMux.HandleFunc("/admin", handleAdminPage)
	orderMux.HandleFunc("/admin/update", handleAdminUpdateProduct)
	orderMux.HandleFunc("/admin/create", handleAdminCreateProduct)
	orderMux.HandleFunc("/admin/delete", handleAdminDeleteProduct)
	orderMux.HandleFunc("/admin/generate-image", handleAdminGenerateImage)

	// Kitchen Routes
	orderMux.HandleFunc("/kitchen", handleKitchenPage)
	orderMux.HandleFunc("/kitchen/orders", handleGetKitchenOrders)
	orderMux.HandleFunc("/kitchen/status", handleKitchenStatus)

	go func() {
		fmt.Println("SEO Landing Page: http://localhost:9002")
		http.ListenAndServe(":9002", landingMux)
	}()

	fmt.Println("Ordering App: http://localhost:9001")
	log.Fatal(http.ListenAndServe(":9001", orderMux))
}

// handleIndex parses both the layout (index.html) and the specific view (customer.html)
func handleIndex(w http.ResponseWriter, r *http.Request) {
	// We parse both files so index.html can use {{template "content" .}} defined in customer.html
	tmpl := template.Must(template.ParseFiles("index.html", "customer.html"))
	tmpl.Execute(w, nil)
}
