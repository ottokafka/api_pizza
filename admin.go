package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Helper function to save uploaded file
func saveImageFile(r *http.Request, formKey string) (string, error) {
	file, header, err := r.FormFile(formKey)
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}
		return "", err
	}
	defer file.Close()

	uploadDir := "./images"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	// Create unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
	filePath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return "./images/" + filename, nil
}

// handleAdminPage renders the products
func handleAdminPage(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, category, name, description, price, image_url, type_tag, in_stock FROM products")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	categories := map[string][]Product{}
	for rows.Next() {
		var p Product
		var imgUrl sql.NullString
		rows.Scan(&p.ID, &p.Category, &p.Name, &p.Description, &p.Price, &imgUrl, &p.TypeTag, &p.InStock)
		if imgUrl.Valid {
			p.ImageURL = imgUrl.String
		}
		categories[p.Category] = append(categories[p.Category], p)
	}

	var sortedCategories []string
	for cat := range categories {
		sortedCategories = append(sortedCategories, cat)
	}
	sort.Strings(sortedCategories)

	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Admin - Apipizza</title>
    <link rel="stylesheet" href="/static/styles.css">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
	<style>
		/* --- Existing Styles --- */
		.admin-card { position: relative; border: 2px dashed transparent; transition: border 0.3s; background: white; border-radius: 8px; box-shadow: 0 2px 5px rgba(0,0,0,0.05); }
		.admin-card:hover { border-color: var(--primary); }
		.edit-input { width: 100%; border: 1px dashed #ddd; background: rgba(255,255,255,0.8); font-family: inherit; color: inherit; padding: 4px; border-radius: 4px; }
		.edit-input:focus { border: 1px solid var(--primary); outline: none; background: #fff; }
		h3 .edit-input { font-size: 1.2rem; font-weight: bold; margin-bottom: 5px; }
		p .edit-input { font-size: 0.9rem; color: #666; resize: none; }
		
		/* --- Image Area & Tabs --- */
		.img-container { background: #eee; position: relative; overflow: hidden; border-radius: 8px 8px 0 0; }
		.img-preview { width:100%; aspect-ratio: 4/3; object-fit: cover; display: block; }
		
		/* --- Loading Indicator Overlay --- */
		.loader-overlay {
			position: absolute; top: 0; left: 0; width: 100%; height: 100%;
			background: rgba(255, 255, 255, 0.85);
			display: flex; flex-direction: column; justify-content: center; align-items: center;
			z-index: 10;
			opacity: 0; pointer-events: none; transition: opacity 0.2s ease-in-out;
		}
		.loader-overlay.htmx-request { opacity: 1; pointer-events: all; }

		.img-tools { padding: 5px; background: #f8f9fa; border-bottom: 1px solid #eee; display: flex; gap: 5px; justify-content: center; }
		.tool-btn { font-size: 0.75rem; padding: 4px 8px; border: 1px solid #ccc; background: white; cursor: pointer; border-radius: 4px; }
		.tool-btn.active { background: var(--dark); color: white; border-color: var(--dark); }
		
		.tool-panel { padding: 10px; background: #f0f0f0; display: none; border-bottom: 1px solid #ddd; font-size: 0.85rem; }
		.tool-panel.show { display: block; }
		
		/* Generation UI */
		.gen-input { width: 100%; padding: 5px; margin-bottom: 5px; border: 1px solid #ccc; border-radius: 4px; }
		.btn-gen { background: #8e44ad; color: white; border: none; padding: 5px 10px; border-radius: 4px; cursor: pointer; width: 100%; }
		.btn-gen:hover { background: #732d91; }
		
		.toggle-switch { display: flex; align-items: center; gap: 5px; font-size: 0.8rem; cursor: pointer; }
		
		/* --- Add Card Styles --- */
		.add-card { border: 2px dashed #ccc; background-color: #f9f9f9; opacity: 0.8; transition: all 0.3s ease; display: flex; flex-direction: column; }
		.add-card:hover, .add-card:focus-within { opacity: 1; background-color: #fff; border-color: var(--primary); transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.1); }
		.add-header { text-align: center; color: #888; font-weight: bold; margin-bottom: 1rem; }
		
		.btn-delete {
			position: absolute; top: 10px; right: 10px; z-index: 10;
			background: white; border: 1px solid #ffcccc; color: red;
			width: 30px; height: 30px; border-radius: 50%;
			display: flex; align-items: center; justify-content: center;
			cursor: pointer; transition: all 0.2s;
			box-shadow: 0 2px 4px rgba(0,0,0,0.1);
		}
		.btn-delete:hover { background: red; color: white; transform: scale(1.1); }

		.new-category-section {
			margin-top: 4rem; padding: 2rem; border-top: 2px dashed #ccc;
			background-color: #f0f4f8; border-radius: 12px;
		}
		.new-cat-grid { display: grid; grid-template-columns: 250px 1fr; gap: 2rem; align-items: start; }
		@media(max-width: 768px) { .new-cat-grid { grid-template-columns: 1fr; } }

		/* --- 🆕 UX IMPROVEMENTS: SAVE BUTTON --- */
		@keyframes bounce-gentle {
			0%, 100% { transform: translateY(0); }
			50% { transform: translateY(-4px); }
		}

		/* Default State: Hidden */
		.btn-save {
			opacity: 0; 
			pointer-events: none; /* Cannot click when hidden */
			transition: all 0.3s ease;
		}

		/* State 1: Dirty / Unsaved Changes (Visible & Bouncing) */
		.btn-dirty {
			opacity: 1 !important;
			pointer-events: auto !important;
			animation: bounce-gentle 2s infinite ease-in-out;
			background-color: #d35400 !important; /* Orange/Warning color */
			box-shadow: 0 4px 6px rgba(0,0,0,0.2);
			color: white;
		}

		/* State 2: Success / Saved (Visible & Green & No Bounce) */
		.btn-success {
			opacity: 1 !important;
			pointer-events: none; /* Prevent double click while showing success */
			background-color: #27ae60 !important; /* Green */
			animation: none !important;
			color: white;
		}

	</style>
	<script>
		function switchTab(btn, mode) {
			const container = btn.closest('.img-container');
			container.querySelectorAll('.tool-btn').forEach(b => b.classList.remove('active'));
			btn.classList.add('active');
			container.querySelectorAll('.tool-panel').forEach(p => p.classList.remove('show'));
			container.querySelector('.panel-' + mode).classList.add('show');
		}

		// 🆕 Detect unsaved changes (Logic Update)
		document.addEventListener('input', function(e) {
			const card = e.target.closest('.admin-card');
			if (card) {
				const saveBtn = card.querySelector('.btn-save');
				if (saveBtn) {
					// 1. Remove success state if user edits again
					saveBtn.classList.remove('btn-success');
					
					// 2. Add dirty state (Orange & Bounce)
					saveBtn.classList.add('btn-dirty');
					saveBtn.innerText = "💾 Save Changes *";
				}
			}
		});
	</script>
</head>
<body>
    <header class="brand-header">
        <div class="header-flex">
            <a href="/" style="text-decoration:none; color:inherit;">
                <h1 style="margin:0; font-size: 1.4rem;">⬅ Back to Menu</h1>
            </a>
            <h2 style="margin:0; margin-left: auto;">Live Admin Editor</h2>
        </div>
    </header>

    <main class="container" style="grid-template-columns: 1fr; max-width: 1200px;">`)

	// 1. Render Existing Categories and Products
	for _, cat := range sortedCategories {
		products := categories[cat]
		fmt.Fprintf(w, "<section class='category-section'><h2 class='category-header'>%s</h2><div class='pizza-grid'>", strings.ToUpper(cat))

		for _, p := range products {
			renderAdminCard(w, p)
		}
		renderAddCard(w, cat)

		fmt.Fprintf(w, "</div></section>")
	}

	// 2. Render "Create New Category" Section
	renderNewCategorySection(w)

	fmt.Fprint(w, `</main>
	<script>
		// 🆕 HTMX Handler for Save State
		document.body.addEventListener('htmx:afterOnLoad', function(evt) {
			const form = evt.detail.elt;
			if(form.classList.contains('admin-card') && !form.classList.contains('add-card')) {
				const btn = form.querySelector('.btn-save');
				if(btn) {
					// 1. Remove dirty animation
					btn.classList.remove('btn-dirty');
					
					// 2. Add Success state (Green, Visible, Static)
					btn.classList.add('btn-success');
					btn.innerText = "✔ Saved!";

					// We do NOT set a timeout to hide it here.
					// It stays green until the user makes another change (handled by the 'input' event above).
				}
			}
		});
	</script>
	</body></html>`)
}

// Reusable Image UI Component (Upload or Generate)
// 🆕 Added productID parameter to support auto-saving
func renderImageControls(w io.Writer, currentImgURL, promptSuggestion string, productID int) {
	if currentImgURL == "" {
		currentImgURL = "https://placehold.co/400x300?text=No+Image"
	}
	uniqueID := fmt.Sprintf("%d_%d", productID, time.Now().UnixNano()) // Added Product ID to uniqueness
	imgID := "img-" + uniqueID
	inputID := "input-" + uniqueID
	loaderID := "loader-" + uniqueID

	fmt.Fprintf(w, `
		<div class="img-container">
			
			<div id="%s" class="loader-overlay htmx-indicator">
				<img src="./images/loading_indicator.svg" width="50" alt="Loading...">
				<div style="font-size: 0.8rem; color: #555; margin-top:5px;">Generating...</div>
			</div>

			<img id="%s" src="%s" class="img-preview" alt="Product Image">
			
			<input type="hidden" name="generated_image_url" id="%s">

			<div class="img-tools">
				<button type="button" class="tool-btn active" onclick="switchTab(this, 'upload')">Upload</button>
				<button type="button" class="tool-btn" onclick="switchTab(this, 'gen')">✨ AI Generate</button>
			</div>

			<div class="tool-panel panel-upload show">
				<input type="file" name="image" accept="image/*" style="width:100%%;" 
					onchange="document.getElementById('%s').src = window.URL.createObjectURL(this.files[0]); this.closest('.admin-card').querySelector('.btn-save').classList.add('btn-dirty');">
			</div>

			<div class="tool-panel panel-gen">
				<textarea name="prompt" class="gen-input" rows="2" placeholder="Describe image...">%s</textarea>
				
				<!-- 🆕 Pass product_id to backend for auto-save -->
				<button type="button" class="btn-gen" 
					hx-post="/admin/generate-image" 
					hx-target="#%s" 
					hx-swap="outerHTML"
					hx-indicator="#%s"
					hx-vals='js:{prompt: event.target.previousElementSibling.value, target_id: "%s", input_id: "%s", product_id: "%d"}'>
					Generate & Auto-Save
				</button>
			</div>
		</div>
	`, loaderID, imgID, currentImgURL, inputID, imgID, promptSuggestion, imgID, loaderID, imgID, inputID, productID)
}

func renderNewCategorySection(w http.ResponseWriter) {
	fmt.Fprint(w, `
	<section class="new-category-section">
		<h2 style="margin-top:0;">✨ Add New Product Category</h2>
		<p style="color: #666; margin-bottom: 1.5rem;">Create a new category by adding its first product.</p>
		
		<form hx-post="/admin/create" hx-encoding="multipart/form-data" hx-target="body" class="new-cat-grid">
			<div style="background: white; padding: 1.5rem; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.05);">
				<label style="font-weight:bold; display:block; margin-bottom: 5px;">Category Name</label>
				<input type="text" name="category" placeholder="e.g. Dessert" class="form-input" style="width:100%; padding: 10px; font-size: 1.1rem; border: 2px solid #ddd; border-radius: 4px;" required>
			</div>
			<div class="pizza-card add-card" style="margin:0; max-width: 300px;">
				`)
	// New products have ID 0
	renderImageControls(w, "", "Delicious food photography", 0)
	fmt.Fprint(w, `
				<div class="card-content">
					<div class="add-header">First Product Details</div>
					<h3><input type="text" name="name" placeholder="Product Name" class="edit-input" required></h3>
					<p><textarea name="description" rows="2" placeholder="Description" class="edit-input"></textarea></p>
					
					<div class="card-footer" style="flex-direction: column; align-items: stretch;">
						<div style="display:flex; justify-content: space-between; align-items:center; margin-bottom: 8px;">
							<span class="price">RM <input type="number" step="0.01" name="price" placeholder="0.00" class="edit-input" style="width: 70px;" required></span>
							<label class="toggle-switch"><input type="checkbox" name="in_stock" checked> In Stock</label>
						</div>
						<button type="submit" class="btn-add" style="width:100%;">Create Category & Item</button>
					</div>
				</div>
			</div>
		</form>
	</section>
	`)
}

func renderAddCard(w http.ResponseWriter, category string) {
	prompt := fmt.Sprintf("A delicious %s", category)

	fmt.Fprintf(w, `
		<form hx-post="/admin/create" hx-encoding="multipart/form-data" hx-target="body" class="pizza-card add-card">
			<input type="hidden" name="category" value="%s">
			`, category)

	// New products have ID 0
	renderImageControls(w, "", prompt, 0)

	fmt.Fprintf(w, `
			<div class="card-content">
				<div class="add-header">Add New %s</div>
				<h3><input type="text" name="name" placeholder="Name" class="edit-input" required></h3>
				<p><textarea name="description" rows="2" placeholder="Description" class="edit-input"></textarea></p>
				<div class="card-footer" style="flex-direction: column; align-items: stretch;">
					<div style="display:flex; justify-content: space-between; align-items:center; margin-bottom: 8px;">
						<span class="price">RM <input type="number" step="0.01" name="price" placeholder="0.00" class="edit-input" style="width: 70px;" required></span>
						<label class="toggle-switch"><input type="checkbox" name="in_stock" checked> In Stock</label>
					</div>
					<button type="submit" class="btn-add" style="width:100%%;">➕ Create Item</button>
				</div>
			</div>
		</form>`, strings.Title(category))
}

func renderAdminCard(w http.ResponseWriter, p Product) {
	opacityClass := ""
	if !p.InStock {
		opacityClass = "product-unavailable"
	}
	checked := ""
	if p.InStock {
		checked = "checked"
	}

	prompt := fmt.Sprintf("%s %s, food photography", p.Name, p.Category)

	fmt.Fprintf(w, `
		<form hx-post="/admin/update" hx-encoding="multipart/form-data" hx-swap="none" class="pizza-card admin-card %s">
			<input type="hidden" name="id" value="%d">
			
			<button type="button" 
				hx-delete="/admin/delete?id=%d" 
				hx-confirm="Delete '%s'?" 
				hx-target="closest .admin-card" 
				hx-swap="outerHTML"
				class="btn-delete" title="Delete Product">🗑️</button>
			`, opacityClass, p.ID, p.ID, p.Name)

	// Pass existing Product ID
	renderImageControls(w, p.ImageURL, prompt, p.ID)

	fmt.Fprintf(w, `
			<div class="card-content">
				<h3><input type="text" name="name" value="%s" class="edit-input"></h3>
				<p><textarea name="description" rows="2" class="edit-input">%s</textarea></p>
				<div class="card-footer" style="flex-direction: column; align-items: stretch;">
					<div style="display:flex; justify-content: space-between; align-items:center; margin-bottom: 8px;">
						<span class="price">RM <input type="number" step="0.01" name="price" value="%.2f" class="edit-input" style="width: 70px;"></span>
						<label class="toggle-switch"><input type="checkbox" name="in_stock" %s> In Stock</label>
					</div>
					<button type="submit" class="btn-add btn-save" style="width:100%%; background-color: var(--dark);">💾 Save Changes</button>
				</div>
			</div>
		</form>`,
		p.Name, p.Description, p.Price, checked)
}

// ---------------- HANDLERS ----------------

func handleAdminGenerateImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	prompt := r.FormValue("prompt")
	targetID := r.FormValue("target_id")
	inputID := r.FormValue("input_id")
	// 🆕 Capture Product ID
	productID, _ := strconv.Atoi(r.FormValue("product_id"))

	if prompt == "" {
		http.Error(w, "Prompt required", http.StatusBadRequest)
		return
	}

	// Assumes GenerateAndSaveImage is defined elsewhere in your package
	imagePath, err := GenerateAndSaveImage(prompt)
	if err != nil {
		log.Println("Gen Error:", err)
		http.Error(w, "Generation failed", http.StatusInternalServerError)
		return
	}

	// 🆕 Auto-Save Logic: If product exists, save to DB immediately
	savedMsg := ""
	if productID > 0 {
		_, err := db.Exec("UPDATE products SET image_url = ? WHERE id = ?", imagePath, productID)
		if err != nil {
			log.Println("Auto-save image error:", err)
		} else {
			savedMsg = "<div style='color:green; font-size:0.8rem; text-align:center; margin-top:5px;'>Image Saved!</div>"
		}
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
		<img id="%s" src="%s" class="img-preview" alt="Generated Image">
		%s
		<script>
			// Set the hidden input value (for manual form submission backup or new items)
			document.getElementById("%s").value = "%s";
		</script>
	`, targetID, imagePath, savedMsg, inputID, imagePath)
}

func handleAdminCreateProduct(w http.ResponseWriter, r *http.Request) {
	// ... (Same as before) ...
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.ParseMultipartForm(10 << 20)

	category := strings.ToLower(strings.TrimSpace(r.FormValue("category")))
	name := r.FormValue("name")
	desc := r.FormValue("description")
	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	inStock := (r.FormValue("in_stock") == "on")

	imagePath, err := saveImageFile(r, "image")
	if err != nil {
		log.Println("Upload err:", err)
	}
	if imagePath == "" {
		imagePath = r.FormValue("generated_image_url")
	}
	if imagePath == "" {
		imagePath = "https://placehold.co/400x300?text=No+Image"
	}

	_, err = db.Exec(`INSERT INTO products (category, name, description, price, in_stock, image_url, type_tag) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		category, name, desc, price, inStock, imagePath, "")

	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	handleAdminPage(w, r)
}

func handleAdminUpdateProduct(w http.ResponseWriter, r *http.Request) {
	// ... (Same as before) ...
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.ParseMultipartForm(10 << 20)

	id, _ := strconv.Atoi(r.FormValue("id"))
	name := r.FormValue("name")
	desc := r.FormValue("description")
	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	inStock := (r.FormValue("in_stock") == "on")

	newImagePath, _ := saveImageFile(r, "image")
	// If a new generated image was set in the hidden input, use it
	if newImagePath == "" {
		newImagePath = r.FormValue("generated_image_url")
	}

	var err error
	// Only update image path if a new one is provided (uploaded or generated)
	if newImagePath != "" {
		_, err = db.Exec(`UPDATE products SET name=?, description=?, price=?, in_stock=?, image_url=? WHERE id=?`,
			name, desc, price, inStock, newImagePath, id)
	} else {
		_, err = db.Exec(`UPDATE products SET name=?, description=?, price=?, in_stock=? WHERE id=?`,
			name, desc, price, inStock, id)
	}

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handleAdminDeleteProduct(w http.ResponseWriter, r *http.Request) {
	// ... (Same as before) ...
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.URL.Query().Get("id")
	if _, err := db.Exec("DELETE FROM products WHERE id = ?", idStr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
