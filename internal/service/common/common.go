package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/event"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	stgx "github.com/lin-snow/ech0/pkg/storagex"
	repository "github.com/lin-snow/ech0/internal/repository/common"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	storageDomain "github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/transaction"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	imgUtil "github.com/lin-snow/ech0/internal/util/img"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	mdUtil "github.com/lin-snow/ech0/internal/util/md"
	storageUtil "github.com/lin-snow/ech0/internal/util/storage"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

type CommonService struct {
	txManager          transaction.TransactionManager
	commonRepository   repository.CommonRepositoryInterface
	storage            *storageDomain.StorageService
	echoRepository     echoRepository.EchoRepositoryInterface
	keyvalueRepository keyvalueRepository.KeyValueRepositoryInterface
	eventBus           event.IEventBus
}

func NewCommonService(
	tm transaction.TransactionManager,
	commonRepository repository.CommonRepositoryInterface,
	echoRepository echoRepository.EchoRepositoryInterface,
	keyvalueRepository keyvalueRepository.KeyValueRepositoryInterface,
	storage *storageDomain.StorageService,
	eventBusProvider func() event.IEventBus,
) *CommonService {
	return &CommonService{
		txManager:          tm,
		commonRepository:   commonRepository,
		echoRepository:     echoRepository,
		keyvalueRepository: keyvalueRepository,
		storage:            storage,
		eventBus:           eventBusProvider(),
	}
}

func (commonService *CommonService) CommonGetUserByUserId(userId uint) (userModel.User, error) {
	return commonService.commonRepository.GetUserByUserId(userId)
}

func (commonService *CommonService) UploadFile(
	userId uint,
	file *multipart.FileHeader,
	source string,
	category storageDomain.Category,
) (commonModel.FileDto, error) {
	_ = source
	user, err := commonService.commonRepository.GetUserByUserId(userId)
	if err != nil {
		return commonModel.FileDto{}, err
	}
	if !user.IsAdmin {
		return commonModel.FileDto{}, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	if !storageUtil.IsAllowedType(
		file.Header.Get("Content-Type"),
		config.Config().Upload.AllowedTypes,
	) {
		return commonModel.FileDto{}, errors.New(commonModel.FILE_TYPE_NOT_ALLOWED)
	}

	limit := int64(config.Config().Upload.ImageMaxSize)
	if category == storageDomain.CategoryAudio {
		limit = int64(config.Config().Upload.AudioMaxSize)
	}
	if file.Size > limit {
		return commonModel.FileDto{}, errors.New(commonModel.FILE_SIZE_EXCEED_LIMIT)
	}

	uploadType := commonModel.ImageType
	if category == storageDomain.CategoryAudio {
		uploadType = commonModel.AudioType
	}

	reader, err := file.Open()
	if err != nil {
		return commonModel.FileDto{}, err
	}
	defer func() { _ = reader.Close() }()

	contentType := file.Header.Get("Content-Type")
	result, err := commonService.storage.Upload(
		context.Background(),
		stgx.Category(category),
		user.ID,
		file.Filename,
		contentType,
		reader,
	)
	if err != nil {
		return commonModel.FileDto{}, err
	}

	width := 0
	height := 0
	if category.IsImageLike() {
		width, height, err = imgUtil.GetImageSizeFromFile(file)
		if err != nil {
			return commonModel.FileDto{}, err
		}
	}

	user.Password = ""
	if err := commonService.eventBus.Publish(context.Background(), event.NewEvent(
		event.EventTypeResourceUploaded,
		event.EventPayload{
			event.EventPayloadUser: user,
			event.EventPayloadFile: file.Filename,
			event.EventPayloadURL:  result.URL,
			event.EventPayloadSize: file.Size,
			event.EventPayloadType: uploadType,
		},
	)); err != nil {
		logUtil.GetLogger().
			Error("Failed to publish resource uploaded event", zap.String("error", err.Error()))
	}

	fileDto := commonModel.FileDto{
		URL:         result.URL,
		AccessURL:   ResolveAccessFileURL(result.URL, commonService.storage.Source()),
		Source:      commonService.storage.Source(),
		ObjectKey:   result.ObjectKey,
		ContentType: result.ContentType,
		Category:    string(category),
		Width:       width,
		Height:      height,
	}
	if category.IsImageLike() {
		fileDto.Metadata.Image = &commonModel.ImageMetadataDto{
			Width:  width,
			Height: height,
		}
	}

	return fileDto, nil
}

func ResolveAccessFileURL(rawURL, source string) string {
	cleanURL := strings.TrimSpace(rawURL)
	if cleanURL == "" {
		return ""
	}
	lowerURL := strings.ToLower(cleanURL)
	if strings.HasPrefix(lowerURL, "http://") || strings.HasPrefix(lowerURL, "https://") {
		return cleanURL
	}

	if source == echoModel.ImageSourceLocal {
		if strings.HasPrefix(cleanURL, "/api/") {
			return cleanURL
		}
		if strings.HasPrefix(cleanURL, "/") {
			return "/api" + cleanURL
		}
		return "/api/" + strings.TrimLeft(cleanURL, "/")
	}

	if strings.HasPrefix(cleanURL, "/") {
		return cleanURL
	}
	return "/" + strings.TrimLeft(cleanURL, "/")
}

func (commonService *CommonService) DeleteFile(userid uint, file commonModel.FileDto) error {
	user, err := commonService.commonRepository.GetUserByUserId(userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	if file.URL == "" {
		return errors.New(commonModel.IMAGE_NOT_FOUND)
	}

	return commonService.deleteFileBySource(file.URL, normalizeFileSource(file.Source), file.ObjectKey)
}

func (commonService *CommonService) deleteFileBySource(url, source, objectKey string) error {
	if source == echoModel.ImageSourceURL {
		return nil
	}
	vpath := objectKeyToVirtualPath(objectKey)
	if vpath == "" {
		return nil
	}
	return commonService.storage.Delete(context.Background(), vpath)
}

func (commonService *CommonService) GetSysAdmin() (userModel.User, error) {
	return commonService.commonRepository.GetSysAdmin()
}

func (commonService *CommonService) GetStatus() (commonModel.Status, error) {
	sysuser, err := commonService.commonRepository.GetSysAdmin()
	if err != nil {
		return commonModel.Status{}, err
	}

	var users []commonModel.UserStatus
	allusers, err := commonService.commonRepository.GetAllUsers()
	if err != nil {
		return commonModel.Status{}, err
	}
	for _, user := range allusers {
		users = append(users, commonModel.UserStatus{
			UserID:   user.ID,
			UserName: user.Username,
			IsAdmin:  user.IsAdmin,
		})
	}

	status := commonModel.Status{}

	echos, err := commonService.commonRepository.GetAllEchos(true)
	if err != nil {
		return status, err
	}

	status.SysAdminID = sysuser.ID
	status.Username = sysuser.Username
	status.Logo = sysuser.Avatar
	status.Users = users
	status.TotalEchos = len(echos)

	return status, nil
}

func (commonService *CommonService) GetHeatMap(timezone string) ([]commonModel.Heatmap, error) {
	loc := timezoneUtil.LoadLocationOrUTC(timezone)
	nowUser := time.Now().UTC().In(loc)
	startUser := time.Date(nowUser.Year(), nowUser.Month(), nowUser.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -29)
	endUserExclusive := startUser.AddDate(0, 0, 30)

	createdAtList, err := commonService.commonRepository.GetHeatMap(startUser.UTC(), endUserExclusive.UTC())
	if err != nil {
		return nil, err
	}

	countMap := make(map[string]int)
	for _, createdAt := range createdAtList {
		day := createdAt.In(loc).Format("2006-01-02")
		countMap[day]++
	}

	var results [30]commonModel.Heatmap
	for i := 0; i < 30; i++ {
		date := startUser.AddDate(0, 0, i).Format("2006-01-02")
		results[i] = commonModel.Heatmap{
			Date:  date,
			Count: countMap[date],
		}
	}

	return results[:], nil
}

func (commonService *CommonService) GenerateRSS(ctx *gin.Context) (string, error) {
	echos, err := commonService.commonRepository.GetAllEchos(false)
	if err != nil {
		return "", err
	}

	schema := "http"
	if ctx.Request.TLS != nil {
		schema = "https"
	}
	host := ctx.Request.Host
	feed := &feeds.Feed{
		Title: "Ech0",
		Link: &feeds.Link{
			Href: fmt.Sprintf("%s://%s/", schema, host),
		},
		Image: &feeds.Image{
			Url: fmt.Sprintf("%s://%s/Ech0.svg", schema, host),
		},
		Description: "Ech0",
		Author: &feeds.Author{
			Name: "Ech0",
		},
		Updated: time.Now().UTC(),
	}

	for _, msg := range echos {
		renderedContent := mdUtil.MdToHTML([]byte(msg.Content))

		title := msg.Username + " - " + msg.CreatedAt.Format("2006-01-02")

		if len(msg.Images) > 0 {
			var imageContent []byte
			for _, image := range msg.Images {
				var imageURL string
				switch image.ImageSource {
				case echoModel.ImageSourceLocal:
					imageURL = fmt.Sprintf("%s://%s/api%s", schema, host, image.ImageURL)
				case echoModel.ImageSourceS3:
					imageURL = image.ImageURL
				}
				imageContent = fmt.Appendf(
					imageContent,
					"<img src=\"%s\" alt=\"Image\" style=\"max-width:100%%;height:auto;\" />",
					imageURL,
				)
			}
			renderedContent = append(imageContent, renderedContent...)
		}

		if len(msg.Tags) > 0 {
			for _, tag := range msg.Tags {
				renderedContent = fmt.Appendf(
					renderedContent,
					"<br /><span class=\"tag\">#%s</span>",
					tag.Name,
				)
			}
		}

		item := &feeds.Item{
			Title:       title,
			Link:        &feeds.Link{Href: fmt.Sprintf("%s://%s/echo/%d", schema, host, msg.ID)},
			Description: string(renderedContent),
			Author: &feeds.Author{
				Name: msg.Username,
			},
			Created: msg.CreatedAt,
		}
		feed.Items = append(feed.Items, item)
	}

	atom, err := feed.ToAtom()
	if err != nil {
		return "", err
	}

	return atom, nil
}

func (commonService *CommonService) UploadMusic(
	userId uint,
	file *multipart.FileHeader,
) (string, error) {
	fileDto, err := commonService.UploadFile(userId, file, string(echoModel.ImageSourceLocal), storageDomain.CategoryAudio)
	if err != nil {
		return "", err
	}

	return fileDto.URL, nil
}

func (commonService *CommonService) DeleteMusic(userid uint) error {
	user, err := commonService.commonRepository.GetUserByUserId(userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	audioFiles := []string{"music.flac", "music.m4a", "music.mp3"}
	ctx := context.Background()

	for _, file := range audioFiles {
		vpath := stgx.JoinPath("audios", file)
		if exists, _ := commonService.storage.Exists(ctx, vpath); exists {
			return commonService.storage.Delete(ctx, vpath)
		}
	}

	return nil
}

func (commonService *CommonService) GetPlayMusicUrl() string {
	audioFiles := []string{"music.flac", "music.m4a", "music.mp3"}
	ctx := context.Background()

	for _, file := range audioFiles {
		vpath := stgx.JoinPath("audios", file)
		if exists, _ := commonService.storage.Exists(ctx, vpath); exists {
			url, err := commonService.storage.ResolveURL(ctx, vpath)
			if err == nil {
				return url
			}
			return fmt.Sprintf("/files/audios/%s", file)
		}
	}

	return ""
}

func (commonService *CommonService) PlayMusic(ctx *gin.Context) {
	musicURL := commonService.GetPlayMusicUrl()
	if musicURL == "" {
		ctx.String(http.StatusNotFound, "音乐文件不存在")
		return
	}

	musicName := filepath.Base(musicURL)
	vpath := stgx.JoinPath("audios", musicName)

	info, err := commonService.storage.Stat(context.Background(), vpath)
	if err != nil {
		ctx.String(http.StatusNotFound, "音乐文件不存在")
		return
	}

	contentType := "audio/mpeg"
	lowerName := strings.ToLower(musicName)
	switch {
	case strings.HasSuffix(lowerName, ".flac"):
		contentType = "audio/flac"
	case strings.HasSuffix(lowerName, ".wav"):
		contentType = "audio/wav"
	case strings.HasSuffix(lowerName, ".m4a"):
		contentType = "audio/mp4"
	}

	ctx.Header("Content-Type", contentType)
	ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.Header("Pragma", "no-cache")
	ctx.Header("Expires", "0")

	reader, err := commonService.storage.Open(context.Background(), vpath)
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

	http.ServeContent(ctx.Writer, ctx.Request, musicName, info.ModTime, readSeeker)
}

func (commonService *CommonService) GetFilePresignURL(
	userid uint,
	s3Dto *commonModel.GetPresignURLDto,
	method string,
) (commonModel.PresignDto, error) {
	var result commonModel.PresignDto

	user, err := commonService.commonRepository.GetUserByUserId(userid)
	if err != nil {
		return result, err
	}
	if !user.IsAdmin {
		return result, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	if s3Dto.FileName == "" {
		return result, errors.New(commonModel.INVALID_PARAMS)
	}
	ext := filepath.Ext(s3Dto.FileName)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	switch contentType[:5] {
	case "image":
		if !storageUtil.IsAllowedType(contentType, config.Config().Upload.AllowedTypes) {
			return result, errors.New(commonModel.FILE_TYPE_NOT_ALLOWED)
		}
	case "audio":
		if !storageUtil.IsAllowedType(contentType, config.Config().Upload.AllowedTypes) {
			return result, errors.New(commonModel.FILE_TYPE_NOT_ALLOWED)
		}
	default:
		return result, errors.New(commonModel.FILE_TYPE_NOT_ALLOWED)
	}

	result.FileName = s3Dto.FileName
	result.ContentType = contentType

	presignResult, err := commonService.storage.Presign(
		context.Background(),
		stgx.CategoryImage,
		userid,
		s3Dto.FileName,
		contentType,
		method,
		24*time.Hour,
	)
	if err != nil {
		return result, err
	}
	result.ObjectKey = presignResult.ObjectKey
	result.PresignURL = presignResult.PresignURL
	result.FileURL = presignResult.FileURL

	now := time.Now().UTC().Unix()
	fileType := string(storageDomain.CategoryImage)
	if strings.HasPrefix(contentType, "audio") {
		fileType = string(storageDomain.CategoryAudio)
	}
	tempFile := commonModel.TempFile{
		FileName:       result.FileName,
		Storage:        string(commonModel.S3_FILE),
		FileType:       fileType,
		Bucket:         "",
		ObjectKey:      result.ObjectKey,
		Deleted:        false,
		CreatedAt:      now,
		LastAccessedAt: now,
	}
	if err := commonService.txManager.Run(func(ctx context.Context) error {
		return commonService.commonRepository.SaveTempFile(ctx, tempFile)
	}); err != nil {
		logUtil.GetLogger().Error("Failed to save temp file", zap.String("error", err.Error()))
	}

	return result, nil
}

func normalizeFileSource(source string) string {
	switch source {
	case echoModel.ImageSourceS3:
		return echoModel.ImageSourceS3
	case echoModel.ImageSourceURL:
		return echoModel.ImageSourceURL
	default:
		return echoModel.ImageSourceLocal
	}
}

func (commonService *CommonService) CleanupTempFiles() error {
	files, err := commonService.commonRepository.GetAllTempFiles()
	if err != nil {
		return err
	}

	now := time.Now().UTC().Unix()
	ctx := context.Background()

	for _, file := range files {
		if now-file.LastAccessedAt > 24*3600 {
			if file.ObjectKey != "" {
				vpath := objectKeyToVirtualPath(file.ObjectKey)
				if vpath != "" {
					if err := commonService.storage.Delete(ctx, vpath); err != nil {
						fmt.Printf("删除临时文件失败: %s, 错误: %v\n", file.ObjectKey, err)
					}
				}
			}

			_ = commonService.txManager.Run(func(ctx context.Context) error {
				return commonService.commonRepository.DeleteTempFilePermanently(ctx, file.ID)
			})
		}
	}

	return nil
}

func (commonService *CommonService) RefreshEchoImageURL(echo *echoModel.Echo) {
	var wg sync.WaitGroup
	mu := sync.Mutex{}

	for i := range echo.Images {
		if echo.Images[i].ImageSource == echoModel.ImageSourceS3 && echo.Images[i].ObjectKey != "" {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				vpath := objectKeyToVirtualPath(echo.Images[i].ObjectKey)
				if newURL, err := commonService.storage.ResolveURL(context.Background(), vpath); err == nil {
					mu.Lock()
					echo.Images[i].ImageURL = newURL
					mu.Unlock()
				}
			}(i)
		}
	}

	wg.Wait()

	_ = commonService.txManager.Run(func(ctx context.Context) error {
		return commonService.echoRepository.UpdateEcho(ctx, echo)
	})
}

func (commonService *CommonService) GetWebsiteTitle(websiteURL string) (string, error) {
	websiteURL = httpUtil.TrimURL(websiteURL)

	body, err := httpUtil.SendRequest(websiteURL, "GET", httpUtil.Header{}, 10*time.Second)
	if err != nil {
		return "", err
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("解析 HTML 失败: %w", err)
	}

	title := extractTitle(doc)
	if title == "" {
		return "", errors.New("未找到网站标题")
	}

	return title, nil
}

func extractTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		if n.FirstChild != nil {
			return strings.TrimSpace(n.FirstChild.Data)
		}
		return ""
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if title := extractTitle(c); title != "" {
			return title
		}
	}

	return ""
}

// objectKeyToVirtualPath converts a stored object key (e.g. "images/a.png")
// to a VFS virtual path (e.g. "/images/a.png").
func objectKeyToVirtualPath(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	return "/" + strings.TrimLeft(key, "/")
}
