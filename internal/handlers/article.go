package handlers

import (
	"BookCollect/internal/db"
	"BookCollect/internal/models"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var translitMap = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d",
	'е': "e", 'ё': "e", 'ж': "zh", 'з': "z", 'и': "i",
	'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n",
	'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t",
	'у': "u", 'ф': "f", 'х': "h", 'ц': "c", 'ч': "ch",
	'ш': "sh", 'щ': "shch", 'ъ': "", 'ы': "y", 'ь': "",
	'э': "e", 'ю': "yu", 'я': "ya",
}

func sanitizeFileName(input string) string {
	input = strings.ToLower(input)
	var b strings.Builder
	for _, r := range input {
		if val, ok := translitMap[r]; ok {
			b.WriteString(val)
		} else if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else if r == ' ' || r == '_' || r == '-' {
			b.WriteRune('_')
		}
	}
	safe := b.String()
	re := regexp.MustCompile(`[^a-z0-9_-]+`)
	return re.ReplaceAllString(safe, "")
}

// PUBLIC: подать заявку
var (
	maxUploadSize int64 = 25 << 20 // 25 MB
	allowedExt          = map[string]bool{".pdf": true, ".docx": true, ".odt": true}
	emailRe             = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
)

// AddArticle обрабатывает публичную подачу заявки со вложением
func AddArticle(w http.ResponseWriter, r *http.Request) {
	// Ограничиваем тело запроса и парсим multipart
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		jsonError(w, http.StatusRequestEntityTooLarge, "Слишком большой файл (лимит 25 МБ)")
		return
	}

	author := strings.TrimSpace(r.FormValue("author"))
	title := strings.TrimSpace(r.FormValue("title"))
	email := strings.TrimSpace(r.FormValue("email"))

	if author == "" || title == "" || email == "" {
		jsonError(w, http.StatusBadRequest, "Заполните все поля: Автор, Название, Email")
		return
	}
	if !emailRe.MatchString(email) {
		jsonError(w, http.StatusBadRequest, "Некорректный email")
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		jsonError(w, http.StatusBadRequest, "Приложите файл рукописи")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(handler.Filename))
	if !allowedExt[ext] {
		jsonError(w, http.StatusBadRequest, "Недопустимый тип файла. Разрешено: PDF, DOCX, ODT")
		return
	}

	// Дополнительная проверка размера, если доступен
	if handler.Size > 0 && handler.Size > maxUploadSize {
		jsonError(w, http.StatusRequestEntityTooLarge, "Слишком большой файл (лимит 25 МБ)")
		return
	}

	// Готовим директорию и имя файла
	if err := os.MkdirAll("uploads/articles", 0o755); err != nil {
		jsonError(w, http.StatusInternalServerError, "Не удалось подготовить хранилище файлов")
		return
	}
	base := safeBaseName(strings.TrimSpace(title))
	if base == "" {
		base = "article"
	}
	fileName := fmt.Sprintf("%s_%d%s", base, time.Now().Unix(), ext)
	dstPath := filepath.Join("uploads", "articles", fileName)

	// Пишем файл на диск
	dst, err := os.Create(dstPath)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Не удалось сохранить файл")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		jsonError(w, http.StatusInternalServerError, "Ошибка записи файла")
		return
	}

	// Пишем запись в БД
	var id int
	err = db.DB.QueryRow(
		`INSERT INTO articles (author, title, email, file_path) VALUES ($1,$2,$3,$4) RETURNING id`,
		author, title, email, "/"+filepath.ToSlash(dstPath),
	).Scan(&id)
	if err != nil {
		// При ошибке БД — удалим сохранённый файл, чтобы не копить мусор
		_ = os.Remove(dstPath)
		jsonError(w, http.StatusInternalServerError, "Ошибка БД при сохранении заявки")
		return
	}

	// Успех
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":  true,
		"id":  id,
		"msg": "Заявка отправлена",
	})
}

// safeBaseName — грубая нормализация имени файла (латиница/цифры/дефис/подчёркивание)
func safeBaseName(s string) string {
	// Убираем расширение, если прилетело целиком
	s = strings.TrimSuffix(s, filepath.Ext(s))
	// Приводим к нижнему регистру, заменяем пробелы
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "_")
	// Удаляем всё, кроме латиницы/цифр/_/-
	re := regexp.MustCompile(`[^a-z0-9_\-]+`)
	return strings.Trim(re.ReplaceAllString(s, ""), "_-")
}

func jsonError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": msg,
	})
}

// ADMIN: список заявок (JSON)
func GetArticle(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`SELECT id, author, title, email, file_path FROM articles ORDER BY id DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var a models.Article
		if err := rows.Scan(&a.ID, &a.Author, &a.Title, &a.Email, &a.FilePath); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		articles = append(articles, a)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(articles)
}

// ADMIN: заявка по id (JSON)
func GetArticleByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	var a models.Article
	if err := db.DB.QueryRow(`
		SELECT id, author, title, email, file_path 
		FROM articles WHERE id = $1`, id).
		Scan(&a.ID, &a.Author, &a.Title, &a.Email, &a.FilePath); err != nil {
		http.Error(w, "Не найдено: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(a)
}

// ADMIN: скачать файл заявки
func DownloadArticleFile(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	var filePath string
	if err := db.DB.QueryRow(`SELECT file_path FROM articles WHERE id = $1`, id).Scan(&filePath); err != nil {
		http.Error(w, "Не найдено: "+err.Error(), http.StatusNotFound)
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Файл недоступен: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(filePath)))
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = io.Copy(w, f)
}

// ADMIN: удалить заявку (+ её файл)
func DeleteArticle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	var filePath string
	if err := db.DB.QueryRow(`SELECT file_path FROM articles WHERE id = $1`, id).Scan(&filePath); err != nil {
		http.Error(w, "Статья не найдена", http.StatusNotFound)
		return
	}

	if _, err := db.DB.Exec(`DELETE FROM articles WHERE id = $1`, id); err != nil {
		http.Error(w, "Ошибка при удалении", http.StatusInternalServerError)
		return
	}

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		http.Error(w, "Файл не удалён", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Статья и файл удалены")
}

func GetArticles(w http.ResponseWriter, r *http.Request) {
	if db.DB == nil {
		http.Error(w, `{"error":"db not initialized"}`, http.StatusInternalServerError)
		return
	}

	rows, err := db.DB.Query(`
		SELECT id, author, title, email, file_path, created_at
		FROM articles
		ORDER BY id DESC`)
	if err != nil {
		http.Error(w, `{"error":"db query failed"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	list := make([]models.ArticleRow, 0, 64)
	for rows.Next() {
		var a models.ArticleRow
		if err := rows.Scan(&a.ID, &a.Author, &a.Title, &a.Email, &a.FilePath, &a.CreatedAt); err != nil {
			http.Error(w, `{"error":"db scan failed"}`, http.StatusInternalServerError)
			return
		}
		list = append(list, a)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, `{"error":"db rows error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// Отдаём просто массив — твой admin.js это понимает
	_ = json.NewEncoder(w).Encode(list)
}
