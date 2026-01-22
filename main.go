package main

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Recipe
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

// allRecipes
var allRecipes = []Recipe{
	{Title: "Keripik Kentang", Desc: "Renyah disetiap gigitan, gurihnya baluran bumbu. Keripik Kentang bikin nagih", Time: "20-30 Menit", Image: "Keripik_kentang.png"},
	{Title: "Kentang Tumbuk", Desc: "Selembut awan, semewah mentega. Rasakan kehangatan Mashed Potato yang meleleh di lidah.", Time: "20-30 Menit", Image: "Kentang_tumbuk.png"},
	{Title: "Kentang Balado", Desc: "Warnanya yang merah menggoda adalah janji rasa yang luar biasa. Kentang Balado yang tak hanya cantik, tapi juga lezat!", Time: "1 Jam +", Image: "Kentang_balado.png"},
	{Title: "Telur Dadar Padang", Desc: "Bukan telur dadar biasa. Rasakan kekayaan bumbu yang terkunci di setiap lapisan telur dadar setebal bantal ini.", Time: "20-30 Menit", Image: "Telur_dadar_padang.png"},
	{Title: "Telur Balado", Desc: "Rasa klasik yang selalu dicintai. Telur Balado kami dibuat dengan bumbu otentik dan cinta.", Time: "1 Jam +", Image: "Telur_Balado.png"},
	{Title: "Omelete Mie", Desc: "Kreasi paling jenius dari anak kos. Omelet Mie ini membuktikan bahwa makanan enak tidak harus rumit!", Time: "15-20 Menit", Image: "Omelet_mie.png"},
	{Title: "Nasi Goreng Seafood", Desc: "Rasa yang medok dan aroma yang menggoda! Nasi Goreng Seafood dengan bumbu rahasia.", Time: "30-40 Menit", Image: "Nasigoreng_Seafood.png"},
	{Title: "Nasi Goreng", Desc: "Comfort food sejati. Nikmati kehangatan dan kelezatan yang tiada tara dari seporsi Nasi Goreng.", Time: "20-30 Menit", Image: "Nasi_goreng.png"},
	{Title: "Nasi Goreng Cabe Hijau", Desc: "Pedasnya cabai hijau lebih light namun lebih aromatik.", Time: "15-20 Menit", Image: "Nasigoreng_cabeijo.png"},
}

// getCombinedRecipes: Gabung data statis & data dari resep.json
func getCombinedRecipes() []Recipe {
	file, err := os.ReadFile("resep.json")
	if err != nil || len(file) == 0 {
		return allRecipes
	}
	var fromFile []Recipe
	json.Unmarshal(file, &fromFile)
	return append(allRecipes, fromFile...)
}

// saveRecipeToFile: Simpan resep baru ke JSON
func saveRecipeToFile(r Recipe) {
	file, err := os.ReadFile("resep.json")
	var saved []Recipe

	if err == nil && len(file) > 0 {
		json.Unmarshal(file, &saved)
	}

	if saved == nil {
		saved = []Recipe{}
	}

	saved = append(saved, r)
	data, _ := json.MarshalIndent(saved, "", "  ")
	_ = os.WriteFile("resep.json", data, 0644)
}

func main() {
	// Menangani file statis (CSS/Gambar)
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

		combined := getCombinedRecipes()
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
		combined := getCombinedRecipes()
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
		saveRecipeToFile(newR)
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

		combined := getCombinedRecipes()
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
			fileContent, _ := os.ReadFile("resep.json")
			var saved []Recipe
			json.Unmarshal(fileContent, &saved)
			imageName := r.FormValue("old_image")
			file, handler, err := r.FormFile("image")
			if err == nil {
				defer file.Close()
				imageName = handler.Filename
				f, _ := os.OpenFile("./static/"+imageName, os.O_WRONLY|os.O_CREATE, 0666)
				defer f.Close()
				io.Copy(f, file)
			}
			for i, v := range saved {
				if v.Title == title {
					saved[i] = Recipe{
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
					break
				}
			}
			data, _ := json.MarshalIndent(saved, "", "  ")
			_ = os.WriteFile("resep.json", data, 0644)
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
		file, _ := os.ReadFile("resep.json")
		var saved []Recipe
		json.Unmarshal(file, &saved)
		var updated []Recipe
		for _, v := range saved {
			if v.Title != title {
				updated = append(updated, v)
			}
		}
		data, _ := json.MarshalIndent(updated, "", "  ")
		_ = os.WriteFile("resep.json", data, 0644)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// DETEKSI PORT DARI SYSTEM 
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port:", port)
	
	// Jalankan ListenAndServe dengan port dinamis
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
