package models

import "time"

// Article — заявка на публикацию статьи.
// file_path — путь к загруженному файлу на сервере.
type Article struct {
	ID       int    `json:"id"`
	Author   string `json:"author"`
	Title    string `json:"title"`
	Email    string `json:"email"`
	FilePath string `json:"file_path"`
}

type ArticleRow struct {
	ID        int        `json:"id"`
	Author    string     `json:"author"`
	Title     string     `json:"title"`
	Email     string     `json:"email"`
	FilePath  string     `json:"file_path"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}
