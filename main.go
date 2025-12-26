package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stripe/stripe-go/v84"
)

const stripeSecretKey = ""

type Product struct {
	ID          int
	Category    string
	Name        string
	Description string
	Price       float64
	ImageURL    string
	TypeTag     string // New field
	InStock     bool   // New field
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
	orderMux.HandleFunc("/", handleIndex)
	orderMux.HandleFunc("/menu", handleGetMenu)
	orderMux.HandleFunc("/cart/add", handleAddToCart)
	orderMux.HandleFunc("/cart/clear", handleClearCart)
	// STRIPE CHECKOUT: KEEP THIS
	// orderMux.HandleFunc("/checkout", handleCreateCheckoutSession)
	orderMux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("."))))
	orderMux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))

	// Admin Routes
	orderMux.HandleFunc("/admin", handleAdminPage)
	orderMux.HandleFunc("/admin/update", handleAdminUpdateProduct)
	orderMux.HandleFunc("/admin/create", handleAdminCreateProduct) // POST to create

	// --- KITCHEN ROUTES ---
	orderMux.HandleFunc("/kitchen", handleKitchenPage)             // The View
	orderMux.HandleFunc("/kitchen/orders", handleGetKitchenOrders) // The Poller
	orderMux.HandleFunc("/kitchen/status", handleKitchenStatus)    // The Action

	// TEMP TESTING orders
	orderMux.HandleFunc("/checkout", handleCheckout)

	// Success Route
	orderMux.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		cart = []CartItem{}
		fmt.Fprint(w, "<h1>Payment Successful! Pickup in 20 mins.</h1><a href='/'>Go Back</a>")
	})

	go func() {
		fmt.Println("SEO Landing Page: http://localhost:9002")
		http.ListenAndServe(":9002", landingMux)
	}()

	fmt.Println("Ordering App: http://localhost:9001")
	log.Fatal(http.ListenAndServe(":9001", orderMux))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, nil)
}

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
		// Added id for scrolling and category-section class
		fmt.Fprintf(w, "<section id='%s' class='category-section'><h2 class='category-header'>%s</h2><div class='pizza-grid'>",
			cat, strings.ToUpper(cat))
		for _, p := range products {
			renderProductCard(w, p)
		}
		fmt.Fprintf(w, "</div></section>")
	}
}

func renderProductCard(w http.ResponseWriter, p Product) {
	optionsHTML := ""

	// Switch based on CUSTOMIZATION TYPE, not Category
	switch p.TypeTag {
	case "pizza_opt":
		optionsHTML = `
			<div class="options-group">
				<span class="option-title">Add-ons</span>
				<label class="chip"><input type="checkbox" name="extra_cheese"><span>🧀 Extra Cheese</span></label>
				<label class="chip"><input type="checkbox" name="extra_topping"><span>🍕 Extra Topping</span></label>
			</div>`
	case "coffee_opt":
		optionsHTML = `
			<!-- Temperature Section -->
			<div class="options-group">
				<span class="option-title">Temperature</span>
				<label class="chip">
					<input type="radio" name="temp" value="Ice" checked> 
					<span>❄️ Ice</span>
				</label>
				<label class="chip">
					<input type="radio" name="temp" value="Hot"> 
					<span>🔥 Hot</span>
				</label>
			</div>

			<!-- Sweetness Section (Updated) -->
			<div class="options-group">
				<span class="option-title">Sweetness Level</span>
				
				<!-- 1. Regular Sweet (Default) -->
				<label class="chip">
					<input type="radio" name="sweetness" value="Regular Sweet" checked> 
					<span>🍬 Regular</span>
				</label>

				<!-- 2. Less Sweet -->
				<label class="chip">
					<input type="radio" name="sweetness" value="Less Sweet"> 
					<span>🥄 Less</span>
				</label>

				<!-- 3. Least Sweet -->
				<label class="chip">
					<input type="radio" name="sweetness" value="Least Sweet"> 
					<span>🤏 Least</span>
				</label>
			</div>`
	case "pasta_opt":
		optionsHTML = `
			<div class="options-group">
				<span class="option-title">Customize</span>
				<label class="chip"><input type="checkbox" name="extra_pasta"><span>🍝 Extra Pasta</span></label>
			</div>`
	case "none":
		// Canned drinks, water, sides get no HTML here
		optionsHTML = ""
	}

	// Logic for Out Of Stock
	buttonHTML := `<button type="submit" class="btn-add">Add +</button>`
	opacityClass := ""

	if !p.InStock {
		buttonHTML = `<button type="button" class="btn-add btn-disabled" disabled>Sold Out</button>`
		opacityClass = "product-unavailable"
	}

	fmt.Fprintf(w, `
		<form hx-post="/cart/add?id=%d" hx-target="#cart-status" class="pizza-card %s">
			<img src="%s" alt="%s" loading="lazy">
			<div class="card-content">
				<h3>%s</h3>
				<p>%s</p>
				%s
				<div class="card-footer">
					<span class="price">RM%.2f</span>
					%s
				</div>
			</div>
		</form>`,
		p.ID, opacityClass, p.ImageURL, p.Name, p.Name, p.Description, optionsHTML, p.Price, buttonHTML)
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
	if r.FormValue("special_cheese") == "on" {
		item.Options = append(item.Options, "Special Cheese")
		item.AddonTotal += 4.0
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
	subtotal := 0.0
	for _, item := range cart {
		subtotal += item.Total()
	}
	tax := subtotal * taxRate

	if len(cart) == 0 {
		fmt.Fprint(w, `<div id="cart-status"><p class="empty-msg">Your cart is empty.</p></div>`)
		return
	}

	fmt.Fprint(w, `<div id="cart-status" class="cart-content-wrapper">`)
	fmt.Fprint(w, `<ul class="cart-items">`)
	for _, item := range cart {
		opts := ""
		if len(item.Options) > 0 {
			opts = fmt.Sprintf("<br><small style='color:#777'>%s</small>", strings.Join(item.Options, ", "))
		}
		fmt.Fprintf(w, "<li><div><strong>%s</strong>%s</div> <span>RM%.2f</span></li>", item.Name, opts, item.Total())
	}
	fmt.Fprintf(w, `</ul>

		<div class="totals">
			<div style="display:flex; justify-content:space-between; font-size: 0.9rem; color: #666;">
				<span>Subtotal</span><span>RM%.2f</span>
			</div>
			<div style="display:flex; justify-content:space-between; font-size: 0.9rem; color: #666; margin: 5px 0 10px 0;">
				<span>Tax (5%%)</span><span>RM%.2f</span>
			</div>
			<div class="grand-total">
				<span>Total</span><span>RM%.2f</span>
			</div>
		</div>

		<div class="cart-actions" style="display: flex; flex-direction: column; gap: 10px; margin-top: 20px;">
			<button hx-post="/checkout" class="btn-add" style="width: 100%; padding: 15px; font-size: 1.1rem;">
				Checkout & Pay
			</button>
			<button hx-post="/cart/clear" hx-target="#cart-status" 
				style="background:none; border:none; color:#999; cursor:pointer; font-size:0.8rem; text-decoration:underline;">
				Clear Order
			</button>
		</div>
	</div>`, subtotal, tax, subtotal+tax)
}

func handleCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	if len(cart) == 0 {
		return
	}
	sc := stripe.NewClient(stripeSecretKey)
	var lineItems []*stripe.CheckoutSessionCreateLineItemParams

	for _, item := range cart {
		lineItems = append(lineItems, &stripe.CheckoutSessionCreateLineItemParams{
			PriceData: &stripe.CheckoutSessionCreateLineItemPriceDataParams{
				Currency: stripe.String("myr"),
				ProductData: &stripe.CheckoutSessionCreateLineItemPriceDataProductDataParams{
					Name: stripe.String(item.Name + " (" + strings.Join(item.Options, ", ") + ")"),
				},
				UnitAmount: stripe.Int64(int64(item.Total() * 100)),
			},
			Quantity: stripe.Int64(1),
		})
	}

	params := &stripe.CheckoutSessionCreateParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems:          lineItems,
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:         stripe.String("http://localhost:9001/success"),
		CancelURL:          stripe.String("http://localhost:9001/"),
	}

	s, _ := sc.V1CheckoutSessions.Create(r.Context(), params)
	w.Header().Set("HX-Redirect", s.URL)
	w.WriteHeader(http.StatusSeeOther)
}

// TEMP to test order without stripe payement Replace the old Stripe handler with this:
func handleCheckout(w http.ResponseWriter, r *http.Request) {
	if len(cart) == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// 1. Calculate Total
	total := 0.0
	for _, item := range cart {
		total += item.Total()
	}
	totalWithTax := total + (total * taxRate)

	// 2. Insert Order
	// In a real app, you'd get the name from a form input
	res, err := db.Exec("INSERT INTO orders (customer_name, total_amount, status) VALUES (?, ?, ?)", "Guest Customer", totalWithTax, "Paid")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	orderID, _ := res.LastInsertId()

	// 3. Insert Items
	for _, item := range cart {
		opts := strings.Join(item.Options, ", ")
		_, err = db.Exec("INSERT INTO order_items (order_id, product_name, options, price) VALUES (?, ?, ?, ?)",
			orderID, item.Name, opts, item.Total())
		if err != nil {
			log.Printf("Error saving item: %v", err)
		}
	}

	// 4. Clear Cart
	cart = []CartItem{}

	// 5. Redirect to Success
	http.Redirect(w, r, "/success", http.StatusSeeOther)
}
