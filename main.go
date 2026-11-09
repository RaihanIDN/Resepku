package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Recipe struct
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

// Data statis sebagai pengganti database agar aplikasi tetap berjalan
var staticRecipes = []Recipe{
	{
		Title:            "Nasi Goreng Spesial",
		Desc:             "Nasi goreng lezat dengan bumbu rahasia.",
		Time:             "20 Menit",
		Image:            "Nasi_goreng.png",
		SimpanDalam:      "Suhu Ruang / Kulkas",
		MasaSimpan:       "1 Hari",
		TempatSimpan:     "Wadah tertutup",
		KualitasMaksimal: "Langsung disantap",
		Bahan:            "Nasi putih, Bawang merah, Bawang putih, Kecap, Telur",
		Langkah:          "1. Haluskan bumbu. 2. Tumis bumbu hingga harum. 3. Masukkan nasi dan telur. 4. Aduk rata dan tambahkan kecap.",
	},
    {
		Title:            "Telur Balado",
		Desc:             "Telur rebus dengan sambal balado pedas manis.",
		Time:             "30 Menit",
		Image:            "Telur_Balado.png",
		SimpanDalam:      "Suhu Ruang / Kulkas",
		MasaSimpan:       "2 Hari",
		TempatSimpan:     "Wadah tertutup rapat",
		KualitasMaksimal: "Dihangatkan sebelum disantap",
		Bahan:            "Telur ayam, Cabe merah, Tomat, Bawang merah, Bawang putih",
		Langkah:          "1. Rebus telur lalu goreng sebentar. 2. Haluskan bumbu. 3. Tumis bumbu balado. 4. Masukkan telur dan masak hingga meresap.",
	},
}

// Data memori untuk menampung resep baru selama aplikasi berjalan
var inMemoryRecipes []Recipe

func getRecipes() []Recipe {
	var combined []Recipe
	combined = append(combined, inMemoryRecipes...)
	combined = append(combined, staticRecipes...)
	return combined
}

func main() {
	// Menangani file aset statis
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

		combined := getRecipes()
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
		combined := getRecipes()
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

		imageUrl := r.FormValue("image_url")
		if imageUrl == "" {
			imageUrl = "default.png"
		}

		newR := Recipe{
			Title:            r.FormValue("title"),
			Desc:             r.FormValue("desc"),
			Time:             r.FormValue("time"),
			Image:            imageUrl,
			SimpanDalam:      r.FormValue("simpan_dalam"),
			MasaSimpan:       r.FormValue("masa_simpan"),
			TempatSimpan:     r.FormValue("tempat_simpan"),
			KualitasMaksimal: r.FormValue("kualitas_maksimal"),
			Bahan:            r.FormValue("bahan"),
			Langkah:          r.FormValue("langkah"),
		}

		// Simpan ke memori (akan hilang jika server restart)
		inMemoryRecipes = append([]Recipe{newR}, inMemoryRecipes...)
		
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// Route lainnya (Simpan, Favorit, dll) dikembalikan sederhana
	http.HandleFunc("/Simpan-Makanan", func(w http.ResponseWriter, r *http.Request) {
		tmpl, _ := template.ParseFiles("simpan.html")
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/favorit", func(w http.ResponseWriter, r *http.Request) {
		tmpl, _ := template.ParseFiles("favorit.html")
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/detailsimpan", func(w http.ResponseWriter, r *http.Request) {
		tmpl, _ := template.ParseFiles("detailsimpan.html")
		tmpl.Execute(w, nil)
	})

	// Port diset statis ke 8080 sesuai Dockerfile
	port := os.Getenv("PORT")
	if port == "" {
		port = "7860"
	}

	log.Println("Server running on port:", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
