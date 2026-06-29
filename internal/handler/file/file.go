// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package handler 暴露文件相关的 HTTP 接口。
//
// JSON 端点（列表/树/元信息/删除/外链/预签名）走 Huma type-first；
// 二进制流式下载与 multipart 上传仍走裸 gin（见本文件下方 + setupFileRoutes）。
package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/handler/humares"
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

type (
	ListFilesInput struct {
		Page        int    `query:"page" doc:"页码"`
		PageSize    int    `query:"pageSize" doc:"每页数量"`
		Search      string `query:"search" doc:"搜索关键词"`
		StorageType string `query:"storage_type" doc:"存储类型"`
	}
	ListFileTreeInput struct {
		StorageType string `query:"storage_type" required:"true" doc:"存储类型"`
		Prefix      string `query:"prefix" doc:"路径前缀"`
	}
	GetFileByIDInput struct {
		ID string `path:"id" doc:"文件 ID"`
	}
	DeleteFileInput struct {
		ID string `path:"id" doc:"文件 ID"`
	}
	UpdateFileMetaInput struct {
		ID   string `path:"id" doc:"文件 ID"`
		Body commonModel.UpdateFileMetaDto
	}
	CreateExternalFileInput struct {
		Body commonModel.CreateExternalFileDto
	}
	GetFilePresignURLInput struct {
		Body commonModel.GetPresignURLDto
	}
)

// ListFiles 分页获取文件列表（file:read）。
func (fileHandler *FileHandler) ListFiles(ctx context.Context, in *ListFilesInput) (*humares.Envelope[commonModel.FileListResultDto], error) {
	query := commonModel.FileListQueryDto{Page: in.Page, PageSize: in.PageSize, Search: in.Search, StorageType: in.StorageType}
	result, err := fileHandler.fileService.ListFiles(ctx, query)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, result), nil
}

// ListFileTree 按存储类型/前缀返回文件树（file:read）。
func (fileHandler *FileHandler) ListFileTree(ctx context.Context, in *ListFileTreeInput) (*humares.Envelope[commonModel.FileTreeResultDto], error) {
	query := commonModel.FileTreeQueryDto{StorageType: in.StorageType, Prefix: in.Prefix}
	result, err := fileHandler.fileService.ListFileTree(ctx, query)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, result), nil
}

// GetFileByID 获取单个文件的元信息（file:read）。
func (fileHandler *FileHandler) GetFileByID(ctx context.Context, in *GetFileByIDInput) (*humares.Envelope[commonModel.FileDto], error) {
	fileDto, err := fileHandler.fileService.GetFileByID(ctx, in.ID)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, fileDto), nil
}

// UpdateFileMeta 更新对象存储文件元信息（file:write，用于预签名直传完成后回填 size/width/height）。
func (fileHandler *FileHandler) UpdateFileMeta(ctx context.Context, in *UpdateFileMetaInput) (*humares.Envelope[commonModel.FileDto], error) {
	fileDto, err := fileHandler.fileService.UpdateFileMeta(ctx, in.ID, in.Body)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, fileDto), nil
}

// CreateExternalFile 登记一个外链文件（file:write）。
func (fileHandler *FileHandler) CreateExternalFile(ctx context.Context, in *CreateExternalFileInput) (*humares.Envelope[commonModel.FileDto], error) {
	fileDto, err := fileHandler.fileService.CreateExternalFile(ctx, in.Body)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, fileDto, commonModel.UPLOAD_SUCCESS), nil
}

// DeleteFile 删除文件（file:write）。
func (fileHandler *FileHandler) DeleteFile(ctx context.Context, in *DeleteFileInput) (*humares.Envelope[any], error) {
	if err := fileHandler.fileService.DeleteFile(ctx, in.ID); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.DELETE_SUCCESS), nil
}

// GetFilePresignURL 获取对象存储直传预签名 URL（file:write）。
func (fileHandler *FileHandler) GetFilePresignURL(ctx context.Context, in *GetFilePresignURLInput) (*humares.Envelope[commonModel.PresignDto], error) {
	presignDto, err := fileHandler.fileService.GetFilePresignURL(ctx, &in.Body)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, presignDto, commonModel.GET_S3_PRESIGN_URL_SUCCESS), nil
}

// --- 以下为非 JSON 端点，仍走裸 gin（multipart 上传 / 二进制流式下载） ---

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

func (fileHandler *FileHandler) StreamFileByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.Status(400)
		return
	}
	fileHandler.fileService.StreamFileByID(ctx, id)
}

func (fileHandler *FileHandler) StreamFileByPath(ctx *gin.Context) {
	var query commonModel.FilePathStreamQueryDto
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.String(400, commonModel.INVALID_QUERY_PARAMS)
		return
	}
	fileHandler.fileService.StreamFileByPath(ctx, query)
}
