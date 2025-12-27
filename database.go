package main

import (
	"database/sql"
	"log"
)

func initDB(db *sql.DB) {

	// 2. Clear old data
	db.Exec("DELETE FROM products")

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
		status TEXT DEFAULT 'Paid', -- Paid, Cooking, Ready, Completed
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
		db.Exec(`INSERT INTO categories (name) VALUES ('pizza'), ('pasta'), ('drink')`)
	}
	if err != nil {
		log.Fatal(err)
	}

	// 3. Helper function to make inserting cleaner
	insert := func(cat, name, desc string, price float64, img, tag string, stock bool) {
		_, err := db.Exec(
			`INSERT INTO products (category, name, description, price, image_url, type_tag, in_stock) 
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			cat, name, desc, price, img, tag, stock,
		)
		if err != nil {
			log.Printf("Error inserting %s: %v", name, err)
		}
	}

	// --- 1. PIZZAS (Tag: 'pizza_opt') ---
	// These will show: Extra Cheese, Extra Topping buttons
	insert("pizza", "Fresh Buffalo Mozzarella", "Red sauce, fresh buffalo mozzarella, olive oil", 39.00, "./images/fresh_buffalo_mozzarella.webp", "pizza_opt", true)
	insert("pizza", "Portobello Beef Mushroom", "Portobello mushroom, beef brisket, truffle oil", 39.00, "./images/portobello_beef_mushroom.webp", "pizza_opt", true)
	insert("pizza", "Carbonara Pizza", "Streaky beef, egg, parmesan, fresh basils", 37.00, "./images/carbonara_pizza.webp", "pizza_opt", true)
	insert("pizza", "Supreme Beef", "Minced beef, salami, beef pepperoni, jalapenos", 35.00, "./images/supreme_beef.webp", "pizza_opt", true)
	insert("pizza", "Beefy Mushroom", "Minced beef, pastrami beef, mushroom slices", 35.00, "./images/beefy_mushroom.webp", "pizza_opt", true)
	insert("pizza", "Roast Beef", "Roast beef slices, jalapenos, chilli oil", 35.00, "./images/roast_beef.webp", "pizza_opt", true)
	insert("pizza", "Honey Drizzled Pepperoni", "Pepperoni, honey, chilli oil, parmesan", 33.00, "./images/honey_drizzled_pepperoni.webp", "pizza_opt", true)
	insert("pizza", "Spicy Prawn", "Marinated prawns, pineapples, cherry tomatoes", 33.00, "./images/spicy_prawn.webp", "pizza_opt", true)
	insert("pizza", "Beef Sambal Hitam", "Minced beef, sambal hitam Pahang, jalapenos", 32.00, "./images/beef_sambal_hitam.webp", "pizza_opt", true)
	insert("pizza", "Smoked Duck", "Smoked duck, pineapples, cherry tomatoes", 32.00, "./images/smoked_duck.webp", "pizza_opt", true)
	insert("pizza", "5 Cheese", "Mozzarella, cheddar, feta, cream cheese, blue cheese", 30.00, "./images/5_cheese.webp", "pizza_opt", true)
	insert("pizza", "Beef Pepperoni", "Beef pepperoni, olives, mozzarella", 30.00, "./images/beef_pepperoni.webp", "pizza_opt", true)
	insert("pizza", "Chicken Pepperoni", "Chicken pepperoni, olives, mozzarella", 30.00, "./images/chicken_pepperoni.webp", "pizza_opt", true)
	insert("pizza", "Hawaiian Chicken", "Smoked chicken, pineapples, BBQ sauce", 30.00, "./images/hawaiian_chicken.webp", "pizza_opt", true)
	insert("pizza", "Margherita", "Mozzarella, cheddar, cherry tomatoes", 26.00, "./images/margherita.webp", "pizza_opt", true)
	insert("pizza", "Marshmallow Nutella", "Marshmallows, nutella, chocolate syrup", 28.00, "./images/marshmallow_nutella.webp", "pizza_opt", true)
	insert("pizza", "Banana Nutella", "Nutella spread, banana slices", 25.00, "./images/banana_nutella.webp", "pizza_opt", true)
	insert("pizza", "Durian Pizza", "Durian IOI/D24, mozzarella", 40.00, "./images/durian_pizza.webp", "pizza_opt", true)

	// --- 2. PASTAS (Tag: 'pasta_opt') ---
	// These will show: Extra Pasta, Extra Topping buttons
	insert("pasta", "Carbonara Samyang Chicken", "Spicy Samyang & cream", 15.50, "", "pasta_opt", true)
	insert("pasta", "Carbonara Samyang Beef", "Spicy Samyang & cream", 15.50, "", "pasta_opt", true)
	insert("pasta", "Carbonara Samyang Prawn", "Spicy Samyang & cream", 17.50, "", "pasta_opt", true)
	insert("pasta", "Carbonara Tomyam Chicken", "Spicy Thai fusion", 15.50, "", "pasta_opt", true)
	insert("pasta", "Carbonara Tomyam Beef", "Spicy Thai fusion", 15.50, "", "pasta_opt", true)
	insert("pasta", "Carbonara Tomyam Prawn", "Spicy Thai fusion", 17.50, "", "pasta_opt", true)
	insert("pasta", "Carbonara Salted Egg Chicken", "Rich salted egg yolk sauce", 15.50, "", "pasta_opt", true)
	insert("pasta", "Carbonara Salted Egg Beef", "Rich salted egg yolk sauce", 15.50, "", "pasta_opt", true)
	insert("pasta", "Carbonara Salted Egg Prawn", "Rich salted egg yolk sauce", 17.50, "", "pasta_opt", true)
	insert("pasta", "Carbonara Original Chicken", "Creamy classic Italian", 13.50, "", "pasta_opt", true)
	insert("pasta", "Carbonara Original Beef", "Creamy classic Italian", 13.50, "", "pasta_opt", true)
	insert("pasta", "Bolognese Chicken", "Classic tomato meat sauce", 13.50, "", "pasta_opt", true)
	insert("pasta", "Bolognese Beef", "Classic tomato meat sauce", 13.50, "", "pasta_opt", true)
	insert("pasta", "Aglio Olio Chicken", "Garlic, olive oil, chilli flakes", 13.50, "./images/aglio_olio_chicken.webp", "pasta_opt", true)
	insert("pasta", "Aglio Olio Beef", "Garlic, olive oil, chilli flakes", 13.50, "", "pasta_opt", true)
	insert("pasta", "Aglio Olio Prawn", "Garlic, olive oil, succulent prawns", 15.50, "./images/aglio_olio_prawn.webp", "pasta_opt", true)

	// --- 3. CUSTOMIZABLE DRINKS (Tag: 'coffee_opt') ---
	// These will show: Hot/Ice and Sweetness selectors
	insert("drink", "Cafe Latte", "Fresh Espresso", 10.00, "./images/cafe_latte.webp", "coffee_opt", true)
	insert("drink", "Chocolate", "Rich Cocoa", 10.00, "./images/chocolate.webp", "coffee_opt", true)
	insert("drink", "Matcha Latte", "Premium Matcha", 10.00, "./images/matcha.webp", "coffee_opt", true)
	insert("drink", "Americano", "Black Coffee", 8.00, "./images/americano.webp", "coffee_opt", true)
	insert("drink", "Rose Latte", "Floral infusion", 11.00, "./images/rose_latte.webp", "coffee_opt", true)

	// --- 4. FIXED DRINKS (Tag: 'none') ---
	// These will show NO selectors (served as is)
	insert("drink", "Ice Lemon Tea", "Chilled refreshing tea", 5.00, "./images/ice_lemon_tea.webp", "none", true)
	insert("drink", "Sirap Bandung", "Rose milk syrup", 6.00, "./images/sirap.webp", "none", true)
	insert("drink", "Can Drinks", "Coke, Sprite, etc.", 4.00, "./images/can_drinks.webp", "none", true)
	insert("drink", "Mineral Water", "Bottled water", 2.00, "./images/mineral_water.webp", "none", true)

	// --- 5. SIDES (Tag: 'none') ---
	// Currently no options for sides
	insert("sides", "Roasted Chicken Wings", "4 pieces baked", 18.00, "./images/roasted_chicken_wings.webp", "none", true)
	insert("sides", "Baked Portobello", "Beef brisket & cheese", 8.00, "./images/baked_portobello_mushroom_1_piece.webp", "none", true)
}
