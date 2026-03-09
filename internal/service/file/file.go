package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	virefs "github.com/lin-snow/VireFS"
	"github.com/lin-snow/ech0/internal/config"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	publisher "github.com/lin-snow/ech0/internal/event/publisher"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/transaction"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	imgUtil "github.com/lin-snow/ech0/internal/util/img"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/lin-snow/ech0/pkg/viewer"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	globalAudioFileIDKey = "global_audio_file_id"
	externalKeyPrefix    = "external/"
	externalDefaultName  = "external"
)

type FileService struct {
	transactor         transaction.Transactor
	commonRepository   CommonRepository
	storageManager     *storage.Manager
	keyvalueRepository KeyValueRepository
	fileRepository     FileRepository
	publisher          *publisher.Publisher
	keyGen             storage.KeyGenerator
}

func NewFileService(
	tx transaction.Transactor,
	commonRepository CommonRepository,
	kvRepo KeyValueRepository,
	fileRepo FileRepository,
	storageManager *storage.Manager,
	publisher *publisher.Publisher,
) *FileService {
	return &FileService{
		transactor:         tx,
		commonRepository:   commonRepository,
		keyvalueRepository: kvRepo,
		fileRepository:     fileRepo,
		storageManager:     storageManager,
		publisher:          publisher,
		keyGen:             storage.NewRandomKeyGenerator(),
	}
}

func (s *FileService) UploadFile(
	ctx context.Context,
	file *multipart.FileHeader,
	category storage.Category,
	storageType storage.StorageType,
) (commonModel.FileDto, error) {
	userID := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonRepository.GetUserByUserId(context.Background(), userID)
	if err != nil {
		return commonModel.FileDto{}, err
	}
	if !user.IsAdmin {
		return commonModel.FileDto{}, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	contentType := file.Header.Get("Content-Type")
	if !isAllowedType(contentType, config.Config().Upload.AllowedTypes) {
		return commonModel.FileDto{}, errors.New(commonModel.FILE_TYPE_NOT_ALLOWED)
	}

	limit := int64(config.Config().Upload.ImageMaxSize)
	if category == storage.CategoryAudio {
		limit = int64(config.Config().Upload.AudioMaxSize)
	}
	if file.Size > limit {
		return commonModel.FileDto{}, errors.New(commonModel.FILE_SIZE_EXCEED_LIMIT)
	}

	gen := s.keyGenForCategory(category, file.Filename)
	key, err := gen.GenerateKey(category, user.ID, file.Filename)
	if err != nil {
		return commonModel.FileDto{}, err
	}

	reader, err := file.Open()
	if err != nil {
		return commonModel.FileDto{}, err
	}
	defer func() { _ = reader.Close() }()

	var opts []virefs.PutOption
	if contentType != "" {
		opts = append(opts, virefs.WithContentType(contentType))
	}

	targetStorageType := storage.NormalizeStorageType(string(storageType))
	if targetStorageType == storage.StorageTypeExternal {
		targetStorageType = storage.StorageTypeLocal
	}
	selector := s.getSelector()
	if err := selector.Put(context.Background(), targetStorageType, key, reader, opts...); err != nil {
		return commonModel.FileDto{}, err
	}

	width, height := 0, 0
	if category.IsImageLike() {
		width, height, err = imgUtil.GetImageSizeFromFile(file)
		if err != nil {
			return commonModel.FileDto{}, err
		}
	}

	fileURL := selector.ResolveURL(targetStorageType, key)
	routeStorageType, provider, bucket := currentStorageRoute(selector, targetStorageType)

	fileRecord := &fileModel.File{
		Key:         key,
		StorageType: routeStorageType,
		Provider:    provider,
		Bucket:      bucket,
		URL:         fileURL,
		Name:        file.Filename,
		ContentType: contentType,
		Size:        file.Size,
		Category:    string(category),
		Width:       width,
		Height:      height,
		UserID:      user.ID,
	}
	if err := s.fileRepository.Create(context.Background(), fileRecord); err != nil {
		return commonModel.FileDto{}, err
	}

	uploadType := commonModel.ImageType
	if category == storage.CategoryAudio {
		uploadType = commonModel.AudioType
	}

	user.Password = ""
	if err := s.publisher.ResourceUploaded(
		context.Background(),
		contracts.ResourceUploadedEvent{
			User:     user,
			FileName: file.Filename,
			URL:      fileURL,
			Size:     file.Size,
			Type:     string(uploadType),
		},
		key,
	); err != nil {
		logUtil.GetLogger().Error("Failed to publish resource uploaded event", zap.String("error", err.Error()))
	}

	return commonModel.FileDto{
		ID:          fileRecord.ID,
		Key:         key,
		URL:         fileURL,
		ContentType: contentType,
		Category:    string(category),
		Size:        file.Size,
		Width:       width,
		Height:      height,
	}, nil
}

func (s *FileService) CreateExternalFile(
	ctx context.Context,
	dto commonModel.CreateExternalFileDto,
) (commonModel.FileDto, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonRepository.GetUserByUserId(context.Background(), userid)
	if err != nil {
		return commonModel.FileDto{}, err
	}
	if !user.IsAdmin {
		return commonModel.FileDto{}, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	rawURL := httpUtil.TrimURL(dto.URL)
	if rawURL == "" {
		return commonModel.FileDto{}, errors.New(commonModel.INVALID_PARAMS)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed == nil {
		return commonModel.FileDto{}, errors.New(commonModel.INVALID_PARAMS)
	}
	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if scheme != "http" && scheme != "https" {
		return commonModel.FileDto{}, errors.New(commonModel.INVALID_PARAMS)
	}

	contentType := strings.TrimSpace(dto.ContentType)
	if contentType == "" {
		contentType = httpUtil.GetMIMETypeFromFilenameOrURL(rawURL)
	}

	category := storage.NormalizeCategory(dto.Category)
	if category == "" {
		if strings.HasPrefix(contentType, "audio/") {
			category = storage.CategoryAudio
		} else {
			category = storage.CategoryImage
		}
	}

	normalizedURL := parsed.String()
	hash := sha256.Sum256([]byte(normalizedURL))
	key := externalKeyPrefix + string(category) + "/" + hex.EncodeToString(hash[:])

	const (
		externalStorageType = string(storage.StorageTypeExternal)
		externalProvider    = string(storage.StorageTypeExternal)
		externalBucket      = ""
	)
	existing, err := s.fileRepository.GetByRoute(
		context.Background(),
		externalStorageType,
		externalProvider,
		externalBucket,
		key,
	)
	if err == nil && existing != nil {
		return commonModel.FileDto{
			ID:          existing.ID,
			Key:         existing.Key,
			URL:         existing.URL,
			ContentType: existing.ContentType,
			Category:    existing.Category,
			Size:        existing.Size,
			Width:       existing.Width,
			Height:      existing.Height,
		}, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return commonModel.FileDto{}, err
	}

	name := strings.TrimSpace(dto.Name)
	if name == "" {
		name = filepath.Base(parsed.Path)
		if name == "" || name == "." || name == "/" {
			name = externalDefaultName
		}
	}

	fileRecord := &fileModel.File{
		Key:         key,
		StorageType: externalStorageType,
		Provider:    externalProvider,
		Bucket:      externalBucket,
		URL:         normalizedURL,
		Name:        name,
		ContentType: contentType,
		Size:        0,
		Category:    string(category),
		Width:       dto.Width,
		Height:      dto.Height,
		UserID:      user.ID,
	}
	if err := s.fileRepository.Create(context.Background(), fileRecord); err != nil {
		return commonModel.FileDto{}, err
	}

	return commonModel.FileDto{
		ID:          fileRecord.ID,
		Key:         key,
		URL:         normalizedURL,
		ContentType: contentType,
		Category:    string(category),
		Size:        0,
		Width:       fileRecord.Width,
		Height:      fileRecord.Height,
	}, nil
}

func (s *FileService) DeleteFile(ctx context.Context, dto commonModel.FileDeleteDto) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonRepository.GetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}
	if dto.Key == "" {
		return errors.New(commonModel.IMAGE_NOT_FOUND)
	}

	fileRecord, err := s.fileRepository.GetByKey(context.Background(), dto.Key)
	if err != nil {
		return err
	}

	if err := s.transactor.Run(context.Background(), func(ctx context.Context) error {
		return s.DeleteFileRecord(ctx, fileRecord.ID)
	}); err != nil {
		return err
	}
	if storage.NormalizeStorageType(fileRecord.StorageType) != storage.StorageTypeExternal {
		_ = s.DeleteStoredFile(fileRecord.StorageType, dto.Key)
	}
	return nil
}

func (s *FileService) UploadAudioFile(
	ctx context.Context,
	file *multipart.FileHeader,
) (commonModel.FileDto, error) {
	fileDto, err := s.UploadFile(ctx, file, storage.CategoryAudio, storage.StorageTypeLocal)
	if err != nil {
		return commonModel.FileDto{}, err
	}

	if err := s.transactor.Run(context.Background(), func(ctx context.Context) error {
		return s.keyvalueRepository.AddOrUpdateKeyValue(ctx, globalAudioFileIDKey, fileDto.ID)
	}); err != nil {
		_ = s.transactor.Run(context.Background(), func(ctx context.Context) error {
			return s.DeleteFileRecord(ctx, fileDto.ID)
		})
		stored, getErr := s.fileRepository.GetByID(context.Background(), fileDto.ID)
		if getErr == nil && stored != nil {
			_ = s.DeleteStoredFile(stored.StorageType, fileDto.Key)
		}
		return commonModel.FileDto{}, err
	}

	return fileDto, nil
}

func (s *FileService) DeleteAudioFile(ctx context.Context) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonRepository.GetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	baseCtx := context.Background()
	val, err := s.keyvalueRepository.GetKeyValue(baseCtx, globalAudioFileIDKey)
	if err != nil || val == "" {
		return nil
	}

	fileRecord, err := s.fileRepository.GetByID(baseCtx, val)
	if err != nil {
		return nil
	}

	if err := s.transactor.Run(baseCtx, func(txCtx context.Context) error {
		if err := s.keyvalueRepository.DeleteKeyValue(txCtx, globalAudioFileIDKey); err != nil {
			return err
		}
		return s.fileRepository.Delete(txCtx, fileRecord.ID)
	}); err != nil {
		return err
	}

	if storage.NormalizeStorageType(fileRecord.StorageType) != storage.StorageTypeExternal {
		_ = s.DeleteStoredFile(fileRecord.StorageType, fileRecord.Key)
	}
	return nil
}

func (s *FileService) GetCurrentAudioURL() string {
	val, err := s.keyvalueRepository.GetKeyValue(context.Background(), globalAudioFileIDKey)
	if err != nil || val == "" {
		return ""
	}
	fileRecord, err := s.fileRepository.GetByID(context.Background(), val)
	if err != nil {
		return ""
	}
	return fileRecord.URL
}

func (s *FileService) StreamCurrentAudio(ctx *gin.Context) {
	val, _ := s.keyvalueRepository.GetKeyValue(context.Background(), globalAudioFileIDKey)
	if val == "" {
		ctx.String(http.StatusNotFound, "音乐文件不存在")
		return
	}
	fileRecord, err := s.fileRepository.GetByID(context.Background(), val)
	if err != nil {
		ctx.String(http.StatusNotFound, "音乐文件不存在")
		return
	}

	contentType := fileRecord.ContentType
	if contentType == "" {
		contentType = "audio/mpeg"
	}
	ctx.Header("Content-Type", contentType)
	ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.Header("Pragma", "no-cache")
	ctx.Header("Expires", "0")

	reader, err := s.getSelector().Get(
		context.Background(),
		storage.NormalizeStorageType(fileRecord.StorageType),
		fileRecord.Key,
	)
	if err != nil {
		ctx.String(http.StatusNotFound, "音乐文件不存在")
		return
	}
	defer func() { _ = reader.Close() }()

	readSeeker, ok := reader.(io.ReadSeeker)
	if !ok {
		ctx.String(http.StatusInternalServerError, "音乐文件读取失败")
		return
	}
	http.ServeContent(ctx.Writer, ctx.Request, fileRecord.Name, fileRecord.CreatedAt, readSeeker)
}

func (s *FileService) GetFilePresignURL(
	ctx context.Context,
	dto *commonModel.GetPresignURLDto,
) (commonModel.PresignDto, error) {
	var result commonModel.PresignDto
	userid := viewer.MustFromContext(ctx).UserID()

	user, err := s.commonRepository.GetUserByUserId(context.Background(), userid)
	if err != nil {
		return result, err
	}
	if !user.IsAdmin {
		return result, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	if dto.FileName == "" {
		return result, errors.New(commonModel.INVALID_PARAMS)
	}
	if st := strings.TrimSpace(dto.StorageType); st != "" &&
		storage.NormalizeStorageType(st) != storage.StorageTypeObject {
		return result, errors.New(commonModel.INVALID_PARAMS)
	}

	contentType := dto.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	category := storage.CategoryImage
	if strings.HasPrefix(contentType, "audio/") {
		category = storage.CategoryAudio
	}
	if !isAllowedType(contentType, config.Config().Upload.AllowedTypes) {
		return result, errors.New(commonModel.FILE_TYPE_NOT_ALLOWED)
	}

	key, err := s.keyGen.GenerateKey(category, userid, dto.FileName)
	if err != nil {
		return result, err
	}

	selector := s.getSelector()
	presignURL, err := selector.PresignPutURL(context.Background(), key, 24*time.Hour)
	if err != nil {
		return result, err
	}

	fileURL := selector.ResolveURL(storage.StorageTypeObject, key)
	storageType, provider, bucket := currentStorageRoute(selector, storage.StorageTypeObject)
	fileRecord := &fileModel.File{
		Key:         key,
		StorageType: storageType,
		Provider:    provider,
		Bucket:      bucket,
		URL:         fileURL,
		Name:        dto.FileName,
		ContentType: contentType,
		Category:    string(category),
		UserID:      userid,
	}
	if err := s.fileRepository.Create(context.Background(), fileRecord); err != nil {
		return result, err
	}

	result.ID = fileRecord.ID
	result.FileName = dto.FileName
	result.ContentType = contentType
	result.Key = key
	result.PresignURL = presignURL
	result.FileURL = fileURL
	return result, nil
}

func (s *FileService) CleanupOrphanFiles() error {
	ctx := context.Background()
	threshold := time.Now().UTC().Add(-24 * time.Hour)

	files, err := s.fileRepository.GetOrphanFiles(ctx, threshold)
	if err != nil {
		return err
	}

	currentAudioFileID, _ := s.keyvalueRepository.GetKeyValue(ctx, globalAudioFileIDKey)
	for _, file := range files {
		if currentAudioFileID != "" && file.ID == currentAudioFileID {
			continue
		}
		if file.Key != "" && storage.NormalizeStorageType(file.StorageType) != storage.StorageTypeExternal {
			_ = s.DeleteStoredFile(file.StorageType, file.Key)
		}
		_ = s.fileRepository.Delete(ctx, file.ID)
	}

	return nil
}

func (s *FileService) DeleteFileRecord(ctx context.Context, id string) error {
	return s.fileRepository.Delete(ctx, id)
}

func (s *FileService) DeleteStoredFile(storageType string, key string) error {
	if key == "" {
		return nil
	}
	normalizedStorageType := storage.NormalizeStorageType(storageType)
	if normalizedStorageType == storage.StorageTypeExternal {
		return nil
	}
	return s.getSelector().Delete(context.Background(), normalizedStorageType, key)
}

func (s *FileService) keyGenForCategory(category storage.Category, fileName string) storage.KeyGenerator {
	if category == storage.CategoryAudio {
		ext := strings.ToLower(filepath.Ext(strings.TrimSpace(fileName)))
		if ext == "" {
			ext = ".bin"
		}
		return &storage.StaticKeyGenerator{Name: "music" + ext}
	}
	return s.keyGen
}

func isAllowedType(contentType string, allowedTypes []string) bool {
	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

func (s *FileService) getSelector() *storage.StorageSelector {
	if s.storageManager == nil {
		return storage.NewStorageSelector(config.Config().Storage)
	}
	if selector := s.storageManager.GetSelector(); selector != nil {
		return selector
	}
	return storage.NewStorageSelector(config.Config().Storage)
}

func currentStorageRoute(
	selector *storage.StorageSelector,
	storageType storage.StorageType,
) (resolvedType, provider, bucket string) {
	switch storage.NormalizeStorageType(string(storageType)) {
	case storage.StorageTypeObject:
		if selector != nil {
			provider, bucket = selector.ObjectRoute()
		}
		return string(storage.StorageTypeObject), provider, bucket
	default:
		return string(storage.StorageTypeLocal), "", ""
	}
}
