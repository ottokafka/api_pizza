func initDB() {
	statement := `
	CREATE TABLE IF NOT EXISTS pizzas (id INTEGER PRIMARY KEY, name TEXT, description TEXT, price REAL, image_url TEXT);
	DELETE FROM pizzas;
	INSERT INTO pizzas (name, description, price, image_url) VALUES 
	('Fresh Buffalo Mozzarella', 'Red sauce, fresh buffalo mozzarella, olive oil, parmesan, fresh basils', 39.00, './images/fresh_buffalo_mozzarella.webp'),
	('Portobello Beef Mushroom', 'Creamy white sauce, mozzarella, cheddar, portobello mushroom, beef brisket, truffle oil, fresh basils', 39.00, './images/portobello_beef_mushroom.webp'),
	('Carbonara Pizza', 'Creamy white sauce, mozzarella, cheddar, streaky beef, egg, parmesan, fresh basils', 37.00, './images/carbonara_pizza.webp'),
	('Supreme Beef', 'Minced beef, salami beef, beef pepperoni, mozzarella, cheddar, jalapenos, white sauce & fresh basils', 35.00, './images/supreme_beef.webp'),
	('Beefy Mushroom', 'Minced beef, pastrami beef, mushroom slices, red onions, white sauce, mozzarella, cheddar, grated parmesan and fresh basils', 35.00, './images/beefy_mushroom.webp'),
	('Roast Beef', 'Roast beef slices, jalapenos, mozzarella, cheddar, grated parmesan, chilli oil, fresh basils', 35.00, './images/roast_beef.webp'),
	('Honey Drizzled Pepperoni', 'White sauce, pepperoni, mozzarella, cheddar, honey, chilli oil, parmesan and fresh basils', 33.00, './images/honey_drizzled_pepperoni.webp'),
	('Spicy Prawn', 'Marinated prawns, pineapples, cherry tomatoes, mozzarella, cheddar, red sauce & fresh basils', 33.00, './images/spicy_prawn.webp'),
	('Beef Sambal Hitam', 'White sauce, mozzarella, cheddar, minced beef, sambal hitam Pahang, jalapenos, parmesan and fresh basils', 32.00, './images/beef_sambal_hitam.webp'),
	('Smoked Duck', 'Smoked duck, pineapples, cherry tomatoes, mozzarella, cheddar, red sauce & fresh basils', 32.00, './images/smoked_duck.webp'),
	('5 Cheese', 'Mozzarella, cheddar, feta, cream cheese, blue cheese, white sauce & fresh basils', 30.00, './images/5_cheese.webp'),
	('Beef Pepperoni', 'Beef pepperoni, olives, mozzarella, cheddar, red sauce & fresh basils', 30.00, './images/beef_pepperoni.webp'),
	('Chicken Pepperoni', 'Chicken pepperoni, olives, mozzarella, cheddar, red sauce & fresh basils', 30.00, './images/chicken_pepperoni.webp'),
	('Hawaiian Chicken', 'Smoked chicken, pineapples, mozzarella, cheddar, red sauce, bbq sauce, fresh basils', 30.00, './images/hawaiian_chicken.webp'),
	('Margherita', 'Mozzarella, cheddar, cherry tomatoes, red sauce, & fresh basils', 26.00, './images/margherita.webp'),
	('Marshmallow Nutella', 'Marshmallows, nutella, Hershey’s chocolate syrup', 28.00, './images/marshmallow_nutella.webp'),
	('Banana Nutella', 'Nutella spread, banana slices, chocolate syrup, sugar powder', 25.00, './images/banana_nutella.webp'),
	('Durian Pizza', 'Seasonal pizza. Ingredients: durian IOI/D24, mozzarella.', 40.00, './images/durian_pizza.webp'),
	('Carbonara Samyang Chicken', '', 15.50, ''),
	('Carbonara Samyang Beef', '', 15.50, ''),
	('Carbonara Samyang Prawn', '', 17.50, ''),
	('Carbonara Tomyam Chicken', '', 15.50, ''),
	('Carbonara Tomyam Beef', '', 15.50, ''),
	('Carbonara Tomyam Prawn', '', 17.50, ''),
	('Carbonara Salted Egg Chicken', '', 15.50, ''),
	('Carbonara Salted Egg Beef', '', 15.50, ''),
	('Carbonara Salted Egg Prawn', '', 17.50, ''),
	('Carbonara Original Chicken', '', 13.50, ''),
	('Carbonara Original Beef', '', 13.50, ''),
	('Carbonara Original Prawn', '', 15.50, ''),
	('Bolognese Chicken', '', 13.50, ''),
	('Bolognese Beef', '', 13.50, ''),
	('Aglio Olio Chicken', 'Classic garlic and oil pasta with roasted chicken.', 13.50, './images/aglio_olio_chicken.webp'),
	('Aglio Olio Beef', '', 13.50, ''),
	('Aglio Olio Prawn', 'Classic garlic and oil pasta with succulent prawns.', 15.50, './images/aglio_olio_prawn.webp'),
	('Roasted Chicken Wings', 'Oven baked chicken wings - 4 pieces', 18.00, './images/roasted_chicken_wings.webp'),
	('Baked Portobello Mushroom (1 Piece)', 'Portobello mushroom, beef brisket, mozzarella, cheddar, truffle oil', 8.00, './images/baked_portobello_mushroom_1_piece.webp'),
	('2 x Buldak 6g Stick Sauce [Not For Sale]', 'Get 2 x Buldak 6g Stick Sauce with purchases above RM20!', 4.40, '');
	`
	_, err := db.Exec(statement)
	if err != nil {
		log.Fatal(err)
	}
	drinks_initDB()
}

func drinks_initDB() {
	statement := `
	CREATE TABLE IF NOT EXISTS drinks (id INTEGER PRIMARY KEY, name TEXT, description TEXT, price REAL, image_url TEXT);
	DELETE FROM drinks;
	INSERT INTO drinks (name, description, price, image_url) VALUES 
	('Cafe Latte', 'Classic hot cafe latte with steamed milk', 10.00, './images/cafe_latte.webp'),
	('Iced Cafe Latte', 'Chilled cafe latte served over ice', 11.00, './images/cafe_latte.webp'),
	('Chocolate', 'Rich hot chocolate drink', 10.00, './images/chocolate.webp'),
	('Iced Chocolate', 'Creamy chilled chocolate drink', 11.00, './images/chocolate.webp'),
	('Matcha', 'Frothy matcha green tea latte', 10.00, './images/matcha.webp'),
	('Iced Matcha', 'Refreshing chilled matcha green tea', 11.00, './images/matcha.webp'),
	('Americano', 'Bold black Americano coffee', 8.00, './images/americano.webp'),
	('Iced Americano', 'Bold black coffee served over ice', 9.00, './images/americano.webp'),
	('Rose Latte', 'Floral rose-infused latte', 11.00, './images/rose_latte.webp'),
	('Iced Rose Latte', 'Chilled floral rose-infused latte', 12.00, './images/rose_latte.webp'),
	('Ice Lemon Tea', 'Refreshing iced lemon tea', 5.00, './images/ice_lemon_tea.webp'),
	('Sirap', 'Sweet rose syrup drink', 5.00, './images/sirap.webp'),
	('Sirap Bandung', 'Sweet rose milk syrup drink', 6.00, './images/sirap.webp'),
	('Can Drinks', 'Assorted canned soft drinks (Coke, Sprite, etc.)', 4.00, './images/can_drinks.webp'),
	('Mineral Water', 'Bottled mineral water', 2.00, './images/mineral_water.webp');
	`
	_, err := db.Exec(statement)
	if err != nil {
		log.Fatal(err)
	}
}
