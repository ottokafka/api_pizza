package main

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Helper function to save uploaded file
func saveImageFile(r *http.Request, formKey string) (string, error) {
	// 1. Get the file from the request
	file, header, err := r.FormFile(formKey)
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil // No file uploaded, not an error
		}
		return "", err
	}
	defer file.Close()

	// 2. Create directory if it doesn't exist
	uploadDir := "./images"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	// 3. Create a unique filename to prevent caching issues or overwrites
	// Example: product_123_filename.webp
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
	filePath := filepath.Join(uploadDir, filename)

	// 4. Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// 5. Copy the uploaded content to the destination file
	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	// Return the path relative to the web root (assuming you serve /images static files)
	return "./images/" + filename, nil
}

// handleAdminPage renders the list of products in editable forms
func handleAdminPage(w http.ResponseWriter, r *http.Request) {
	// Query includes image_url now
	rows, err := db.Query("SELECT id, category, name, description, price, image_url, in_stock FROM products ORDER BY category, name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Admin - Apipizza</title>
    <link rel="stylesheet" href="/static/styles.css">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
	<style>
		.admin-table { width: 100%; border-collapse: collapse; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 10px rgba(0,0,0,0.05); }
		.admin-table th, .admin-table td { padding: 12px; text-align: left; border-bottom: 1px solid #eee; vertical-align: middle; }
		.admin-table th { background: var(--dark); color: white; vertical-align: top;}
		.admin-input { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; font-family: inherit; }
		.admin-input:focus { border-color: var(--primary); outline: none; }
		.stock-toggle { transform: scale(1.5); cursor: pointer; }
		.img-preview { width: 50px; height: 50px; object-fit: cover; border-radius: 4px; border: 1px solid #ddd; display: block; margin-bottom: 5px; }
		.file-input { font-size: 0.8rem; width: 180px; }
	</style>
</head>
<body>
    <header class="brand-header">
        <div class="header-flex">
            <a href="/" style="text-decoration:none; color:inherit;">
                <h1 style="margin:0; font-size: 1.4rem;">⬅ Back to Menu</h1>
            </a>
            <h2 style="margin:0; margin-left: auto;">Admin Dashboard</h2>
        </div>
    </header>

    <main class="container" style="grid-template-columns: 1fr; max-width: 1400px;">
		<table class="admin-table">
			<thead>
				<tr>
					<th style="width: 50px;">ID</th>
					<th style="width: 100px;">Cat</th>
					<th style="width: 200px;">Image</th>
					<th style="width: 200px;">Name</th>
					<th>Description</th>
					<th style="width: 100px;">Price (RM)</th>
					<th style="width: 60px;">Stock</th>
					<th style="width: 80px;">Action</th>
				</tr>
			</thead>
			<tbody>`)

	for rows.Next() {
		var p Product
		// We handle NULL image_urls gracefully in Go, usually empty string
		var imgUrl sql.NullString

		rows.Scan(&p.ID, &p.Category, &p.Name, &p.Description, &p.Price, &imgUrl, &p.InStock)

		p.ImageURL = ""
		if imgUrl.Valid {
			p.ImageURL = imgUrl.String
		}

		checked := ""
		if p.InStock {
			checked = "checked"
		}

		// HTMX Note: hx-encoding="multipart/form-data" is required for file uploads
		fmt.Fprintf(w, `
			<tr id="row-%d">
				<form hx-post="/admin/update" hx-trigger="submit" hx-swap="none" hx-encoding="multipart/form-data">
					<input type="hidden" name="id" value="%d">
					<td>%d</td>
					<td><span class="badge" style="background:#eee; color:#333;">%s</span></td>
					
					<!-- Image Column -->
					<td>
						<img src="%s" class="img-preview" alt="img">
						<input type="file" name="image" class="file-input" accept="image/*">
					</td>

					<td><input type="text" name="name" value="%s" class="admin-input"></td>
					<td><textarea name="description" class="admin-input" rows="3">%s</textarea></td>
					<td><input type="number" step="0.01" name="price" value="%.2f" class="admin-input"></td>
					<td style="text-align:center;">
						<input type="checkbox" name="in_stock" %s class="stock-toggle">
					</td>
					<td>
						<button type="submit" class="btn-add" style="padding: 8px 16px;">Save</button>
						<div id="msg-%d" style="font-size:0.8rem; color:green; min-height:1.2em; margin-top:5px;"></div>
					</td>
				</form>
				
				<script>
					document.body.addEventListener('htmx:afterOnLoad', function(evt) {
						if(evt.detail.elt.closest('tr').id === 'row-%d') {
							// Optional: Force reload to see new image if one was uploaded
							// window.location.reload(); 
							
							const msg = document.getElementById('msg-%d');
							msg.innerText = 'Saved!';
							setTimeout(() => msg.innerText = '', 2000);
						}
					});
				</script>
			</tr>
		`, p.ID, p.ID, p.ID, p.Category, p.ImageURL, p.Name, p.Description, p.Price, checked, p.ID, p.ID, p.ID)
	}

	fmt.Fprint(w, `</tbody></table></main></body></html>`)
}

// handleAdminUpdateProduct processes the form submission with file support
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

	// 2. Extract Standard Fields
	idStr := r.FormValue("id")
	name := r.FormValue("name")
	desc := r.FormValue("description")
	priceStr := r.FormValue("price")
	inStockStr := r.FormValue("in_stock")

	id, _ := strconv.Atoi(idStr)
	price, _ := strconv.ParseFloat(priceStr, 64)
	inStock := (inStockStr == "on")

	// 3. Handle File Upload
	newImagePath, err := saveImageFile(r, "image")
	if err != nil {
		fmt.Println("File upload error:", err)
		http.Error(w, "File upload failed", http.StatusInternalServerError)
		return
	}

	// 4. Update Database
	if newImagePath != "" {
		// Scenario A: User uploaded a new image -> Update image_url
		_, err = db.Exec(`UPDATE products SET name=?, description=?, price=?, in_stock=?, image_url=? WHERE id=?`,
			name, desc, price, inStock, newImagePath, id)
	} else {
		// Scenario B: No new image -> Keep existing image_url
		_, err = db.Exec(`UPDATE products SET name=?, description=?, price=?, in_stock=? WHERE id=?`,
			name, desc, price, inStock, id)
	}

	if err != nil {
		fmt.Printf("Error updating product %d: %v\n", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Returns 200 OK. HTMX will handle the UI feedback via script.
	w.WriteHeader(http.StatusOK)
}
