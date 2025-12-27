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

// handleAdminPage renders the products in a grid, grouped by dynamic categories
func handleAdminPage(w http.ResponseWriter, r *http.Request) {
	// Query all products
	rows, err := db.Query("SELECT id, category, name, description, price, image_url, type_tag, in_stock FROM products")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Group products by category map
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

	// Sort categories alphabetically so they don't jump around
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
		.img-upload-wrapper { position: relative; display: block; overflow: hidden; border-radius: 8px 8px 0 0; }
		.file-input-overlay { position: absolute; bottom: 0; left: 0; right: 0; background: rgba(0,0,0,0.6); color: white; font-size: 0.7rem; padding: 5px; text-align: center; opacity: 0; transition: opacity 0.2s; cursor: pointer; }
		.img-upload-wrapper:hover .file-input-overlay { opacity: 1; }
		.toggle-switch { display: flex; align-items: center; gap: 5px; font-size: 0.8rem; cursor: pointer; }
		
		/* --- Add Card Styles --- */
		.add-card { border: 2px dashed #ccc; background-color: #f9f9f9; opacity: 0.8; transition: all 0.3s ease; display: flex; flex-direction: column; }
		.add-card:hover, .add-card:focus-within { opacity: 1; background-color: #fff; border-color: var(--primary); transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.1); }
		.add-header { text-align: center; color: #888; font-weight: bold; margin-bottom: 1rem; }
		
		/* --- New Feature: Delete Button --- */
		.btn-delete {
			position: absolute; top: 10px; right: 10px; z-index: 10;
			background: white; border: 1px solid #ffcccc; color: red;
			width: 30px; height: 30px; border-radius: 50%;
			display: flex; align-items: center; justify-content: center;
			cursor: pointer; transition: all 0.2s;
			box-shadow: 0 2px 4px rgba(0,0,0,0.1);
		}
		.btn-delete:hover { background: red; color: white; transform: scale(1.1); }

		/* --- New Feature: New Category Section --- */
		.new-category-section {
			margin-top: 4rem; padding: 2rem; border-top: 2px dashed #ccc;
			background-color: #f0f4f8; border-radius: 12px;
		}
		.new-cat-grid { display: grid; grid-template-columns: 250px 1fr; gap: 2rem; align-items: start; }
		@media(max-width: 768px) { .new-cat-grid { grid-template-columns: 1fr; } }
	</style>
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

		// Render products
		for _, p := range products {
			renderAdminCard(w, p)
		}

		// Render "Add Item to THIS category" card
		renderAddCard(w, cat)

		fmt.Fprintf(w, "</div></section>")
	}

	// 2. Render "Create New Category" Section at the bottom
	renderNewCategorySection(w)

	fmt.Fprint(w, `</main>
	<script>
		document.body.addEventListener('htmx:afterOnLoad', function(evt) {
			const form = evt.detail.elt;
			// Flash success on save buttons
			if(form.classList.contains('admin-card') && !form.classList.contains('add-card')) {
				const btn = form.querySelector('.btn-save');
				if(btn) {
					const originalText = btn.innerText;
					btn.innerText = "Saved!";
					btn.style.backgroundColor = "#27ae60";
					setTimeout(() => {
						btn.innerText = originalText;
						btn.style.backgroundColor = "";
					}, 2000);
				}
			}
		});
	</script>
	</body></html>`)
}

// Function to render the "Create New Category" block
func renderNewCategorySection(w http.ResponseWriter) {
	fmt.Fprint(w, `
	<section class="new-category-section">
		<h2 style="margin-top:0;">✨ Add New Product Category</h2>
		<p style="color: #666; margin-bottom: 1.5rem;">Create a new category by adding its first product.</p>
		
		<form hx-post="/admin/create" hx-encoding="multipart/form-data" hx-target="body" class="new-cat-grid">
			
			<!-- Left: Category Name Input -->
			<div style="background: white; padding: 1.5rem; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.05);">
				<label style="font-weight:bold; display:block; margin-bottom: 5px;">Category Name</label>
				<input type="text" name="category" placeholder="e.g. Dessert" class="form-input" style="width:100%; padding: 10px; font-size: 1.1rem; border: 2px solid #ddd; border-radius: 4px;" required>
			</div>

			<!-- Right: Product Details (Mini Card) -->
			<div class="pizza-card add-card" style="margin:0; max-width: 300px;">
				<div class="img-upload-wrapper" style="background: #eee; display:flex; align-items:center; justify-content:center; aspect-ratio:4/3;">
					<span style="font-size: 2rem;">➕</span>
					<label class="file-input-overlay">
						Upload Image
						<input type="file" name="image" accept="image/*" style="display:none;" onchange="this.form.querySelector('.img-upload-wrapper').style.backgroundImage = 'url(' + window.URL.createObjectURL(this.files[0]) + ')'; this.form.querySelector('.img-upload-wrapper').style.backgroundSize = 'cover';">
					</label>
				</div>
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
	// Standard Add Card for existing categories
	fmt.Fprintf(w, `
		<form hx-post="/admin/create" hx-encoding="multipart/form-data" hx-target="body" class="pizza-card add-card">
			<input type="hidden" name="category" value="%s">
			<div class="img-upload-wrapper" style="background: #eee; display:flex; align-items:center; justify-content:center; aspect-ratio:4/3;">
				<span style="font-size: 2rem;">➕</span>
				<label class="file-input-overlay">
					Upload Image
					<input type="file" name="image" accept="image/*" style="display:none;" onchange="this.form.querySelector('.img-upload-wrapper').style.backgroundImage = 'url(' + window.URL.createObjectURL(this.files[0]) + ')'; this.form.querySelector('.img-upload-wrapper').style.backgroundSize = 'cover';">
				</label>
			</div>
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
		</form>`, category, strings.Title(category))
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

	fmt.Fprintf(w, `
		<form hx-post="/admin/update" hx-encoding="multipart/form-data" hx-swap="none" class="pizza-card admin-card %s">
			<input type="hidden" name="id" value="%d">
			
			<!-- NEW: Delete Button -->
			<button type="button" 
				hx-delete="/admin/delete?id=%d" 
				hx-confirm="Are you sure you want to delete '%s'? This cannot be undone." 
				hx-target="closest .admin-card" 
				hx-swap="outerHTML"
				class="btn-delete" title="Delete Product">
				🗑️
			</button>

			<div class="img-upload-wrapper">
				<img src="%s" alt="%s" loading="lazy" style="width:100%%; aspect-ratio: 4/3; object-fit: cover;">
				<label class="file-input-overlay">
					📸 Change Image
					<input type="file" name="image" accept="image/*" style="display:none;" onchange="this.form.querySelector('img').src = window.URL.createObjectURL(this.files[0])">
				</label>
			</div>

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
		opacityClass, p.ID, p.ID, p.Name, p.ImageURL, p.Name, p.Name, p.Description, p.Price, checked)
}

// Handler to Create Product (Used by both "Add Item" and "Add Category")
func handleAdminCreateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.ParseMultipartForm(10 << 20)

	// We trim spaces to ensure "Pizza" and "pizza " don't create separate categories
	category := strings.TrimSpace(r.FormValue("category"))
	category = strings.ToLower(category) // Normalize category to lowercase

	name := r.FormValue("name")
	desc := r.FormValue("description")
	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	inStock := (r.FormValue("in_stock") == "on")

	imagePath, err := saveImageFile(r, "image")
	if err != nil {
		log.Println("Upload err:", err)
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

	// Full Reload to show new category placement or new item
	handleAdminPage(w, r)
}

// Handler to Update Product
func handleAdminUpdateProduct(w http.ResponseWriter, r *http.Request) {
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

	var err error
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

// NEW: Handler to Delete Product
func handleAdminDeleteProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM products WHERE id = ?", idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return empty string to HTMX, which removes the element from DOM
	w.WriteHeader(http.StatusOK)
}
