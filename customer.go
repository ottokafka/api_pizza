package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Product struct {
	ID          int
	Category    string
	Name        string
	Description string
	Price       float64
	ImageURL    string
	TypeTag     string
	InStock     bool
}

type CartItem struct {
	Name       string
	BasePrice  float64
	AddonTotal float64
	Options    []string
}

func (c CartItem) Total() float64 { return c.BasePrice + c.AddonTotal }

var db *sql.DB
var cart []CartItem

const taxRate = 0.05

// handleGetMenu generates the grid of products
func handleGetMenu(w http.ResponseWriter, r *http.Request) {
	rows, _ := db.Query("SELECT id, category, name, description, price, image_url, type_tag, in_stock FROM products")
	defer rows.Close()

	categories := map[string][]Product{}
	for rows.Next() {
		var p Product
		rows.Scan(&p.ID, &p.Category, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.TypeTag, &p.InStock)
		categories[p.Category] = append(categories[p.Category], p)
	}

	order := []string{"pizza", "pasta", "drink"}
	for _, cat := range order {
		products := categories[cat]
		if len(products) == 0 {
			continue
		}

		// Section Header with ID for anchor links
		fmt.Fprintf(w, `
			<section id='%s' class='scroll-mt-28 mb-10'>
				<div class="flex items-center gap-4 mb-6">
					<h2 class='text-2xl font-bold uppercase tracking-tight text-gray-800'>%s</h2>
					<div class="h-1 bg-gray-200 flex-grow rounded-full"></div>
				</div>
				<div class='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-2 xl:grid-cols-2 gap-6'>`,
			cat, strings.ToUpper(cat))

		for _, p := range products {
			renderProductCard(w, p)
		}
		fmt.Fprintf(w, "</div></section>")
	}
}

// renderProductCard generates the HTML for a single item card
func renderProductCard(w http.ResponseWriter, p Product) {
	optionsHTML := ""

	// Reusable styling for option chips using Tailwind
	// Note: We use has-[:checked] to style the label based on the hidden input state
	radioStyle := `class="cursor-pointer border border-gray-200 rounded-full px-3 py-1 text-xs font-medium text-gray-600 bg-white shadow-sm hover:bg-gray-50 has-[:checked]:bg-orange-50 has-[:checked]:text-brand has-[:checked]:border-brand transition-all select-none"`

	switch p.TypeTag {
	case "pizza_opt":
		optionsHTML = fmt.Sprintf(`
			<div class="mt-3 space-y-2">
				<p class="text-xs font-bold text-gray-500 uppercase">Add-ons</p>
				<div class="flex flex-wrap gap-2">
					<label %s><input type="checkbox" name="extra_cheese" class="hidden"><span>🧀 Ex. Cheese</span></label>
					<label %s><input type="checkbox" name="extra_topping" class="hidden"><span>🍕 Ex. Topping</span></label>
				</div>
			</div>`, radioStyle, radioStyle)
	case "coffee_opt":
		optionsHTML = fmt.Sprintf(`
			<div class="mt-3 space-y-2">
				<!-- Temp -->
				<div class="flex gap-2">
					<label %s><input type="radio" name="temp" value="Ice" checked class="hidden"><span>❄️ Ice</span></label>
					<label %s><input type="radio" name="temp" value="Hot" class="hidden"><span>🔥 Hot</span></label>
				</div>
				<!-- Sweetness -->
				<div class="flex flex-wrap gap-2">
					<label %s><input type="radio" name="sweetness" value="Regular" checked class="hidden"><span>100%%</span></label>
					<label %s><input type="radio" name="sweetness" value="Less Sweet" class="hidden"><span>50%%</span></label>
					<label %s><input type="radio" name="sweetness" value="Least Sweet" class="hidden"><span>0%%</span></label>
				</div>
			</div>`, radioStyle, radioStyle, radioStyle, radioStyle, radioStyle)
	case "pasta_opt":
		optionsHTML = fmt.Sprintf(`
			<div class="mt-3">
				<label %s><input type="checkbox" name="extra_pasta" class="hidden"><span>🍝 Extra Portion (+RM3)</span></label>
			</div>`, radioStyle)
	}

	// Logic for Button and Availability
	// Note: hx-target="#desktop-cart-status" targets the ID inside index.html.
	// The javascript in index.html then syncs this to the mobile drawer.
	btnClass := "w-full bg-brand hover:bg-brand-dark text-white font-bold py-2 px-4 rounded-lg shadow-md active:scale-95 transition-all flex justify-center items-center gap-2"
	btnText := "Add to Order <span>+</span>"
	cardOpacity := ""
	disabledAttr := ""

	if !p.InStock {
		btnClass = "w-full bg-gray-200 text-gray-400 font-bold py-2 px-4 rounded-lg cursor-not-allowed"
		btnText = "Sold Out"
		cardOpacity = "opacity-60 grayscale"
		disabledAttr = "disabled"
	}

	fmt.Fprintf(w, `
		<form hx-post="/cart/add?id=%d" hx-target="#desktop-cart-status" class="bg-white rounded-xl shadow border border-gray-100 overflow-hidden flex flex-col h-full hover:shadow-lg transition-shadow duration-300 %s">
			<div class="relative h-48 overflow-hidden bg-gray-100 group">
				<img src="%s" alt="%s" loading="lazy" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500">
				<div class="absolute bottom-2 right-2 bg-white/90 backdrop-blur-sm px-2 py-1 rounded text-sm font-bold text-gray-900 shadow-sm">
					RM%.2f
				</div>
			</div>
			
			<div class="p-4 flex flex-col flex-grow">
				<div class="flex-grow">
					<h3 class="font-bold text-lg text-gray-800 leading-tight">%s</h3>
					<p class="text-sm text-gray-500 mt-1 line-clamp-2">%s</p>
					%s
				</div>
				
				<div class="mt-4 pt-4 border-t border-gray-50">
					<button type="submit" class="%s" %s>
						%s
					</button>
				</div>
			</div>
		</form>`,
		p.ID, cardOpacity, p.ImageURL, p.Name, p.Price, p.Name, p.Description, optionsHTML, btnClass, disabledAttr, btnText)
}

func handleAddToCart(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := r.URL.Query().Get("id")
	var p Product
	err := db.QueryRow("SELECT name, price, category FROM products WHERE id = ?", id).Scan(&p.Name, &p.Price, &p.Category)
	if err != nil {
		return
	}

	item := CartItem{Name: p.Name, BasePrice: p.Price}

	// Logic for Add-ons
	if r.FormValue("extra_cheese") == "on" {
		item.Options = append(item.Options, "Extra Cheese")
		item.AddonTotal += 3.0
	}
	if r.FormValue("extra_topping") == "on" {
		item.Options = append(item.Options, "Extra Topping")
		item.AddonTotal += 5.0
	}
	if r.FormValue("extra_pasta") == "on" {
		item.Options = append(item.Options, "Extra Pasta")
		item.AddonTotal += 3.0
	}
	if sw := r.FormValue("sweetness"); sw != "" {
		item.Options = append(item.Options, sw)
	}
	if t := r.FormValue("temp"); t != "" {
		item.Options = append(item.Options, t)
	}

	cart = append(cart, item)
	renderCart(w)
}

func handleClearCart(w http.ResponseWriter, r *http.Request) {
	cart = []CartItem{}
	renderCart(w)
}

func renderCart(w http.ResponseWriter) {
	if len(cart) == 0 {
		fmt.Fprint(w, `
			<div class="flex flex-col items-center justify-center py-10 text-gray-400">
				<svg class="w-12 h-12 mb-2 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z"></path></svg>
				<p>Your cart is empty.</p>
			</div>
		`)
		return
	}

	subtotal := 0.0
	fmt.Fprint(w, `<ul class="divide-y divide-gray-100 max-h-[50vh] overflow-y-auto mb-4 custom-scrollbar">`)

	for _, item := range cart {
		subtotal += item.Total()
		opts := ""
		if len(item.Options) > 0 {
			opts = fmt.Sprintf(`<div class="text-xs text-gray-500 mt-0.5">%s</div>`, strings.Join(item.Options, ", "))
		}

		fmt.Fprintf(w, `
			<li class="py-3 flex justify-between group">
				<div>
					<div class="font-medium text-gray-800 text-sm">%s</div>
					%s
				</div>
				<span class="font-bold text-gray-700 text-sm">RM%.2f</span>
			</li>`, item.Name, opts, item.Total())
	}

	tax := subtotal * taxRate
	total := subtotal + tax

	fmt.Fprintf(w, `</ul>

		<div class="bg-gray-50 rounded-lg p-4 space-y-2 border border-gray-100">
			<div class="flex justify-between text-sm text-gray-600">
				<span>Subtotal</span><span>RM%.2f</span>
			</div>
			<div class="flex justify-between text-sm text-gray-600">
				<span>Tax (5%%)</span><span>RM%.2f</span>
			</div>
			<div class="flex justify-between text-lg font-bold text-gray-900 border-t border-gray-200 pt-2 mt-1">
				<span>Total</span><span class="grand-total-value">RM%.2f</span>
			</div>
		</div>

		<div class="mt-6 space-y-3">
			<form action="/checkout" method="post">
				<button type="submit" class="w-full bg-gray-900 hover:bg-black text-white font-bold py-3 px-4 rounded-lg shadow-lg hover:shadow-xl transition-all transform active:scale-95 flex justify-center items-center gap-2">
					<span>Checkout & Pay</span>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3"></path></svg>
				</button>
			</form>
			<button hx-post="/cart/clear" hx-target="#desktop-cart-status" 
				class="w-full text-xs text-gray-400 hover:text-red-500 underline decoration-dotted transition-colors">
				Clear Order
			</button>
		</div>`, subtotal, tax, total)
}

func handleCheckout(w http.ResponseWriter, r *http.Request) {
	if len(cart) == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	total := 0.0
	for _, item := range cart {
		total += item.Total()
	}
	totalWithTax := total + (total * taxRate)

	// Save to DB
	res, err := db.Exec("INSERT INTO orders (customer_name, total_amount, status) VALUES (?, ?, ?)", "Guest Customer", totalWithTax, "Paid")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	orderID, _ := res.LastInsertId()

	for _, item := range cart {
		opts := strings.Join(item.Options, ", ")
		_, err = db.Exec("INSERT INTO order_items (order_id, product_name, options, price) VALUES (?, ?, ?, ?)",
			orderID, item.Name, opts, item.Total())
		if err != nil {
			log.Printf("Error saving item: %v", err)
		}
	}

	// Redirect to success
	http.Redirect(w, r, "/success", http.StatusSeeOther)
}
