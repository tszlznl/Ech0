package services

import (
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/config"
	"github.com/lin-snow/ech0/internal/dto"
	"github.com/lin-snow/ech0/pkg"
)

// UploadImage 上传图片
func UploadImage(c *gin.Context) dto.Result[string] {
	// 从配置中读取支持的扩展名
	allowedExtensions := config.Config.Upload.AllowedTypes

	// 调用 pkg 中的图片上传方法
	imageURL, err := pkg.UploadImage(c, allowedExtensions)
	if err != nil {
		return dto.Fail[string](err.Error())
	}

	return dto.OK(imageURL)
}
