package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/file"
	"github.com/lin-snow/ech0/internal/storage"
	errorUtil "github.com/lin-snow/ech0/internal/util/err"
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
		userID := ctx.MustGet("userid").(string)

		fileDto, err := fileHandler.fileService.UploadFile(userID, file, category, storageType)
		if err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Data: fileDto, Msg: commonModel.UPLOAD_SUCCESS}
	})
}

func (fileHandler *FileHandler) CreateExternalFile() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userID := ctx.MustGet("userid").(string)
		var dto commonModel.CreateExternalFileDto
		if err := ctx.ShouldBindJSON(&dto); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		fileDto, err := fileHandler.fileService.CreateExternalFile(userID, dto)
		if err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Data: fileDto, Msg: commonModel.UPLOAD_SUCCESS}
	})
}

func (fileHandler *FileHandler) DeleteFile() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userID := ctx.MustGet("userid").(string)
		var dto commonModel.FileDeleteDto
		if err := ctx.ShouldBindJSON(&dto); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		if err := fileHandler.fileService.DeleteFile(userID, dto); err != nil {
			ctx.JSON(
				http.StatusOK,
				commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
					Msg: "",
					Err: err,
				})),
			)
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Msg: commonModel.DELETE_SUCCESS}
	})
}

func (fileHandler *FileHandler) UploadAudioFile() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userID := ctx.MustGet("userid").(string)
		file, err := ctx.FormFile("file")
		if err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		audioFile, err := fileHandler.fileService.UploadAudioFile(userID, file)
		if err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Data: audioFile, Msg: commonModel.UPLOAD_SUCCESS}
	})
}

func (fileHandler *FileHandler) DeleteAudioFile() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userID := ctx.MustGet("userid").(string)
		if err := fileHandler.fileService.DeleteAudioFile(userID); err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Msg: commonModel.DELETE_SUCCESS}
	})
}

func (fileHandler *FileHandler) GetCurrentAudio() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		audioURL := fileHandler.fileService.GetCurrentAudioURL()
		return res.Response{Data: audioURL, Msg: commonModel.GET_MUSIC_URL_SUCCESS}
	})
}

func (fileHandler *FileHandler) StreamCurrentAudio(ctx *gin.Context) {
	fileHandler.fileService.StreamCurrentAudio(ctx)
}

func (fileHandler *FileHandler) GetFilePresignURL() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userID := ctx.MustGet("userid").(string)
		var dto commonModel.GetPresignURLDto
		if err := ctx.ShouldBindJSON(&dto); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		presignDto, err := fileHandler.fileService.GetFilePresignURL(userID, &dto)
		if err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Data: presignDto, Msg: commonModel.GET_S3_PRESIGN_URL_SUCCESS}
	})
}
