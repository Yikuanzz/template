package file

import "mime/multipart"

type UploadFileReq struct {
	File *multipart.FileHeader `form:"file" binding:"required" label:"文件" example:"file.jpg"`
}

// UploadFileResp 文件上传响应
type UploadFileResp struct {
	FileID   uint   `json:"file_id"`   // 文件ID
	FileName string `json:"file_name"` // 文件名
	FileURL  string `json:"file_url"`  // 文件访问URL
}
