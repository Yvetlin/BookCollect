package sessions

import (
	"crypto/sha256"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

const sessionName = "admin_session"

func init() {
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		// в докере это может быть пусто — но без секрета работать нельзя
		secret = "dev-insecure-secret-change-me-now"
	}

	// Делаем 2 ключа: подпись + шифрование (устойчивее, чем только подпись).
	// Длины подходящие для securecookie.
	h := sha256.Sum256([]byte("auth:" + secret))
	e := sha256.Sum256([]byte("enc:" + secret))

	store = sessions.NewCookieStore(h[:], e[:])
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 дней
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,          // кука по GET тоже отправится
		Secure:   os.Getenv("APP_HTTPS") == "1", // локально 0, за HTTPS-прокси — 1
	}
}

func GetSession(r *http.Request) (*sessions.Session, error) {
	return store.Get(r, sessionName)
}

func SetAdminID(w http.ResponseWriter, r *http.Request, adminID int) error {
	s, err := GetSession(r)
	if err != nil {
		return err
	}
	s.Values["admin_id"] = adminID
	return s.Save(r, w) // выставит Set-Cookie
}

func GetAdminID(r *http.Request) (int, bool) {
	s, err := GetSession(r)
	if err != nil {
		return 0, false
	}
	if v, ok := s.Values["admin_id"].(int); ok {
		return v, true
	}
	return 0, false
}

func ClearAdminID(w http.ResponseWriter, r *http.Request) error {
	s, err := GetSession(r)
	if err != nil {
		return err
	}
	delete(s.Values, "admin_id")
	return s.Save(r, w)
}
