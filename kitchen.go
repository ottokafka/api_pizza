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
        body { background-color: #222; color: #fff; font-family: sans-serif; margin: 0; }
        
        header {
            background: #111; 
            padding: 15px 20px; 
            display: flex; 
            justify-content: space-between; 
            align-items: center;
            border-bottom: 2px solid #444;
            position: sticky;
            top: 0;
            z-index: 100;
            height: 60px; /* Fixed height for calculation */
            box-sizing: border-box;
        }
        #system-clock {
            font-size: 2.5rem;
            font-weight: bold;
            color: #fff;
            font-family: monospace;
            background: #000;
            padding: 5px 15px;
            border-radius: 5px;
            border: 1px solid #444;
        }

        /* LAYOUT CONTAINERS */
        /* This ensures the active section takes up at least the full viewport height */
        /* forcing the completed section 'below the fold' */
        .active-wrapper {
            min-height: 95vh; 
            display: flex;
            flex-direction: column;
            padding: 1rem;
            box-sizing: border-box;
        }

        .kitchen-grid { 
            display: grid; 
            grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); 
            gap: 1rem; 
            width: 100%;
        }

        /* COMPLETED SECTION STYLES */
        .completed-section {
            background-color: #1a1a1a;
            border-top: 5px solid #333;
            padding: 2rem 1rem 4rem 1rem; /* Extra padding at bottom */
        }
        
        .completed-header {
            font-size: 1.2rem;
            color: #555;
            text-transform: uppercase;
            letter-spacing: 2px;
            margin-bottom: 2rem;
            text-align: center;
            border-bottom: 1px solid #333;
            line-height: 0.1em;
            margin: 10px 0 30px; 
        }
        .completed-header span { 
            background: #1a1a1a; 
            padding: 0 10px; 
        }

        /* Ticket Styles */
        .ticket { 
            background: #333; 
            border: 1px solid #555;
            border-radius: 6px; 
            display: flex; 
            flex-direction: column;
            animation: fadeIn 0.5s;
            box-shadow: 0 4px 6px rgba(0,0,0,0.3);
            height: fit-content;
        }
        .ticket.completed-ticket {
            opacity: 0.6;
            filter: grayscale(0.5);
            border-color: #444;
            background: #222;
        }
        .ticket.completed-ticket .ticket-header { background: #2a2a2a; color: #888; }

        @keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

        .ticket-header { 
            background: #444;
            padding: 10px 15px;
            border-bottom: 1px solid #555;
            display: flex; 
            justify-content: space-between;
            align-items: center;
            border-radius: 6px 6px 0 0;
        }
        .ticket-body { padding: 15px; flex-grow: 1; }
        .ticket-meta { font-size: 0.9rem; color: #aaa; margin-bottom: 10px; display: flex; justify-content: space-between; border-bottom: 1px dashed #555; padding-bottom: 8px; }
        .elapsed-time { color: #f39c12; font-weight: bold; font-family: monospace; font-size: 1.1rem; }
        .ticket-items { list-style: none; padding: 0; margin: 0; }
        .ticket-items li { margin-bottom: 8px; font-size: 1.2rem; font-weight: 500; }
        .ticket-opt { display: block; font-size: 0.9rem; color: #bbb; margin-left: 10px; font-style: italic;}
        
        .action-area { padding: 10px 15px; background: #2a2a2a; border-radius: 0 0 6px 6px; }
        .btn-kds { width: 100%; padding: 15px; font-size: 1.2rem; font-weight: bold; cursor: pointer; border: none; border-radius: 4px; text-transform: uppercase; }
        
        .btn-complete { background-color: #27ae60; color: white; }
        .btn-complete:hover { background-color: #219150; }
        
        .btn-restore { background-color: #444; color: #aaa; font-size: 1rem; padding: 10px;}
        .btn-restore:hover { background-color: #666; color: white; }

    </style>
</head>
<body>
    <header>
        <div>
            <h1 style="margin:0; font-size: 1.5rem;">🔥 KITCHEN DISPLAY</h1>
            <div style="font-size: 0.8rem; color: #888; margin-top:5px;">
                <span id="sound-status" style="cursor:pointer;" onclick="testSound()">🔊 Test Sound</span>
            </div>
        </div>
        <div id="system-clock">--:--:--</div>
        <a href="/" style="color: #999; text-decoration: none;">Exit</a>
    </header>
    <audio id="alert-sound" src="/images/alert.mp3" preload="auto"></audio>

    <!-- Main Content Wrapper -->
    <div id="kds-container"
         hx-get="/kitchen/orders" 
         hx-trigger="load, every 5s">
         <!-- Content injected here via HTMX -->
    </div>

    <script>
        let seenOrders = new Set();
        let isFirstLoad = true;

        function updateTime() {
            const now = new Date();
            document.getElementById('system-clock').innerText = now.toLocaleTimeString([], { hour12: true });

            document.querySelectorAll('.ticket').forEach(ticket => {
                if(ticket.classList.contains('completed-ticket')) return; // Stop timer for completed

                const createdStr = ticket.getAttribute('data-created');
                const timerElem = ticket.querySelector('.elapsed-time');
                
                if (createdStr && timerElem) {
                    const createdDate = new Date(createdStr.replace(" ", "T"));
                    const diffMs = now - createdDate;
                    if (!isNaN(diffMs)) {
                        const totalSeconds = Math.floor(diffMs / 1000);
                        const minutes = Math.floor(totalSeconds / 60);
                        const seconds = totalSeconds % 60;
                        timerElem.innerText = minutes.toString().padStart(2, '0') + ":" + seconds.toString().padStart(2, '0');
                        if(minutes >= 15) timerElem.style.color = "#e74c3c";
                    }
                }
            });
        }

        setInterval(updateTime, 1000);

        document.body.addEventListener('htmx:afterOnLoad', function(evt) {
            if (evt.target.id === 'kds-container') {
                updateTime();
                const tickets = document.querySelectorAll('.ticket:not(.completed-ticket)'); 
                let hasNewOrder = false;
                tickets.forEach(t => {
                    const id = t.getAttribute('data-id');
                    if (!seenOrders.has(id)) {
                        seenOrders.add(id);
                        hasNewOrder = true;
                    }
                });
                if (hasNewOrder && !isFirstLoad) playSound();
                isFirstLoad = false;
            }
        });

        function playSound() {
            const audio = document.getElementById('alert-sound');
            audio.currentTime = 0;
            audio.play().catch(e => console.log("Audio needed interaction"));
        }
        window.testSound = function() { playSound(); document.getElementById('sound-status').style.color = "#fff"; }
    </script>
</body>
</html>`)
}

// 2. Fetch Orders
func handleGetKitchenOrders(w http.ResponseWriter, r *http.Request) {
	// Query 1: Active Orders (Created within last 24 hours)
	// We use datetime('now', '-24 hours') for SQLite.
	activeQuery := `
		SELECT id, customer_name, total_amount, status, created_at 
		FROM orders 
		WHERE status != 'Completed' 
		AND created_at >= datetime('now', '-24 hours')
		ORDER BY id ASC
	`
	activeOrders := getOrdersByQuery(activeQuery)

	// Query 2: Recently Completed (Created within last 24 hours, Limit 4)
	completedQuery := `
		SELECT id, customer_name, total_amount, status, created_at 
		FROM orders 
		WHERE status = 'Completed' 
		AND created_at >= datetime('now', '-24 hours')
		ORDER BY id DESC 
		LIMIT 4
	`
	completedOrders := getOrdersByQuery(completedQuery)

	// --- RENDER ACTIVE SECTION ---
	// The .active-wrapper min-height: 95vh ensures this pushes the bottom section down
	fmt.Fprint(w, `<div class="active-wrapper">`)

	if len(activeOrders) == 0 {
		fmt.Fprint(w, `<div style="text-align: center; color: #555; margin-top: 100px;"><h2>All Caught Up!</h2><p>No active orders in the last 24 hours.</p></div>`)
	} else {
		fmt.Fprint(w, `<div class="kitchen-grid">`)
		for _, o := range activeOrders {
			renderTicket(w, o, false)
		}
		fmt.Fprint(w, `</div>`)
	}
	fmt.Fprint(w, `</div>`) // End active-wrapper

	// --- RENDER COMPLETED SECTION ---
	if len(completedOrders) > 0 {
		fmt.Fprint(w, `
		<div class="completed-section">
            <h2 class="completed-header"><span>Recently Completed (24h)</span></h2>
            <div class="kitchen-grid">`)
		for _, o := range completedOrders {
			renderTicket(w, o, true)
		}
		fmt.Fprint(w, `</div></div>`)
	}
}

// Helper to avoid code duplication
func getOrdersByQuery(query string) []Order {
	rows, err := db.Query(query)
	if err != nil {
		fmt.Println("DB Error:", err)
		return []Order{}
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
	return orders
}

// 3. Render Ticket
func renderTicket(w http.ResponseWriter, o Order, isCompleted bool) {
	cssClass := ""
	btnText := "Complete Order"
	targetStatus := "Completed"
	btnClass := "btn-complete"

	if isCompleted {
		cssClass = "completed-ticket"
		btnText = "↩ Restore"
		targetStatus = "Paid" // Returns to active stack
		btnClass = "btn-restore"
	}

	fmt.Fprintf(w, `
	<div class="ticket %s" id="order-%d" data-id="%d" data-created="%s">
		<div class="ticket-header">
			<span style="font-weight:bold; font-size:1.3rem;">#%d</span>
			<span style="font-weight:bold;">%s</span>
		</div>
		<div class="ticket-body">
			<div class="ticket-meta">
				<span>%s</span>
				<span>Wait: <span class="elapsed-time">--:--</span></span>
			</div>
			<ul class="ticket-items">`, cssClass, o.ID, o.ID, o.CreatedAt, o.ID, o.Customer, o.CreatedAt)

	for _, item := range o.Items {
		opts := ""
		if item.Options != "" {
			opts = fmt.Sprintf(`<span class="ticket-opt">+ %s</span>`, item.Options)
		}
		fmt.Fprintf(w, `<li>%s %s</li>`, item.Name, opts)
	}

	fmt.Fprintf(w, `
			</ul>
		</div>
		<div class="action-area">
			<button class="btn-kds %s" 
				hx-post="/kitchen/status?id=%d&status=%s"
				hx-target="#kds-container" 
				hx-swap="innerHTML">
				%s
			</button>
		</div>
	</div>`, btnClass, o.ID, targetStatus, btnText)
}

// 4. Status Handler
func handleKitchenStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	newStatus := r.URL.Query().Get("status")

	_, err := db.Exec("UPDATE orders SET status = ? WHERE id = ?", newStatus, id)
	if err != nil {
		fmt.Printf("Error updating: %v", err)
	}

	// Reload the entire board
	handleGetKitchenOrders(w, r)
}
