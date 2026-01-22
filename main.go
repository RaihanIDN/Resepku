package main

import (
	"encoding/json"
	"html/template"
	"io" 
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Recipe: Blueprint data resep. Field-field ini yang bakal muncul di detail.html dan resep.json
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

// allRecipes: Database sementara. Isinya resep default yang muncul pertama kali sebelum user nambahin data
var allRecipes = []Recipe{
	{Title: "Keripik Kentang", Desc: "Renyah disetiap gigitan, gurihnya baluran bumbu. Keripik Kentang bikin nagih", Time: "20-30 Menit", Image: "Keripik_kentang.png"},
	{Title: "Kentang Tumbuk", Desc: "Selembut awan, semewah mentega. Rasakan kehangatan Mashed Potato yang meleleh di lidah.", Time: "20-30 Menit", Image: "Kentang_tumbuk.png"},
	{Title: "Kentang Balado", Desc: "Warnanya yang merah menggoda adalah janji rasa yang luar biasa. Kentang Balado yang tak hanya cantik, tapi juga lezat!", Time: "1 Jam +", Image: "Kentang_balado.png"},
	{Title: "Telur Dadar Padang", Desc: "Bukan telur dadar biasa. Rasakan kekayaan bumbu yang terkunci di setiap lapisan telur dadar setebal bantal ini.", Time: "20-30 Menit", Image: "Telur_dadar_padang.png"},
	{Title: "Telur Balado", Desc: "Rasa klasik yang selalu dicintai. Telur Balado kami dibuat dengan bumbu otentik dan cinta, menghasilkan kelezatan yang tiada duanya.", Time: "1 Jam +", Image: "Telur_Balado.png"},
	{Title: "Omelete Mie", Desc: "Kreasi paling jenius dari anak kos. Omelet Mie ini membuktikan bahwa makanan enak tidak harus rumit!", Time: "15-20 Menit", Image: "Omelet_mie.png"},
	{Title: "Nasi Goreng Seafood", Desc: "Rasa yang medok dan aroma yang menggoda! Nasi Goreng Seafood yang dimasak dengan bumbu rahasia dan teknik api yang sempurna.", Time: "30-40 Menit", Image: "Nasigoreng_Seafood.png"},
	{Title: "Nasi Goreng", Desc: "Comfort food sejati. Nikmati kehangatan dan kelezatan yang tiada tara dari seporsi Nasi Goreng.", Time: "20-30 Menit", Image: "Nasi_goreng.png"},
	{Title: "Nasi Goreng Cabe Hijau", Desc: "Pedasnya cabai hijau lebih light namun lebih aromatik. Inilah nasi goreng favorit bagi yang mencari sensasi pedas yang unik.", Time: "15-20 Menit", Image: "Nasigoreng_cabeijo.png"},
}

// getCombinedRecipes: Menggabungkan data hardcoded (allRecipes) dengan data dinamis dari resep.json
func getCombinedRecipes() []Recipe {
	file, err := os.ReadFile("resep.json")
	if err != nil || len(file) == 0 {
		return allRecipes
	}
	var fromFile []Recipe
	json.Unmarshal(file, &fromFile)
	return append(allRecipes, fromFile...)
}

// saveRecipeToFile: Logic buat nyimpen resep baru ke file resep.json supaya data gak ilang pas server mati
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
	// http.FileServer: Biar browser bisa manggil gambar/CSS dari folder 'static'
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route Utama (/): Nampilin daftar resep, search, dan pagination (halaman 1, 2, 3)
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

	// Route Detail (/recipe): Nyari satu resep spesifik pake parameter 'title'
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
			http.Error(w, "File detail.html gak ketemu bray!", http.StatusNotFound)
			return
		}
		tmpl.Execute(w, selectedRecipe)
	})

	// Route Tambah (/add-recipe): Nampilin form (GET) dan nerima inputan resep baru (POST)
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

	// Route Edit: Fitur moderasi admin pake 'key'
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

	// Route Simpan Makanan: Nampilin halaman 'simpan.html'
	http.HandleFunc("/Simpan-Makanan", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("simpan.html")
		if err != nil {
			http.Error(w, "File simpan.html gak ketemu bray!", http.StatusNotFound)
			return
		}
		tmpl.Execute(w, nil)
	})

	// Route Favorit Resepku: Nampilin halaman 'favorit.html'
	http.HandleFunc("/favorit", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("favorit.html")
		if err != nil {
			http.Error(w, "File favorit.html gak ketemu bray!", http.StatusNotFound)
			return
		}
		tmpl.Execute(w, nil)
	})

	// Route Detail Simpan (Baru): Detail buat resep favorit dari browser
	http.HandleFunc("/detailsimpan", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("detailsimpan.html")
		if err != nil {
			http.Error(w, "File detailsimpan.html gak ketemu bray!", http.StatusNotFound)
			return
		}
		tmpl.Execute(w, nil)
	})

	// Route Delete: Hapus resep dari file JSON
	http.HandleFunc("/moderasi/delete", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		title := r.URL.Query().Get("title")
		if key != "kuncirahasia77" {
			http.Error(w, "Akses ditolak bray!", http.StatusUnauthorized)
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

	println("Server running: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}