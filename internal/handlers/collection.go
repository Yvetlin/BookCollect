package handlers

import (
	"BookCollect/internal/db"
	"BookCollect/internal/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// ---------- PUBLIC API (JSON) ----------

func GetCollections(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT id, release_number, release_year, title, description, cover_image, publication_link, pdf_path
		FROM collections
		ORDER BY id DESC`)
	if err != nil {
		http.Error(w, "Ошибка запроса: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var out []models.CollectionResponse
	for rows.Next() {
		var c models.Collection
		if err := rows.Scan(
			&c.ID,
			&c.ReleaseNumber,
			&c.ReleaseYear,
			&c.Title,
			&c.Description,
			&c.CoverImage,
			&c.PublicationLink,
			&c.PDFPath,
		); err != nil {
			http.Error(w, "Ошибка чтения строк: "+err.Error(), http.StatusInternalServerError)
			return
		}
		out = append(out, models.CollectionToResponse(c))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func GetCollectionByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	row := db.DB.QueryRow(`
		SELECT id, release_number, release_year, title, description, cover_image, publication_link, pdf_path
		FROM collections
		WHERE id = $1`, id)

	var c models.Collection
	if err := row.Scan(
		&c.ID, &c.ReleaseNumber, &c.ReleaseYear, &c.Title, &c.Description,
		&c.CoverImage, &c.PublicationLink, &c.PDFPath,
	); err == sql.ErrNoRows {
		http.Error(w, "Сборник не найден", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Ошибка запроса: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(models.CollectionToResponse(c))
}

// ---------- ADMIN (multipart create; JSON update/delete) ----------

func CreateCollection(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Ошибка парсинга формы: "+err.Error(), http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Поле 'title' обязательно", http.StatusBadRequest)
		return
	}
	description := r.FormValue("description")
	releaseYearStr := r.FormValue("release_year")
	releaseNumberStr := r.FormValue("release_number")
	publicationLink := r.FormValue("publication_link")

	var releaseYear, releaseNumber *int32
	if releaseYearStr != "" {
		if v, err := strconv.Atoi(releaseYearStr); err == nil {
			tmp := int32(v)
			releaseYear = &tmp
		}
	}
	if releaseNumberStr != "" {
		if v, err := strconv.Atoi(releaseNumberStr); err == nil {
			tmp := int32(v)
			releaseNumber = &tmp
		}
	}

	// ensure dirs
	_ = os.MkdirAll("uploads/covers", 0755)
	_ = os.MkdirAll("uploads/pdfs", 0755)

	// cover (optional)
	coverPath := ""
	if file, hdr, err := r.FormFile("cover"); err == nil {
		defer file.Close()
		dstPath := filepath.Join("uploads", "covers", filepath.Base(hdr.Filename))
		if dst, err := os.Create(dstPath); err == nil {
			defer dst.Close()
			_, _ = io.Copy(dst, file)
			coverPath = "/" + filepath.ToSlash(dstPath) // чтобы открыть через /uploads/...
		}
	}

	// pdf (optional)
	pdfPath := ""
	if file, hdr, err := r.FormFile("pdf"); err == nil {
		defer file.Close()
		dstPath := filepath.Join("uploads", "pdfs", filepath.Base(hdr.Filename))
		if dst, err := os.Create(dstPath); err == nil {
			defer dst.Close()
			_, _ = io.Copy(dst, file)
			pdfPath = "/" + filepath.ToSlash(dstPath)
		}
	}

	var id int
	if err := db.DB.QueryRow(`
		INSERT INTO collections (release_number, release_year, title, description, cover_image, publication_link, pdf_path)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		releaseNumber, releaseYear, title, description, coverPath, publicationLink, pdfPath,
	).Scan(&id); err != nil {
		http.Error(w, "Ошибка вставки: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "Сборник создан",
		"id":      id,
	})
}

func UpdateCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	var in models.CollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "Неверный JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := db.DB.Exec(`
		UPDATE collections SET
			release_number = $1,
			release_year = $2,
			title = $3,
			description = $4,
			cover_image = $5,
			publication_link = $6,
			pdf_path = $7
		WHERE id = $8`,
		in.ReleaseNumber, in.ReleaseYear, in.Title, in.Description,
		in.CoverImage, in.PublicationLink, in.PDFPath, id,
	); err != nil {
		http.Error(w, "Ошибка обновления: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Сборник с ID %d обновлён", id),
	})
}

func DeleteCollection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	res, err := db.DB.Exec(`DELETE FROM collections WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "Ошибка удаления: "+err.Error(), http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, fmt.Sprintf("Сборник с ID %d не найден", id), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Сборник с ID %d удалён", id),
	})
}
