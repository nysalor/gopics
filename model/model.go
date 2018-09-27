package model

type Album struct {
	Id          int64     `db:"id"`
	UpdatedAt   string    `db:"updated_at"`
	CreatedAt   string    `db:"created_at"`
	Name        string    `db:"name"`
	DirName     string    `db:"dirname"`
	Description string    `db:"description"`
	ImagesCount int       `db:"images_count"`
	Cover       string    `db:"cover"`
	Images      []Image
}

type Image struct {
	Id          int64     `db:"id"`
	UpdatedAt   string    `db:"updated_at"`
	CreatedAt   string    `db:"created_at"`
	AlbumId     int64     `db:"album_id"`
	Filename    string    `db:"filename"`
	Exif
}

type Exif struct {
	Maker       string    `db:"maker"`
	Model       string    `db:"model"`
	LensMaker   string    `db:"lens_maker"`
	LensModel   string    `db:"lens_model"`
	TookAt      string    `db:"took_at"`
	FNumber     string    `db:"f_number"`
	FocalLength string    `db:"focal_length"`
	Iso         string    `db:"iso"`
	Latitude    float64   `db:"latitude"`
	Longitude   float64   `db:"longitude"`
}
