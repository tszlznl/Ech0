package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/file"
	"github.com/lin-snow/ech0/internal/storage"
)

type FileHandler struct {
	fileService service.Service
}

func NewFileHandler(fileService service.Service) *FileHandler {
	return &FileHandler{fileService: fileService}
}

func (fileHandler *FileHandler) UploadFile() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		file, err := ctx.FormFile("file")
		if err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		category := storage.NormalizeCategory(ctx.PostForm("category"))
		storageType := storage.NormalizeStorageType(ctx.PostForm("storage_type"))
		fileDto, err := fileHandler.fileService.UploadFile(ctx.Request.Context(), file, category, storageType)
		if err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Data: fileDto, Msg: commonModel.UPLOAD_SUCCESS}
	})
}

func (fileHandler *FileHandler) CreateExternalFile() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var dto commonModel.CreateExternalFileDto
		if err := ctx.ShouldBindJSON(&dto); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		fileDto, err := fileHandler.fileService.CreateExternalFile(ctx.Request.Context(), dto)
		if err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Data: fileDto, Msg: commonModel.UPLOAD_SUCCESS}
	})
}

func (fileHandler *FileHandler) DeleteFile() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		id := ctx.Param("id")
		if id == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS, Err: errors.New(commonModel.INVALID_PARAMS)}
		}

		if err := fileHandler.fileService.DeleteFile(ctx.Request.Context(), id); err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Msg: commonModel.DELETE_SUCCESS}
	})
}

func (fileHandler *FileHandler) GetFileByID() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		id := ctx.Param("id")
		if id == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS, Err: errors.New(commonModel.INVALID_PARAMS)}
		}

		fileDto, err := fileHandler.fileService.GetFileByID(ctx.Request.Context(), id)
		if err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Data: fileDto}
	})
}

func (fileHandler *FileHandler) StreamFileByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.Status(400)
		return
	}
	fileHandler.fileService.StreamFileByID(ctx, id)
}

func (fileHandler *FileHandler) GetFilePresignURL() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var dto commonModel.GetPresignURLDto
		if err := ctx.ShouldBindJSON(&dto); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		presignDto, err := fileHandler.fileService.GetFilePresignURL(ctx.Request.Context(), &dto)
		if err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Data: presignDto, Msg: commonModel.GET_S3_PRESIGN_URL_SUCCESS}
	})
}
