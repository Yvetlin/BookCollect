package handlers

import (
	"BookCollect/internal/db"
	"BookCollect/internal/sessions"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ShowLoginPage отображает страницу входа администратора
func ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title": "Вход администратора",
		"Year":  time.Now().Year(),
	}
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		data["Error"] = errMsg
	}

	tmpl, err := template.ParseFiles(
		"web/templates/base.html",
		"web/templates/admin/login.html",
	)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		return
	}
	_ = tmpl.ExecuteTemplate(w, "base", data)
}

// HandleLogin обрабатывает POST-запрос входа администратора
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/login?error=Ошибка формы", http.StatusFound)
		return
	}

	if db.DB == nil {
		http.Error(w, "БД не инициализирована (db.DB == nil)", http.StatusInternalServerError)
		return
	}

	login := strings.TrimSpace(r.FormValue("login"))
	password := r.FormValue("password")
	if login == "" || password == "" {
		http.Redirect(w, r, "/admin/login?error=Заполните все поля", http.StatusFound)
		return
	}

	var id int
	var passwordHash string
	err := db.DB.QueryRow(`SELECT id, password_hash FROM administrators WHERE login = $1`, login).
		Scan(&id, &passwordHash)
	if err == sql.ErrNoRows {
		http.Redirect(w, r, "/admin/login?error=Неверный логин или пароль", http.StatusFound)
		return
	} else if err != nil {
		http.Redirect(w, r, "/admin/login?error=Ошибка БД", http.StatusFound)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil {
		http.Redirect(w, r, "/admin/login?error=Неверный логин или пароль", http.StatusFound)
		return
	}

	if err := sessions.SetAdminID(w, r, id); err != nil {
		log.Println("session save error: %v", err)
		http.Redirect(w, r, "/admin/login?error=Ошибка сессии", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/admin/panel/collections", http.StatusFound)
}

// HandleLogout удаляет сессию и возвращает на логин
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	if err := sessions.ClearAdminID(w, r); err != nil {
		http.Error(w, "Ошибка выхода", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/login", http.StatusFound)
}
