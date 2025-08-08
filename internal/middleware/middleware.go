package middleware

import (
	"BookCollect/internal/sessions"
	"net/http"
)

// Вариант 1: обёртка для конкретных хендлеров (оставляем как есть)
// Позволяет писать: r.Post("/path", middleware.AdminOnly(handler))
func AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := sessions.GetAdminID(r); !ok {
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}

// Вариант 2: chi-совместимая мидлварь
// Позволяет писать: g.Use(middleware.AdminOnlyMW)
func AdminOnlyMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := sessions.GetAdminID(r); !ok {
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
