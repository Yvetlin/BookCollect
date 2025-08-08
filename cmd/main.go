package main

import (
	"BookCollect/internal/db"
	"BookCollect/internal/handlers"
	mw "BookCollect/internal/middleware"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	log.Println("Boot: calling db.InitDB()")
	db.InitDB()

	r := chi.NewRouter()

	// базовые middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.RedirectSlashes) // /path/ -> /path

	// статика
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	r.Handle("/images/*", http.StripPrefix("/images/", http.FileServer(http.Dir("web/images"))))
	// Если понадобится — раскомментируй раздачу загруженных файлов:
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// ---------- Публичные HTML-страницы ----------
	r.Get("/", handlers.ShowIndexPage)
	r.Get("/collections", handlers.ShowCollectionsPage)
	r.Get("/collections/{id}", handlers.ShowCollectionPage)

	// Подача статьи (форма + приём)
	r.Get("/article", handlers.ShowArticleForm)
	r.Post("/article", handlers.AddArticle)

	// ---------- Аутентификация администратора ----------
	r.Get("/admin/login", handlers.ShowLoginPage)
	r.Post("/admin/login", handlers.HandleLogin)
	r.Post("/admin/logout", handlers.HandleLogout)

	// ---------- Админ-панель (UI-страницы) ----------
	r.Group(func(g chi.Router) {
		g.Use(mw.AdminOnlyMW) // доступ только с валидной сессией

		g.Get("/admin/panel/collections", handlers.AdminCollectionsPage)
		g.Get("/admin/panel/articles", handlers.AdminArticlesPage)
	})

	// ---------- Публичное JSON API для сборников ----------
	r.Get("/api/collections", handlers.GetCollections)
	r.Get("/api/collections/{id}", handlers.GetCollectionByID)

	// ---------- Админ API для сборников ----------
	// create
	r.Post("/admin/collection", mw.AdminOnly(handlers.CreateCollection))
	// update — поддерживаем и PUT, и POST с _method=PUT (как делает твой admin.js)
	r.Put("/admin/collection/{id}", mw.AdminOnly(handlers.UpdateCollection))
	r.Post("/admin/collection/{id}", mw.AdminOnly(handlers.UpdateCollection))
	// delete
	r.Delete("/admin/collection/{id}", mw.AdminOnly(handlers.DeleteCollection))

	// ---------- Админ API для заявок (статей) ----------
	r.Get("/admin/articles", mw.AdminOnly(handlers.GetArticles))
	r.Get("/admin/articles/{id}", mw.AdminOnly(handlers.GetArticleByID))
	r.Delete("/admin/articles/{id}", mw.AdminOnly(handlers.DeleteArticle))
	r.Get("/admin/articles/{id}/download", mw.AdminOnly(handlers.DownloadArticleFile))

	// ---------- Старт сервера ----------
	host := getenv("HOST", "127.0.0.1")
	addr := host + ":" + getenv("PORT", "8080")
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
