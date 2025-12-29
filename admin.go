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
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
	<style>
		/* HTMX Loading State Logic - Tailwind doesn't have a built-in parent selector for this specific lib behavior */
		.loader-overlay.htmx-request { opacity: 1; pointer-events: all; }
		.tool-panel { display: none; }
		.tool-panel.show { display: block; }
	</style>
	<script>
		function switchTab(btn, mode) {
			const container = btn.closest('.img-container');
			// Reset buttons
			container.querySelectorAll('.tool-btn').forEach(b => {
				b.classList.remove('bg-gray-900', 'text-white', 'border-gray-900');
				b.classList.add('bg-white', 'text-gray-700', 'border-gray-300');
			});
			// Active button
			btn.classList.remove('bg-white', 'text-gray-700', 'border-gray-300');
			btn.classList.add('bg-gray-900', 'text-white', 'border-gray-900');
			
			// Switch Panels
			container.querySelectorAll('.tool-panel').forEach(p => p.classList.remove('show'));
			container.querySelector('.panel-' + mode).classList.add('show');
		}

		// 🆕 Detect unsaved changes (Tailwind Logic Update)
		document.addEventListener('input', function(e) {
			const card = e.target.closest('.admin-card');
			if (card) {
				const saveBtn = card.querySelector('.btn-save');
				if (saveBtn) {
					// 1. Remove success state
					saveBtn.classList.remove('bg-green-600', 'text-white');
					// Remove hidden state
					saveBtn.classList.remove('opacity-0', 'pointer-events-none');
					
					// 2. Add dirty state (Orange & Bounce)
					saveBtn.classList.add('bg-orange-600', 'text-white', 'animate-bounce', 'opacity-100', 'pointer-events-auto');
					saveBtn.innerText = "💾 Save Changes *";
				}
			}
		});
	</script>
</head>
<body class="bg-gray-50 text-gray-800 font-sans pb-20">
    <header class="bg-white shadow mb-8 sticky top-0 z-50">
        <div class="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
            <a href="/" class="no-underline text-gray-800 flex items-center gap-2 hover:text-blue-600 transition">
                <span class="text-xl font-bold">⬅ Back to Menu</span>
            </a>
            <h2 class="text-xl font-semibold text-gray-500">Live Admin Editor</h2>
        </div>
    </header>

    <main class="max-w-7xl mx-auto px-4 space-y-12">`)

	// 1. Render Existing Categories and Products
	for _, cat := range sortedCategories {
		products := categories[cat]
		fmt.Fprintf(w, "<section><h2 class='text-2xl font-bold mb-6 text-gray-800 border-b pb-2'>%s</h2><div class='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6'>", strings.ToUpper(cat))

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
					btn.classList.remove('bg-orange-600', 'animate-bounce');
					
					// 2. Add Success state
					btn.classList.add('bg-green-600');
					btn.innerText = "✔ Saved!";
				}
			}
		});
	</script>
	</body></html>`)
}

// Reusable Image UI Component (Upload or Generate)
func renderImageControls(w io.Writer, currentImgURL, promptSuggestion string, productID int) {
	if currentImgURL == "" {
		currentImgURL = "https://placehold.co/400x300?text=No+Image"
	}
	uniqueID := fmt.Sprintf("%d_%d", productID, time.Now().UnixNano())
	imgID := "img-" + uniqueID
	inputID := "input-" + uniqueID
	loaderID := "loader-" + uniqueID

	fmt.Fprintf(w, `
		<div class="img-container bg-gray-100 relative rounded-t-lg overflow-hidden group">
			
			<div id="%s" class="loader-overlay htmx-indicator absolute inset-0 bg-white/90 flex flex-col justify-center items-center z-10 opacity-0 pointer-events-none transition-opacity duration-200">
				<img src="./images/loading_indicator.svg" width="50" alt="Loading...">
				<div class="text-sm text-gray-600 mt-2">Generating...</div>
			</div>

			<img id="%s" src="%s" class="w-full aspect-[4/3] object-cover block" alt="Product Image">
			
			<input type="hidden" name="generated_image_url" id="%s">

			<div class="img-tools flex justify-center gap-2 p-2 bg-gray-50 border-b border-gray-200">
				<button type="button" class="tool-btn bg-gray-900 text-white border border-gray-900 text-xs px-3 py-1 rounded cursor-pointer transition hover:opacity-90" onclick="switchTab(this, 'upload')">Upload</button>
				<button type="button" class="tool-btn bg-white text-gray-700 border border-gray-300 text-xs px-3 py-1 rounded cursor-pointer transition hover:bg-gray-100" onclick="switchTab(this, 'gen')">✨ AI Generate</button>
			</div>

			<div class="tool-panel panel-upload show p-3 bg-gray-50 border-b border-gray-200 text-sm">
				<input type="file" name="image" accept="image/*" class="w-full text-xs text-gray-500 file:mr-2 file:py-1 file:px-2 file:rounded file:border-0 file:text-xs file:font-semibold file:bg-gray-200 file:text-gray-700 hover:file:bg-gray-300" 
					onchange="document.getElementById('%s').src = window.URL.createObjectURL(this.files[0]); this.closest('.admin-card').querySelector('.btn-save').classList.add('bg-orange-600', 'text-white', 'animate-bounce', 'opacity-100', 'pointer-events-auto'); this.closest('.admin-card').querySelector('.btn-save').classList.remove('opacity-0', 'pointer-events-none');">
			</div>

			<div class="tool-panel panel-gen p-3 bg-gray-50 border-b border-gray-200">
				<textarea name="prompt" class="w-full p-2 mb-2 border border-gray-300 rounded text-sm focus:ring-1 focus:ring-purple-500 outline-none" rows="2" placeholder="Describe image...">%s</textarea>
				
				<button type="button" class="w-full bg-purple-600 text-white border-none py-1.5 px-3 rounded text-sm font-medium hover:bg-purple-700 transition" 
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
	<section class="mt-16 p-8 bg-blue-50 border-2 border-dashed border-blue-200 rounded-xl">
		<h2 class="text-2xl font-bold mt-0 text-gray-800">✨ Add New Product Category</h2>
		<p class="text-gray-600 mb-6">Create a new category by adding its first product.</p>
		
		<form hx-post="/admin/create" hx-encoding="multipart/form-data" hx-target="body" class="grid grid-cols-1 md:grid-cols-[250px_1fr] gap-8 items-start">
			<div class="bg-white p-6 rounded-lg shadow-sm border border-gray-100">
				<label class="font-bold block mb-2 text-gray-700">Category Name</label>
				<input type="text" name="category" placeholder="e.g. Dessert" class="w-full p-3 text-lg border-2 border-gray-200 rounded focus:border-blue-500 focus:outline-none" required>
			</div>
			
			<div class="pizza-card add-card bg-white rounded-lg border-2 border-dashed border-gray-300 hover:border-blue-500 hover:shadow-lg transition-all duration-300 max-w-sm mx-auto md:mx-0 w-full flex flex-col">
				`)
	renderImageControls(w, "", "Delicious food photography", 0)
	fmt.Fprint(w, `
				<div class="p-4 flex flex-col flex-grow">
					<div class="text-center text-gray-400 font-bold mb-4 text-sm uppercase tracking-wider">First Product Details</div>
					<h3 class="mb-2"><input type="text" name="name" placeholder="Product Name" class="w-full font-bold text-lg p-1 border border-dashed border-gray-300 rounded bg-gray-50 focus:bg-white focus:border-blue-500 focus:outline-none" required></h3>
					<p class="mb-4"><textarea name="description" rows="2" placeholder="Description" class="w-full text-sm text-gray-600 p-1 border border-dashed border-gray-300 rounded bg-gray-50 focus:bg-white focus:border-blue-500 focus:outline-none"></textarea></p>
					
					<div class="mt-auto flex flex-col gap-3">
						<div class="flex justify-between items-center">
							<span class="font-bold text-gray-800">RM <input type="number" step="0.01" name="price" placeholder="0.00" class="w-20 p-1 border border-dashed border-gray-300 rounded bg-gray-50 focus:bg-white focus:border-blue-500 focus:outline-none" required></span>
							<label class="flex items-center gap-2 text-sm cursor-pointer"><input type="checkbox" name="in_stock" checked class="rounded text-blue-600"> In Stock</label>
						</div>
						<button type="submit" class="w-full bg-blue-600 text-white py-2 rounded font-medium hover:bg-blue-700 transition shadow">Create Category & Item</button>
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
		<form hx-post="/admin/create" hx-encoding="multipart/form-data" hx-target="body" class="pizza-card add-card bg-gray-50 rounded-lg border-2 border-dashed border-gray-300 opacity-80 hover:opacity-100 hover:bg-white hover:border-blue-500 hover:shadow-lg hover:-translate-y-1 transition-all duration-300 flex flex-col h-full">
			<input type="hidden" name="category" value="%s">
			`, category)

	renderImageControls(w, "", prompt, 0)

	fmt.Fprintf(w, `
			<div class="p-4 flex flex-col flex-grow">
				<div class="text-center text-gray-400 font-bold mb-4 text-xs uppercase tracking-wider">Add New %s</div>
				<h3 class="mb-2"><input type="text" name="name" placeholder="Name" class="w-full font-bold text-lg p-1 border border-dashed border-gray-300 rounded bg-white/50 focus:bg-white focus:border-blue-500 focus:outline-none" required></h3>
				<p class="mb-4"><textarea name="description" rows="2" placeholder="Description" class="w-full text-sm text-gray-600 p-1 border border-dashed border-gray-300 rounded bg-white/50 focus:bg-white focus:border-blue-500 focus:outline-none resize-none"></textarea></p>
				<div class="mt-auto flex flex-col gap-3">
					<div class="flex justify-between items-center">
						<span class="font-bold text-gray-800">RM <input type="number" step="0.01" name="price" placeholder="0.00" class="w-20 p-1 border border-dashed border-gray-300 rounded bg-white/50 focus:bg-white focus:border-blue-500 focus:outline-none" required></span>
						<label class="flex items-center gap-2 text-sm cursor-pointer"><input type="checkbox" name="in_stock" checked class="rounded text-blue-600"> In Stock</label>
					</div>
					<button type="submit" class="w-full bg-blue-600 text-white py-2 rounded font-medium hover:bg-blue-700 transition">➕ Create Item</button>
				</div>
			</div>
		</form>`, strings.Title(category))
}

func renderAdminCard(w http.ResponseWriter, p Product) {
	opacityClass := ""
	if !p.InStock {
		opacityClass = "opacity-60 grayscale-[0.8]"
	}
	checked := ""
	if p.InStock {
		checked = "checked"
	}

	prompt := fmt.Sprintf("%s %s, food photography", p.Name, p.Category)

	fmt.Fprintf(w, `
		<form hx-post="/admin/update" hx-encoding="multipart/form-data" hx-swap="none" class="pizza-card admin-card %s relative bg-white rounded-lg shadow-sm hover:shadow-md hover:ring-2 hover:ring-blue-500 transition-all duration-300 flex flex-col h-full">
			<input type="hidden" name="id" value="%d">
			
			<button type="button" 
				hx-delete="/admin/delete?id=%d" 
				hx-confirm="Delete '%s'?" 
				hx-target="closest .admin-card" 
				hx-swap="outerHTML"
				class="absolute top-2 right-2 z-20 w-8 h-8 flex items-center justify-center bg-white text-red-500 border border-red-200 rounded-full shadow hover:bg-red-500 hover:text-white hover:scale-110 transition-all" title="Delete Product">🗑️</button>
			`, opacityClass, p.ID, p.ID, p.Name)

	renderImageControls(w, p.ImageURL, prompt, p.ID)

	fmt.Fprintf(w, `
			<div class="p-4 flex flex-col flex-grow">
				<h3 class="mb-2"><input type="text" name="name" value="%s" class="w-full font-bold text-lg p-1 border border-dashed border-gray-300 rounded bg-transparent focus:bg-white focus:border-blue-500 focus:outline-none text-gray-800"></h3>
				<p class="mb-4"><textarea name="description" rows="2" class="w-full text-sm text-gray-600 p-1 border border-dashed border-gray-300 rounded bg-transparent focus:bg-white focus:border-blue-500 focus:outline-none resize-none">%s</textarea></p>
				<div class="mt-auto flex flex-col gap-3">
					<div class="flex justify-between items-center">
						<span class="font-bold text-gray-800">RM <input type="number" step="0.01" name="price" value="%.2f" class="w-20 p-1 border border-dashed border-gray-300 rounded bg-transparent focus:bg-white focus:border-blue-500 focus:outline-none"></span>
						<label class="flex items-center gap-2 text-sm cursor-pointer select-none"><input type="checkbox" name="in_stock" %s class="rounded text-blue-600"> In Stock</label>
					</div>
					<button type="submit" class="btn-save w-full py-2 rounded font-medium shadow transition-all duration-300 opacity-0 pointer-events-none">💾 Save Changes</button>
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
	productID, _ := strconv.Atoi(r.FormValue("product_id"))

	if prompt == "" {
		http.Error(w, "Prompt required", http.StatusBadRequest)
		return
	}

	imagePath, err := GenerateAndSaveImage(prompt)
	if err != nil {
		log.Println("Gen Error:", err)
		http.Error(w, "Generation failed", http.StatusInternalServerError)
		return
	}

	savedMsg := ""
	if productID > 0 {
		_, err := db.Exec("UPDATE products SET image_url = ? WHERE id = ?", imagePath, productID)
		if err != nil {
			log.Println("Auto-save image error:", err)
		} else {
			savedMsg = "<div class='text-green-600 text-xs text-center mt-1 font-medium'>Image Saved!</div>"
		}
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
		<img id="%s" src="%s" class="w-full aspect-[4/3] object-cover block" alt="Generated Image">
		%s
		<script>
			// Set the hidden input value
			document.getElementById("%s").value = "%s";
		</script>
	`, targetID, imagePath, savedMsg, inputID, imagePath)
}

func handleAdminCreateProduct(w http.ResponseWriter, r *http.Request) {
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
	if newImagePath == "" {
		newImagePath = r.FormValue("generated_image_url")
	}

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

func handleAdminDeleteProduct(w http.ResponseWriter, r *http.Request) {
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
