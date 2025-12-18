package file

import (
	"context"
	"mime/multipart"

	"backend/app/types/dto"
	fileErr "backend/app/types/errorn"
	"backend/utils/bind"
	"backend/utils/handle"
	"backend/utils/logs"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type FileLogic interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader) (*dto.FileDTO, error)
}

type FileHandlerParams struct {
	fx.In

	FileLogic FileLogic
}

type FileHandler struct {
	fileLogic FileLogic
}

func NewFileHandler(params FileHandlerParams) *FileHandler {
	return &FileHandler{
		fileLogic: params.FileLogic,
	}
}

var fileBindConfig = bind.FieldErrorConfig{
	InvalidParamCode: fileErr.FileErrInvalidFile,
	RequiredCode:     fileErr.FileErrInvalidFile,
	FieldLabels: map[string]string{
		"file": "文件",
	},
}

// UploadFile 上传文件
// @Summary 上传文件
// @Description 上传文件到服务器，支持多种文件类型
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "要上传的文件"
// @Success 200 {object} handle.Response{data=UploadFileResp} "成功"
// @Failure 400 {object} handle.Response "请求参数错误"
// @Failure 500 {object} handle.Response "服务器内部错误"
// @Router /api/file/upload [post]
func (h *FileHandler) UploadFile(c *gin.Context) {
	ctx := c.Request.Context()

	var req UploadFileReq
	if err := bind.ShouldBind(c, &req, fileBindConfig); err != nil {
		handle.HandleErrorWithContext(c, err, "上传文件", nil)
		return
	}

	fileRecord, err := h.fileLogic.UploadFile(ctx, req.File)
	if err != nil {
		handle.HandleErrorWithContext(c, err, "上传文件", nil)
		return
	}

	logs.CtxInfof(ctx, "文件上传成功: file_id=%d, filename=%s", fileRecord.FileID, fileRecord.FileName)

	// 构建响应
	resp := UploadFileResp{
		FileID:   fileRecord.FileID,
		FileName: fileRecord.FileName,
		FileURL:  fileRecord.FileURL,
	}
	handle.Success(c, resp)
}
