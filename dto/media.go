package dto

// PhotoOrderRequest adalah body PATCH .../photos/order untuk mengubah urutan foto.
type PhotoOrderRequest struct {
	PhotoIDs []string `json:"photo_ids" validate:"required,min=1,dive,required"`
}

// UploadPhotoResponse adalah respons unggah satu atau beberapa foto.
type UploadPhotoResponse struct {
	PropertyID string      `json:"property_id"`
	Uploaded   []PhotoView `json:"uploaded"`
	Total      int         `json:"total"`
	Limit      int         `json:"limit"`
}
