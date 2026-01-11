package main

import (
	"fmt"
	"net/http"
	"time"
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

// 1. Render the Kitchen Page Skeleton (Updated CSS, JS, and Header)
func handleKitchenPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Kitchen Display System</title>
    <link rel="stylesheet" href="/static/styles.css">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        /* --- CSS VARIABLES FOR SCALING --- */
        :root {
            --scale-factor: 1; /* Default Scale */
            --bg-color: #222;
            --card-bg: #333;
            --text-color: #fff;
        }

        body { 
            background-color: var(--bg-color); 
            color: var(--text-color); 
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; 
            margin: 0; 
            /* This allows the whole page content to scale based on the variable */
            font-size: calc(16px * var(--scale-factor)); 
        }
        
        header {
            background: #111; 
            padding: 10px 20px; 
            display: flex; 
            justify-content: space-between; 
            align-items: center;
            border-bottom: 2px solid #444;
            position: sticky;
            top: 0;
            z-index: 100;
            height: 80px; 
            box-sizing: border-box;
        }

        /* --- CONTROL GROUP (Sound + Zoom) --- */
        .controls { display: flex; gap: 10px; align-items: center; }

        .icon-btn {
            font-size: 1.5rem;
            cursor: pointer;
            width: 50px;
            height: 50px;
            border: 2px solid #444;
            border-radius: 8px;
            color: #888;
            background: #000;
            display: flex;
            align-items: center;
            justify-content: center;
            user-select: none;
            transition: all 0.1s;
        }
        .icon-btn:hover { border-color: #666; color: #fff; background: #222; }
        .icon-btn:active { transform: scale(0.95); }
        
        .icon-btn.active-sound {
            color: #2ecc71;
            border-color: #2ecc71;
            box-shadow: 0 0 8px rgba(46, 204, 113, 0.2);
        }

        #zoom-level-display { font-family: monospace; font-size: 1.2rem; min-width: 60px; text-align: center; color: #aaa; }

        #system-clock {
            font-size: 2rem;
            font-weight: bold;
            font-family: monospace;
            background: #000;
            padding: 5px 15px;
            border-radius: 5px;
            border: 1px solid #444;
        }

        /* LAYOUT */
        .active-wrapper {
            min-height: 95vh; 
            display: flex;
            flex-direction: column;
            padding: 1rem;
            box-sizing: border-box;
        }

        .kitchen-grid { 
            display: grid; 
            /* Grid items grow automatically based on content size (zoom) */
            grid-template-columns: repeat(auto-fill, minmax(calc(320px * var(--scale-factor)), 1fr)); 
            gap: 1rem; 
            width: 100%;
        }

        /* TICKET STYLES */
        .ticket { 
            background: var(--card-bg); 
            border: 1px solid #555;
            border-radius: 8px; 
            display: flex; 
            flex-direction: column;
            animation: fadeIn 0.4s;
            box-shadow: 0 4px 8px rgba(0,0,0,0.4);
            overflow: hidden; /* Contains corners */
        }
        .ticket.completed-ticket { opacity: 0.6; filter: grayscale(0.6); }

        .ticket-header { 
            background: #444;
            padding: 0.8rem;
            border-bottom: 1px solid #555;
            display: flex; 
            justify-content: space-between;
            align-items: flex-start; /* Align top if name is long */
        }

        /* UPDATED: ID is huge, Name is smaller as requested */
        .header-left { display: flex; flex-direction: column; }
        .order-id { font-size: 2.5rem; font-weight: 800; line-height: 1; color: #fff; }
        .customer-name { font-size: 0.9rem; color: #bbb; margin-top: 4px; font-weight: normal; }

        .ticket-meta { 
            font-size: 0.9rem; 
            color: #aaa; 
            margin-bottom: 10px; 
            display: flex; 
            justify-content: space-between; 
            border-bottom: 1px dashed #555; 
            padding: 0.5rem 1rem; 
        }

        .elapsed-time { color: #f39c12; font-weight: bold; font-family: monospace; font-size: 1.1rem; }

        .ticket-body { padding: 0.5rem 1rem 1rem 1rem; flex-grow: 1; }
        
        /* UPDATED: Main Items larger */
        .ticket-items { list-style: none; padding: 0; margin: 0; }
        .ticket-items li { 
            margin-bottom: 12px; 
            font-size: 1.5rem; /* Increased size */
            font-weight: 600; 
            line-height: 1.2;
            color: #fff;
        }
        
        .ticket-opt { 
            display: block; 
            font-size: 0.95rem; 
            color: #3498db; /* Blue hint for mods */
            margin-left: 5px; 
            font-weight: normal;
            font-style: italic;
            margin-top: 2px;
        }
        
        .action-area { padding: 0; }
        
        /* UPDATED: Complete Button */
        .btn-kds { 
            width: 100%; 
            min-height: 60px; /* Fixed minimum large target */
            padding: 10px; 
            font-size: 1.3rem; 
            font-weight: bold; 
            cursor: pointer; 
            border: none; 
            text-transform: uppercase; 
            transition: background 0.2s;
        }
        
        .btn-complete { background-color: #27ae60; color: white; }
        .btn-complete:hover { background-color: #219150; }
        
        .btn-restore { background-color: #555; color: #ccc; }
        .btn-restore:hover { background-color: #777; color: white; }

        /* Animation */
        @keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
        
        /* Completed Section */
        .completed-section { margin-top: 2rem; border-top: 4px solid #333; padding-top: 1rem; background: #1a1a1a; padding: 2rem;}
        .completed-header { text-align: center; color: #555; text-transform: uppercase; letter-spacing: 2px; margin-bottom: 2rem; }

    </style>
</head>
<body>
    <header>
        <div class="controls">
            <!-- SOUND TOGGLE -->
            <div id="sound-toggle" class="icon-btn" onclick="testSound()" title="Enable/Test Sound">🔇</div>
            
            <!-- ZOOM CONTROLS -->
            <div class="icon-btn" onclick="adjustZoom(-0.1)">-</div>
            <div id="zoom-level-display">100%</div>
            <div class="icon-btn" onclick="adjustZoom(0.1)">+</div>
        </div>

        <div style="font-size: 1.5rem; font-weight:bold; color: #444;">KDS</div>

        <div class="controls">
            <div id="system-clock">--:--:--</div>
            <a href="/" class="icon-btn" style="text-decoration:none; font-size:1rem; width: auto; padding: 0 15px;">Exit</a>
        </div>
    </header>
    <audio id="alert-sound" src="/images/alert.mp3" preload="auto"></audio>

    <div id="kds-container"
         hx-get="/kitchen/orders" 
         hx-trigger="load, every 5s">
    </div>

    <script>
        let seenOrders = new Set();
        let isFirstLoad = true;
        let soundEnabled = false;

        // --- ZOOM LOGIC ---
        let currentScale = parseFloat(localStorage.getItem('kds_scale')) || 1.0;

        function applyZoom() {
            // Apply CSS Variable
            document.documentElement.style.setProperty('--scale-factor', currentScale);
            // Update Display Text
            document.getElementById('zoom-level-display').innerText = Math.round(currentScale * 100) + "%";
            // Save preference
            localStorage.setItem('kds_scale', currentScale);
        }

        function adjustZoom(delta) {
            currentScale += delta;
            if(currentScale < 0.5) currentScale = 0.5; // Min limit
            if(currentScale > 2.0) currentScale = 2.0; // Max limit
            applyZoom();
        }

        // Initialize Zoom on Load
        applyZoom();
        // ------------------

        function updateTime() {
            const now = new Date();
            document.getElementById('system-clock').innerText = now.toLocaleTimeString([], { hour12: true });

            document.querySelectorAll('.ticket').forEach(ticket => {
                if(ticket.classList.contains('completed-ticket')) return;
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

        // Sound Logic
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
                if (hasNewOrder && !isFirstLoad && soundEnabled) playSound();
                isFirstLoad = false;
            }
        });

        function playSound() {
            const audio = document.getElementById('alert-sound');
            audio.currentTime = 0;
            audio.play().catch(e => console.log("Audio needed interaction"));
        }

        window.testSound = function() { 
            playSound(); 
            soundEnabled = true;
            const btn = document.getElementById('sound-toggle');
            btn.classList.add('active-sound');
            btn.innerText = "🔊"; 
        }
    </script>
</body>
</html>`)
}

// 2. Fetch Orders (No logic changes, just layout structure calls)
func handleGetKitchenOrders(w http.ResponseWriter, r *http.Request) {
	// (Database queries remain identical to your previous code)
	activeQuery := `SELECT id, customer_name, total_amount, status, created_at FROM orders WHERE status != 'Completed' AND created_at >= datetime('now', '-24 hours') ORDER BY id ASC`
	activeOrders := getOrdersByQuery(activeQuery)

	completedQuery := `SELECT id, customer_name, total_amount, status, created_at FROM orders WHERE status = 'Completed' AND created_at >= datetime('now', '-24 hours') ORDER BY id DESC LIMIT 4`
	completedOrders := getOrdersByQuery(completedQuery)

	fmt.Fprint(w, `<div class="active-wrapper">`)
	if len(activeOrders) == 0 {
		fmt.Fprint(w, `<div style="text-align: center; color: #555; margin-top: 100px;"><h2>All Caught Up!</h2></div>`)
	} else {
		fmt.Fprint(w, `<div class="kitchen-grid">`)
		for _, o := range activeOrders {
			renderTicket(w, o, false)
		}
		fmt.Fprint(w, `</div>`)
	}
	fmt.Fprint(w, `</div>`)

	if len(completedOrders) > 0 {
		fmt.Fprint(w, `<div class="completed-section"><h2 class="completed-header">Recently Completed</h2><div class="kitchen-grid">`)
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

	// Date formatting logic
	t, err := time.Parse(time.RFC3339, o.CreatedAt)
	if err != nil {
		t, _ = time.Parse("2006-01-02 15:04:05", o.CreatedAt)
	}

	displayTime := o.CreatedAt
	if !t.IsZero() {
		displayTime = fmt.Sprintf("Time: %s", t.Local().Format("3:04 pm"))
	}

	// --- HUGE ID CHANGE BELOW ---
	// changed font-size:1.3rem to 3rem and added line-height:1 for better fit
	fmt.Fprintf(w, `
	<div class="ticket %s" id="order-%d" data-id="%d" data-created="%s">
		<div class="ticket-header">
			<span style="font-weight:bold; font-size:3rem; line-height:1;">#%d</span>
			<span style="font-weight:bold; font-size:1.2rem;">%s</span>
		</div>
		<div class="ticket-body">
			<div class="ticket-meta">
				<span>%s</span> 
				<span>Wait: <span class="elapsed-time">--:--</span></span>
			</div>
			<ul class="ticket-items">`,
		cssClass,
		o.ID,
		o.ID,
		o.CreatedAt,
		o.ID, // This ID is now huge
		o.Customer,
		displayTime,
	)

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
