package handlers

import (
	"BookCollect/internal/db"
	"BookCollect/internal/sessions"
	"database/sql"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

/* ========= ВСПОМОГАТЕЛЬНОЕ ========= */

// Единый рендер: сам прокидывает .IsAdmin во все шаблоны
func render(w http.ResponseWriter, r *http.Request, files []string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	_, isAdmin := sessions.GetAdminID(r)
	data["IsAdmin"] = isAdmin

	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		http.Error(w, "Ошибка шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_ = tmpl.ExecuteTemplate(w, "base", data)
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
func firstN(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

/* ========= ПУБЛИЧНЫЕ СТРАНИЦЫ ========= */

func ShowIndexPage(w http.ResponseWriter, r *http.Request) {
	render(w, r,
		[]string{"web/templates/base.html", "web/templates/index.html"},
		map[string]any{
			"Title": "Главная",
			"Year":  time.Now().Year(),
		},
	)
}

type Collection struct {
	ID              int
	ReleaseNumber   *int
	ReleaseYear     *int
	Title           string
	Description     *string
	CoverImage      *string
	PublicationLink string
	PDFPath         *string
}

func ShowCollectionsPage(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT id, release_number, release_year, title, description, cover_image, publication_link, pdf_path
		FROM collections
		ORDER BY id DESC`)
	if err != nil {
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []Collection
	for rows.Next() {
		var c Collection
		if err := rows.Scan(
			&c.ID, &c.ReleaseNumber, &c.ReleaseYear, &c.Title, &c.Description,
			&c.CoverImage, &c.PublicationLink, &c.PDFPath,
		); err != nil {
			http.Error(w, "Ошибка чтения БД", http.StatusInternalServerError)
			return
		}
		// Нормализуем относительные пути
		if p := deref(c.PDFPath); p != "" && !strings.HasPrefix(p, "/") {
			pp := filepath.ToSlash("/" + p)
			c.PDFPath = &pp
		}
		if img := deref(c.CoverImage); img != "" && !strings.HasPrefix(img, "/") && !strings.HasPrefix(img, "http") {
			ii := filepath.ToSlash("/" + img)
			c.CoverImage = &ii
		}
		list = append(list, c)
	}

	render(w, r,
		[]string{"web/templates/base.html", "web/templates/collections.html"},
		map[string]any{
			"Title":       "Все сборники",
			"Year":        time.Now().Year(),
			"Collections": list,
		},
	)
}

func ShowCollectionPage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	row := db.DB.QueryRow(`
		SELECT id, release_number, release_year, title, description, cover_image, publication_link, pdf_path
		FROM collections WHERE id=$1`, id)

	var c Collection
	if err := row.Scan(
		&c.ID, &c.ReleaseNumber, &c.ReleaseYear, &c.Title, &c.Description,
		&c.CoverImage, &c.PublicationLink, &c.PDFPath,
	); err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}

	// нормализуем пути
	if p := deref(c.PDFPath); p != "" && !strings.HasPrefix(p, "/") {
		pp := filepath.ToSlash("/" + p)
		c.PDFPath = &pp
	}
	if img := deref(c.CoverImage); img != "" && !strings.HasPrefix(img, "/") && !strings.HasPrefix(img, "http") {
		ii := filepath.ToSlash("/" + img)
		c.CoverImage = &ii
	}

	meta := map[string]string{
		"description":    firstN(deref(c.Description), 180),
		"og:title":       c.Title,
		"og:description": firstN(deref(c.Description), 200),
	}
	if img := deref(c.CoverImage); img != "" {
		meta["og:image"] = img
		meta["twitter:card"] = "summary_large_image"
	}

	render(w, r,
		[]string{"web/templates/base.html", "web/templates/collection.html"},
		map[string]any{
			"Title":      c.Title,
			"Year":       time.Now().Year(),
			"Meta":       meta,
			"Collection": c,
		},
	)
}

func ShowArticleForm(w http.ResponseWriter, r *http.Request) {
	render(w, r,
		[]string{"web/templates/base.html", "web/templates/article_form.html"},
		map[string]any{
			"Title": "Подать статью",
			"Year":  time.Now().Year(),
		},
	)
}

/* ========= АДМИН UI (НЕ API) ========= */

func AdminCollectionsPage(w http.ResponseWriter, r *http.Request) {
	render(w, r,
		[]string{"web/templates/base.html", "web/templates/admin/collections.html"},
		map[string]any{
			"Title": "Админ · Сборники",
			"Year":  time.Now().Year(),
		},
	)
}

func AdminArticlesPage(w http.ResponseWriter, r *http.Request) {
	render(w, r,
		[]string{"web/templates/base.html", "web/templates/admin/articles.html"},
		map[string]any{
			"Title": "Админ · Заявки",
			"Year":  time.Now().Year(),
		},
	)
}
