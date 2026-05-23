package main

import (
	"database/sql"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv" // Driver untuk membaca .env di lokal (opsional di prod)
	_ "github.com/lib/pq"      // Driver PostgreSQL resmi untuk koneksi ke Supabase
)

// Recipe struct disesuaikan dengan skema tabel database Supabase
type Recipe struct {
	Title            string
	Desc             string
	Time             string
	Image            string
	SimpanDalam      string
	MasaSimpan       string
	TempatSimpan     string
	KualitasMaksimal string
	Bahan            string
	Langkah          string
}

// Instance database global
var db *sql.DB

// Fungsi inisialisasi koneksi ke Supabase Cloud
func initDatabase() {
	var err error
	// Membaca string koneksi dari environment variable sistem
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("Error: DATABASE_URL tidak ditemukan di environment variables ataupun file .env!")
	}

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Gagal membuka gerbang koneksi database:", err)
	}

	// Cek apakah database benar-benar merespon
	err = db.Ping()
	if err != nil {
		log.Fatal("Koneksi gagal! Server Supabase tidak merespon bray:", err)
	}
	log.Println("Berhasil terhubung ke database Supabase!")
}

// Mengambil seluruh data resep dari database Supabase
func getRecipesFromDB() []Recipe {
	var recipes []Recipe
	// Menggunakan COALESCE agar jika ada kolom bernilai NULL di DB, dibaca sebagai string kosong "" oleh Go
	query := `SELECT title, description, time_estimation, image_url, 
	          coalesce(simpan_dalam, ''), coalesce(masa_simpan, ''), coalesce(tempat_simpan, ''), 
	          coalesce(kualitas_maksimal, ''), coalesce(bahan, ''), coalesce(langkah, '') 
	          FROM recipes ORDER BY created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		log.Println("Gagal mengambil data dari database:", err)
		return recipes
	}
	defer rows.Close()

	for rows.Next() {
		var r Recipe
		err := rows.Scan(&r.Title, &r.Desc, &r.Time, &r.Image, &r.SimpanDalam, &r.MasaSimpan, &r.TempatSimpan, &r.KualitasMaksimal, &r.Bahan, &r.Langkah)
		if err != nil {
			log.Println("Gagal membaca struktur baris data:", err)
			continue
		}
		recipes = append(recipes, r)
	}
	return recipes
}

// Menyimpan entri resep baru ke database Supabase
func saveRecipeToDB(r Recipe) error {
	query := `INSERT INTO recipes (title, description, time_estimation, image_url, simpan_dalam, masa_simpan, tempat_simpan, kualitas_maksimal, bahan, langkah) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := db.Exec(query, r.Title, r.Desc, r.Time, r.Image, r.SimpanDalam, r.MasaSimpan, r.TempatSimpan, r.KualitasMaksimal, r.Bahan, r.Langkah)
	return err
}

func main() {
	// Memuat file .env jika ada (hanya berjalan untuk pengujian lokal di laptop)
	_ = godotenv.Load()

	// Jalankan inisialisasi koneksi database saat server start
	initDatabase()
	defer db.Close()

	// Menangani file aset statis (CSS/Gambar)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route Utama (/)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		pageStr := r.URL.Query().Get("page")
		page := 1
		if pageStr != "" {
			p, err := strconv.Atoi(pageStr)
			if err == nil && p > 0 {
				page = p
			}
		}

		combined := getRecipesFromDB()
		var filtered []Recipe
		for _, res := range combined {
			if search == "" || strings.Contains(strings.ToLower(res.Title), strings.ToLower(search)) {
				filtered = append(filtered, res)
			}
		}

		limit := 9
		total := len(filtered)
		offset := (page - 1) * limit
		var paginated []Recipe
		if offset < total {
			end := offset + limit
			if end > total {
				end = total
			}
			paginated = filtered[offset:end]
		}

		data := struct {
			Recipes     []Recipe
			CurrentPage int
			Search      string
		}{
			Recipes:     paginated,
			CurrentPage: page,
			Search:      search,
		}
		tmpl, _ := template.ParseFiles("index.html")
		tmpl.Execute(w, data)
	})

	// Route Detail (/recipe)
	http.HandleFunc("/recipe", func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Query().Get("title")
		combined := getRecipesFromDB()
		var selectedRecipe Recipe
		found := false
		for _, res := range combined {
			if res.Title == title {
				selectedRecipe = res
				found = true
				break
			}
		}
		if !found {
			http.NotFound(w, r)
			return
		}
		tmpl, err := template.ParseFiles("detail.html")
		if err != nil {
			http.Error(w, "File detail.html gak ketemu!", http.StatusNotFound)
			return
		}
		tmpl.Execute(w, selectedRecipe)
	})

	// Route Tambah (/add-recipe)
	http.HandleFunc("/add-recipe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			tmpl, _ := template.ParseFiles("add.html")
			tmpl.Execute(w, nil)
			return
		}

		file, handler, err := r.FormFile("image")
		imageName := "default.png"
		if err == nil {
			defer file.Close()
			imageName = handler.Filename
			f, _ := os.OpenFile("./static/"+imageName, os.O_WRONLY|os.O_CREATE, 0666)
			defer f.Close()
			io.Copy(f, file)
		}

		newR := Recipe{
			Title:            r.FormValue("title"),
			Desc:             r.FormValue("desc"),
			Time:             r.FormValue("time"),
			Image:            imageName,
			SimpanDalam:      r.FormValue("simpan_dalam"),
			MasaSimpan:       r.FormValue("masa_simpan"),
			TempatSimpan:     r.FormValue("tempat_simpan"),
			KualitasMaksimal: r.FormValue("kualitas_maksimal"),
			Bahan:            r.FormValue("bahan"),
			Langkah:          r.FormValue("langkah"),
		}

		err = saveRecipeToDB(newR)
		if err != nil {
			log.Println("Gagal menyimpan resep baru ke DB:", err)
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// Route Moderasi Edit
	http.HandleFunc("/moderasi/edit", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		title := r.URL.Query().Get("title")
		if key != "kuncirahasia77" {
			http.Error(w, "Akses ditolak bray!", http.StatusUnauthorized)
			return
		}

		combined := getRecipesFromDB()
		var selected Recipe
		found := false
		for _, res := range combined {
			if res.Title == title {
				selected = res
				found = true
				break
			}
		}

		if r.Method == "GET" {
			if !found {
				http.NotFound(w, r)
				return
			}
			tmpl, _ := template.ParseFiles("edit.html")
			tmpl.Execute(w, selected)
			return
		}

		if r.Method == "POST" {
			imageName := r.FormValue("old_image")
			file, handler, err := r.FormFile("image")
			if err == nil {
				defer file.Close()
				imageName = handler.Filename
				f, _ := os.OpenFile("./static/"+imageName, os.O_WRONLY|os.O_CREATE, 0666)
				defer f.Close()
				io.Copy(f, file)
			}

			// Mengubah data langsung di baris database berdasarkan judul lama
			query := `UPDATE recipes SET title=$1, description=$2, time_estimation=$3, image_url=$4, 
			          simpan_dalam=$5, masa_simpan=$6, tempat_simpan=$7, kualitas_maksimal=$8, bahan=$9, langkah=$10 
			          WHERE title=$11`
			_, err = db.Exec(query, r.FormValue("title"), r.FormValue("desc"), r.FormValue("time"), imageName,
				r.FormValue("simpan_dalam"), r.FormValue("masa_simpan"), r.FormValue("tempat_simpan"),
				r.FormValue("kualitas_maksimal"), r.FormValue("bahan"), r.FormValue("langkah"), title)

			if err != nil {
				log.Println("Gagal memperbarui data resep di DB:", err)
			}

			http.Redirect(w, r, "/recipe?title="+r.FormValue("title"), http.StatusSeeOther)
		}
	})

	// Route Simpan Makanan
	http.HandleFunc("/Simpan-Makanan", func(w http.ResponseWriter, r *http.Request) {
		tmpl, _ := template.ParseFiles("simpan.html")
		tmpl.Execute(w, nil)
	})

	// Route Favorit
	http.HandleFunc("/favorit", func(w http.ResponseWriter, r *http.Request) {
		tmpl, _ := template.ParseFiles("favorit.html")
		tmpl.Execute(w, nil)
	})

	// Route Detail Simpan
	http.HandleFunc("/detailsimpan", func(w http.ResponseWriter, r *http.Request) {
		tmpl, _ := template.ParseFiles("detailsimpan.html")
		tmpl.Execute(w, nil)
	})

	// Route Delete
	http.HandleFunc("/moderasi/delete", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		title := r.URL.Query().Get("title")
		if key != "kuncirahasia77" {
			http.Error(w, "Akses ditolak!", http.StatusUnauthorized)
			return
		}

		// Menghapus baris resep langsung dari database Supabase
		query := "DELETE FROM recipes WHERE title = $1"
		_, err := db.Exec(query, title)
		if err != nil {
			log.Println("Gagal menghapus data resep dari DB:", err)
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// DETEKSI PORT DINAMIS (Hugging Face mewajibkan port dibaca dari env sistem)
	port := os.Getenv("PORT")
	if port == "" {
		port = "7860" // Menggunakan port default Hugging Face Spaces jika kosong
	}

	log.Println("Server running on port:", port)
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
