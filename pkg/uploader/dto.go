package uploader

type GetPresignedUrlRequest struct {
	Extension string `form:"extension" binding:"required"`
}

type GetPresignedDownloadRequest struct {
	ObjectKeyBase64 string `form:"objectKeyBase64" binding:"required"`
}

type UploadedFileDto struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type UploadedFilesDto struct {
	UploadedFiles []UploadedFileDto `json:"uploadedFiles"`
}

type PresignedUpload struct {
	ObjectKey string `json:"objectKey"`
	Url       string `json:"url"`
}

type PresignedDownload struct {
	Url string `json:"url"`
}
