package models

import "database/sql"

// Базовая сущность из таблицы collections
type Collection struct {
	ID              int            `json:"id"`
	ReleaseNumber   sql.NullInt32  `json:"release_number"`
	ReleaseYear     sql.NullInt32  `json:"release_year"`
	Title           string         `json:"title"`
	Description     *string        `json:"description"`      // может быть NULL
	CoverImage      sql.NullString `json:"cover_image"`      // путь/URL обложки (nullable)
	PublicationLink string         `json:"publication_link"` // может быть пустой строкой
	PDFPath         sql.NullString `json:"pdf_path"`         // путь к PDF (nullable)
}

// Удобный ответ наружу (JSON API), уже без sql.Null*
type CollectionResponse struct {
	ID              int     `json:"id"`
	ReleaseNumber   *int32  `json:"release_number,omitempty"`
	ReleaseYear     *int32  `json:"release_year,omitempty"`
	Title           string  `json:"title,omitempty"`
	Description     *string `json:"description,omitempty"`
	CoverImage      *string `json:"cover_image,omitempty"`
	PublicationLink string  `json:"publication_link,omitempty"`
	PDFPath         *string `json:"pdf_path,omitempty"`
}

// Запрос на создание/обновление через JSON API (PUT /admin/collection/{id})
type CollectionRequest struct {
	ID              int     `json:"id,omitempty"`
	ReleaseNumber   *int32  `json:"release_number,omitempty"`
	ReleaseYear     *int32  `json:"release_year,omitempty"`
	Title           string  `json:"title"`
	Description     *string `json:"description,omitempty"`
	CoverImage      *string `json:"cover_image,omitempty"`
	PublicationLink string  `json:"publication_link"`
	PDFPath         *string `json:"pdf_path"`
}

// Маппинг из базы (с Null*) в удобный API-ответ
func CollectionToResponse(c Collection) CollectionResponse {
	var releaseNumber *int32
	if c.ReleaseNumber.Valid {
		releaseNumber = &c.ReleaseNumber.Int32
	}

	var releaseYear *int32
	if c.ReleaseYear.Valid {
		releaseYear = &c.ReleaseYear.Int32
	}

	var coverImage *string
	if c.CoverImage.Valid {
		coverImage = &c.CoverImage.String
	}

	var pdfPath *string
	if c.PDFPath.Valid {
		pdfPath = &c.PDFPath.String
	}

	// Description уже *string — достаточно передать как есть
	return CollectionResponse{
		ID:              c.ID,
		ReleaseNumber:   releaseNumber,
		ReleaseYear:     releaseYear,
		Title:           c.Title,
		Description:     c.Description,
		CoverImage:      coverImage,
		PublicationLink: c.PublicationLink,
		PDFPath:         pdfPath,
	}
}
