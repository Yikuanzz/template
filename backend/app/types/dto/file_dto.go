package dto

type FileDTO struct {
	FileID   uint   `json:"file_id"`
	FileName string `json:"file_name"`
	FileURL  string `json:"file_url"`
}
