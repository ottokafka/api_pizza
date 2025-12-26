package main

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

// handleAdminPage renders the products in the exact grid layout as the client
func handleAdminPage(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, category, name, description, price, image_url, type_tag, in_stock FROM products ORDER BY category, name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 1. Group data (Same logic as main.go)
	categories := map[string][]Product{}
	for rows.Next() {
		var p Product
		var imgUrl sql.NullString // Handle potential nulls
		rows.Scan(&p.ID, &p.Category, &p.Name, &p.Description, &p.Price, &imgUrl, &p.TypeTag, &p.InStock)
		if imgUrl.Valid {
			p.ImageURL = imgUrl.String
		}
		categories[p.Category] = append(categories[p.Category], p)
	}

	// 2. Render Page
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Admin - Apipizza</title>
    <link rel="stylesheet" href="/static/styles.css">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
	<style>
		/* Admin Specific Overrides to make inputs look like text */
		.admin-card { position: relative; border: 2px dashed transparent; transition: border 0.3s; }
		.admin-card:hover { border-color: var(--primary); }
		
		.edit-input { 
			width: 100%; border: 1px dashed #ddd; background: rgba(255,255,255,0.8); 
			font-family: inherit; color: inherit; padding: 2px; border-radius: 4px;
		}
		.edit-input:focus { border: 1px solid var(--primary); outline: none; background: #fff; }
		
		/* Typography overrides for inputs */
		h3 .edit-input { font-size: 1.2rem; font-weight: bold; margin-bottom: 5px; }
		p .edit-input { font-size: 0.9rem; color: #666; resize: none; }
		
		/* Image Upload Overlay */
		.img-upload-wrapper { position: relative; display: block; overflow: hidden; }
		.file-input-overlay {
			position: absolute; bottom: 0; left: 0; right: 0;
			background: rgba(0,0,0,0.6); color: white;
			font-size: 0.7rem; padding: 5px; text-align: center;
			opacity: 0; transition: opacity 0.2s; cursor: pointer;
		}
		.img-upload-wrapper:hover .file-input-overlay { opacity: 1; }
		
		/* Controls Footer */
		.admin-controls {
			margin-top: 10px; padding-top: 10px; border-top: 1px solid #eee;
			display: flex; gap: 10px; align-items: center; justify-content: space-between;
		}
		
		.toggle-switch { display: flex; align-items: center; gap: 5px; font-size: 0.8rem; cursor: pointer; }
		.save-indicator { font-size: 0.8rem; color: #27ae60; opacity: 0; transition: opacity 0.5s; }
		.save-indicator.show { opacity: 1; }
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

	// 3. Render Grid by Category
	order := []string{"pizza", "pasta", "drink"}
	for _, cat := range order {
		products := categories[cat]
		fmt.Fprintf(w, "<section class='category-section'><h2 class='category-header'>%s</h2><div class='pizza-grid'>", strings.ToUpper(cat))
		for _, p := range products {
			renderAdminCard(w, p)
		}
		fmt.Fprintf(w, "</div></section>")
	}

	fmt.Fprint(w, `</main>
	<script>
		// Simple flash message logic
		document.body.addEventListener('htmx:afterOnLoad', function(evt) {
			const form = evt.detail.elt;
			if(form.tagName === 'FORM') {
				const btn = form.querySelector('.btn-save');
				const originalText = btn.innerText;
				btn.innerText = "Saved!";
				btn.style.backgroundColor = "#27ae60";
				setTimeout(() => {
					btn.innerText = originalText;
					btn.style.backgroundColor = "";
				}, 2000);
			}
		});
	</script>
	</body></html>`)
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

	// We use the exact same classes (.pizza-card) but wrap content in a form
	fmt.Fprintf(w, `
		<form hx-post="/admin/update" hx-encoding="multipart/form-data" hx-swap="none" class="pizza-card admin-card %s">
			<input type="hidden" name="id" value="%d">
			
			<!-- Image Section -->
			<div class="img-upload-wrapper">
				<img src="%s" alt="%s" loading="lazy" style="width:100%%; aspect-ratio: 4/3; object-fit: cover;">
				<label class="file-input-overlay">
					📸 Change Image
					<input type="file" name="image" accept="image/*" style="display:none;" onchange="this.form.querySelector('img').src = window.URL.createObjectURL(this.files[0])">
				</label>
			</div>

			<div class="card-content">
				<!-- Editable Title -->
				<h3><input type="text" name="name" value="%s" class="edit-input"></h3>
				
				<!-- Editable Description -->
				<p><textarea name="description" rows="2" class="edit-input">%s</textarea></p>
				
				<div class="card-footer" style="flex-direction: column; align-items: stretch;">
					
					<!-- Price Row -->
					<div style="display:flex; justify-content: space-between; align-items:center; margin-bottom: 8px;">
						<span class="price" style="font-size: 1rem;">RM 
							<input type="number" step="0.01" name="price" value="%.2f" class="edit-input" style="width: 70px; display:inline-block;">
						</span>
						<label class="toggle-switch">
							<input type="checkbox" name="in_stock" %s> In Stock
						</label>
					</div>

					<!-- Save Button -->
					<button type="submit" class="btn-add btn-save" style="width:100%%; background-color: var(--dark);">
						💾 Save Changes
					</button>

				</div>
			</div>
		</form>`,
		opacityClass, p.ID, p.ImageURL, p.Name, p.Name, p.Description, p.Price, checked)
}

// handleAdminUpdateProduct processes the form submission (File Upload + Data)
func handleAdminUpdateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse Multipart Form (Max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// 2. Extract Data
	idStr := r.FormValue("id")
	name := r.FormValue("name")
	desc := r.FormValue("description")
	priceStr := r.FormValue("price")
	inStockStr := r.FormValue("in_stock")

	id, _ := strconv.Atoi(idStr)
	price, _ := strconv.ParseFloat(priceStr, 64)
	inStock := (inStockStr == "on") // Checkbox sends "on" if checked, nothing if unchecked

	// 3. Handle File Upload
	newImagePath, err := saveImageFile(r, "image")
	if err != nil {
		fmt.Println("File upload error:", err)
		http.Error(w, "File upload failed", http.StatusInternalServerError)
		return
	}

	// 4. Update Database
	if newImagePath != "" {
		_, err = db.Exec(`UPDATE products SET name=?, description=?, price=?, in_stock=?, image_url=? WHERE id=?`,
			name, desc, price, inStock, newImagePath, id)
	} else {
		_, err = db.Exec(`UPDATE products SET name=?, description=?, price=?, in_stock=? WHERE id=?`,
			name, desc, price, inStock, id)
	}

	if err != nil {
		fmt.Printf("Error updating product %d: %v\n", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
