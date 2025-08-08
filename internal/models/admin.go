package models

// Administrator — запись из таблицы administrators.
// Пароль хранится в виде bcrypt-хэша в поле password_hash (в БД).
type Administrator struct {
	ID       int
	Login    string
	Password string // не используем напрямую; тут для совместимости, обычно держим только хэш в БД
}
