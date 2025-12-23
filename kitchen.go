package main

import (
	"fmt"
	"net/http"
)

type Order struct {
	ID        int
	Customer  string
	Total     float64
	Status    string
	CreatedAt string
	Items     []OrderItem
}

type OrderItem struct {
	Name    string
	Options string
	Price   float64
}

// 1. Render the Kitchen Page Skeleton
func handleKitchenPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Kitchen Display System</title>
    <link rel="stylesheet" href="/static/styles.css">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        body { background-color: #222; color: #fff; }
        .kitchen-grid { 
            display: grid; 
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); 
            gap: 1rem; 
            padding: 1rem; 
        }
        .ticket { 
            background: #333; 
            border-top: 5px solid #666; 
            border-radius: 4px; 
            padding: 1rem; 
            display: flex; 
            flex-direction: column;
            animation: fadeIn 0.5s;
        }
        @keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

        .status-Paid { border-top-color: #e31837; }      
        .status-Cooking { border-top-color: #f39c12; }   
        .status-Ready { border-top-color: #27ae60; }     
        
        .ticket-header { display: flex; justify-content: space-between; border-bottom: 1px solid #444; padding-bottom: 8px; margin-bottom: 8px;}
        .ticket-items { list-style: none; padding: 0; margin: 0; flex-grow: 1;}
        .ticket-items li { margin-bottom: 8px; font-size: 1.1rem; }
        .ticket-opt { display: block; font-size: 0.85rem; color: #aaa; }
        
        .action-area { margin-top: 1rem; display: grid; gap: 5px; }
        .btn-kds { width: 100%; padding: 12px; font-size: 1rem; font-weight: bold; cursor: pointer; border: none; border-radius: 4px;}
    </style>
</head>
<body>
    <header style="background: #111; padding: 10px 20px; display: flex; justify-content: space-between; align-items: center;">
        <h1 style="margin:0;">🔥 KITCHEN DISPLAY</h1>
        <div style="font-size: 0.8rem; color: #666;">
            System Active • <span id="sound-status" style="cursor:pointer;" onclick="testSound()">🔊 Test Sound</span>
        </div>
        <a href="/" style="color: #999; text-decoration: none;">Exit</a>
    </header>

    <!-- Audio Element -->
    <audio id="alert-sound" src="/images/alert.mp3" preload="auto"></audio>

    <!-- The Container that Polls -->
    <div class="kitchen-grid" 
         hx-get="/kitchen/orders" 
         hx-trigger="load, every 5s">
    </div>

    <script>
        // Track orders we have already seen to avoid repeating sound for existing orders
        let seenOrders = new Set();
        let isFirstLoad = true;

        function playSound() {
            const audio = document.getElementById('alert-sound');
            audio.currentTime = 0;
            audio.play().catch(e => {
                console.log("Audio play failed (browser interaction required):", e);
                document.getElementById('sound-status').innerText = "🔇 Click here to enable sound";
                document.getElementById('sound-status').style.color = "red";
            });
        }

        // Exposed for the "Test Sound" button
        window.testSound = function() {
            playSound();
            document.getElementById('sound-status').innerText = "🔊 Sound Active";
            document.getElementById('sound-status').style.color = "#666";
        }

        // Hook into HTMX to check for new orders after every poll
        document.body.addEventListener('htmx:afterOnLoad', function(evt) {
            // Only care if it's the orders list update
            if (evt.target.classList.contains('kitchen-grid')) {
                const tickets = document.querySelectorAll('.ticket');
                let hasNewOrder = false;

                tickets.forEach(t => {
                    const id = t.getAttribute('data-id');
                    if (!seenOrders.has(id)) {
                        seenOrders.add(id);
                        hasNewOrder = true;
                    }
                });

                // Play sound if new order found (skip sound on very first page load to avoid noise explosion)
                if (hasNewOrder && !isFirstLoad) {
                    playSound();
                }
                
                isFirstLoad = false;
            }
        });
    </script>
</body>
</html>`)
}

// 2. Fetch Orders (The Polling Endpoint)
func handleGetKitchenOrders(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, customer_name, total_amount, status, created_at FROM orders WHERE status != 'Completed' ORDER BY id ASC")
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		rows.Scan(&o.ID, &o.Customer, &o.Total, &o.Status, &o.CreatedAt)

		itemRows, _ := db.Query("SELECT product_name, options, price FROM order_items WHERE order_id = ?", o.ID)
		for itemRows.Next() {
			var i OrderItem
			itemRows.Scan(&i.Name, &i.Options, &i.Price)
			o.Items = append(o.Items, i)
		}
		itemRows.Close()

		orders = append(orders, o)
	}

	if len(orders) == 0 {
		fmt.Fprint(w, `<div style="grid-column: 1/-1; text-align: center; color: #555; margin-top: 50px;">
			<h2>No active orders</h2><p>Waiting for customers...</p>
		</div>`)
		return
	}

	for _, o := range orders {
		renderTicket(w, o)
	}
}

// 3. Helper to render a single ticket
func renderTicket(w http.ResponseWriter, o Order) {
	btnText := "Start Cooking"
	nextStatus := "Cooking"
	btnColor := "#e31837" // Red

	if o.Status == "Cooking" {
		btnText = "Mark Ready"
		nextStatus = "Ready"
		btnColor = "#f39c12" // Orange
	} else if o.Status == "Ready" {
		btnText = "Complete Order"
		nextStatus = "Completed"
		btnColor = "#27ae60" // Green
	}

	// Added data-id attribute here for the JS to track
	fmt.Fprintf(w, `
	<div class="ticket status-%s" id="order-%d" data-id="%d">
		<div class="ticket-header">
			<span style="font-weight:bold; font-size:1.2rem;">#%d</span>
			<span>%s</span>
		</div>
		<div style="font-size: 0.8rem; color: #888; margin-bottom: 10px;">%s</div>
		
		<ul class="ticket-items">`, o.Status, o.ID, o.ID, o.ID, o.Customer, o.CreatedAt)

	for _, item := range o.Items {
		opts := ""
		if item.Options != "" {
			opts = fmt.Sprintf(`<span class="ticket-opt">+ %s</span>`, item.Options)
		}
		fmt.Fprintf(w, `<li>%s %s</li>`, item.Name, opts)
	}

	fmt.Fprintf(w, `
		</ul>
		<div class="action-area">
			<button class="btn-kds" 
				style="background-color: %s; color: white;"
				hx-post="/kitchen/status?id=%d&status=%s"
				hx-target="#order-%d"
				hx-swap="outerHTML">
				%s
			</button>
		</div>
	</div>`, btnColor, o.ID, nextStatus, o.ID, btnText)
}

// 4. Update Status Handler (Same as before)
func handleKitchenStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	newStatus := r.URL.Query().Get("status")

	_, err := db.Exec("UPDATE orders SET status = ? WHERE id = ?", newStatus, id)
	if err != nil {
		fmt.Printf("Error updating: %v", err)
		return
	}

	if newStatus == "Completed" {
		return
	}

	var o Order
	db.QueryRow("SELECT id, customer_name, total_amount, status, created_at FROM orders WHERE id = ?", id).
		Scan(&o.ID, &o.Customer, &o.Total, &o.Status, &o.CreatedAt)

	itemRows, _ := db.Query("SELECT product_name, options, price FROM order_items WHERE order_id = ?", id)
	defer itemRows.Close()
	for itemRows.Next() {
		var i OrderItem
		itemRows.Scan(&i.Name, &i.Options, &i.Price)
		o.Items = append(o.Items, i)
	}

	renderTicket(w, o)
}
