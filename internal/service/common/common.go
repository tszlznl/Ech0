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
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

type CommonService struct {
	txManager          transaction.TransactionManager
	commonRepository   repository.CommonRepositoryInterface
	storagePort        storageDomain.StoragePort
	echoRepository     echoRepository.EchoRepositoryInterface
	keyvalueRepository keyvalueRepository.KeyValueRepositoryInterface
	fs                 afero.Fs
	eventBus           event.IEventBus
}

func NewCommonService(
	tm transaction.TransactionManager,
	commonRepository repository.CommonRepositoryInterface,
	echoRepository echoRepository.EchoRepositoryInterface,
	keyvalueRepository keyvalueRepository.KeyValueRepositoryInterface,
	storagePort storageDomain.StoragePort,
	fs afero.Fs,
	eventBusProvider func() event.IEventBus,
) *CommonService {
	return &CommonService{
		txManager:          tm,
		commonRepository:   commonRepository,
		echoRepository:     echoRepository,
		keyvalueRepository: keyvalueRepository,
		storagePort:        storagePort,
		fs:                 fs,
		eventBus:           eventBusProvider(),
	}
}

func (commonService *CommonService) CommonGetUserByUserId(userId uint) (userModel.User, error) {
	return commonService.commonRepository.GetUserByUserId(userId)
}

func (commonService *CommonService) UploadImage(
	userId uint,
	file *multipart.FileHeader,
	source string,
) (commonModel.ImageDto, error) {
	fileDto, err := commonService.UploadFile(userId, file, source, storageDomain.CategoryImage)
	if err != nil {
		return commonModel.ImageDto{}, err
	}
	return commonModel.ImageDto{
		URL:       fileDto.URL,
		SOURCE:    fileDto.Source,
		ObjectKey: fileDto.ObjectKey,
		Width:     fileDto.Width,
		Height:    fileDto.Height,
	}, nil
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

	// 检查文件类型是否合法
	if !storageUtil.IsAllowedType(
		file.Header.Get("Content-Type"),
		config.Config().Upload.AllowedTypes,
	) {
		return commonModel.FileDto{}, errors.New(commonModel.FILE_TYPE_NOT_ALLOWED)
	}

	// 检查文件大小是否合法
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

	readSeeker, ok := reader.(storageDomain.ReadSeekCloser)
	if !ok {
		return commonModel.FileDto{}, errors.New("uploaded file does not support seek")
	}

	storedObject, err := commonService.storagePort.Save(context.Background(), storageDomain.SaveRequest{
		UserID:      user.ID,
		FileName:    file.Filename,
		ContentType: file.Header.Get("Content-Type"),
		Category:    category,
		Reader:      readSeeker,
	})
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

	// 触发图片上传事件
	user.Password = "" // 清除密码字段，避免泄露
	if err := commonService.eventBus.Publish(context.Background(), event.NewEvent(
		event.EventTypeResourceUploaded,
		event.EventPayload{
			event.EventPayloadUser: user,
			event.EventPayloadFile: file.Filename,
			event.EventPayloadURL:  storedObject.URL,
			event.EventPayloadSize: file.Size,
			event.EventPayloadType: uploadType,
		},
	)); err != nil {
		logUtil.GetLogger().
			Error("Failed to publish resource uploaded event", zap.String("error", err.Error()))
	}

	fileDto := commonModel.FileDto{
		URL:         storedObject.URL,
		Source:      string(storedObject.Source),
		ObjectKey:   storedObject.ObjectKey,
		ContentType: storedObject.ContentType,
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

func (commonService *CommonService) DeleteImage(userid uint, url, source, objectKey string) error {
	return commonService.DeleteFile(userid, commonModel.FileDto{
		URL:       url,
		Source:    source,
		ObjectKey: objectKey,
	})
}

func (commonService *CommonService) DeleteFile(userid uint, file commonModel.FileDto) error {
	user, err := commonService.commonRepository.GetUserByUserId(userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 检查图片是否存在
	if file.URL == "" {
		return errors.New(commonModel.IMAGE_NOT_FOUND)
	}

	return commonService.deleteFileBySource(file.URL, normalizeFileSource(file.Source), file.ObjectKey)
}

func (commonService *CommonService) DirectDeleteImage(url, source, objectKey string) error {
	// 检查图片是否存在
	if url == "" {
		return errors.New(commonModel.IMAGE_NOT_FOUND)
	}

	return commonService.deleteFileBySource(url, normalizeFileSource(source), objectKey)
}

func (commonService *CommonService) deleteFileBySource(url, source, objectKey string) error {
	if source == echoModel.ImageSourceURL {
		return nil
	}
	return commonService.storagePort.Delete(context.Background(), storageDomain.DeleteRequest{
		URL:       url,
		Source:    storageDomain.Source(source),
		ObjectKey: objectKey,
		Category:  storageDomain.CategoryFile,
	})
}

func (commonService *CommonService) GetSysAdmin() (userModel.User, error) {
	return commonService.commonRepository.GetSysAdmin()
}

func (commonService *CommonService) GetStatus() (commonModel.Status, error) {
	// 获取系统管理员信息
	sysuser, err := commonService.commonRepository.GetSysAdmin()
	if err != nil {
		return commonModel.Status{}, err
	}

	// 获取所有用户状态信息
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

	status.SysAdminID = sysuser.ID     // 管理员ID
	status.Username = sysuser.Username // 管理员用户名
	status.Logo = sysuser.Avatar       // 管理员头像
	status.Users = users               // 所有用户状态
	status.TotalEchos = len(echos)     // Echo总数

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
	// 获取所有Echo
	echos, err := commonService.commonRepository.GetAllEchos(false)
	if err != nil {
		return "", err
	}

	// 生成 RSS 订阅链接
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

		// 添加图片链接到正文前(scheme://host/api/ImageURL)
		if len(msg.Images) > 0 {
			var imageContent []byte
			for _, image := range msg.Images {
				// 根据图片来源生成链接
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

		// 添加标签到正文后
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

	// 支持的音频格式
	audioFiles := []string{"music.flac", "music.m4a", "music.mp3"}

	for _, file := range audioFiles {
		audioPath := fmt.Sprintf("data/audios/%s", file)
		if storageUtil.FileExists(commonService.fs, audioPath) {
			return storageUtil.DeleteFileFromLocal(commonService.fs, audioPath)
		}
	}

	return nil
}

func (commonService *CommonService) GetPlayMusicUrl() string {
	// 支持的音频格式
	audioFiles := []string{"music.flac", "music.m4a", "music.mp3"}

	for _, file := range audioFiles {
		audioPath := fmt.Sprintf("data/audios/%s", file)
		if storageUtil.FileExists(commonService.fs, audioPath) {
			return fmt.Sprintf("/audios/%s", file)
		}
	}

	// 没有找到音频文件
	return ""
}

func (commonService *CommonService) PlayMusic(ctx *gin.Context) {
	// 以文件流的形式返回音乐文件
	musicURL := commonService.GetPlayMusicUrl()
	musicName := ""
	if musicURL != "" {
		// 只保留最后的文件名
		musicName = musicURL[len("/audios/"):]
	}

	// 获取音乐文件的路径
	musicPath := config.Config().Upload.AudioPath + musicName

	// 检查文件是否存在
	stat, err := commonService.fs.Stat(musicPath)
	if err != nil {
		ctx.String(http.StatusNotFound, "音乐文件不存在")
		return
	}

	// 获取 Content-Type
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

	// 设置响应头
	ctx.Header("Content-Type", contentType)
	ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.Header("Pragma", "no-cache")
	ctx.Header("Expires", "0")

	reader, err := commonService.fs.Open(musicPath)
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

	http.ServeContent(ctx.Writer, ctx.Request, musicName, stat.ModTime(), readSeeker)
}

// GetS3PresignURL 获取 S3 预签名 URL
func (commonService *CommonService) GetS3PresignURL(
	userid uint,
	s3Dto *commonModel.GetPresignURLDto,
	method string,
) (commonModel.PresignDto, error) {
	return commonService.GetFilePresignURL(userid, s3Dto, method)
}

func (commonService *CommonService) GetFilePresignURL(
	userid uint,
	s3Dto *commonModel.GetPresignURLDto,
	method string,
) (commonModel.PresignDto, error) {
	var result commonModel.PresignDto

	// 权限检查
	user, err := commonService.commonRepository.GetUserByUserId(userid)
	if err != nil {
		return result, err
	}
	if !user.IsAdmin {
		return result, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 参数校验
	if s3Dto.FileName == "" {
		return result, errors.New(commonModel.INVALID_PARAMS)
	}
	ext := filepath.Ext(s3Dto.FileName) // ".png"
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 检查Content-Type是否为Image开头
	switch contentType[:5] {
	case "image":
		// 检查文件类型是否合法
		if !storageUtil.IsAllowedType(contentType, config.Config().Upload.AllowedTypes) {
			return result, errors.New(commonModel.FILE_TYPE_NOT_ALLOWED)
		}
	case "audio":
		// 检查文件类型是否合法
		if !storageUtil.IsAllowedType(contentType, config.Config().Upload.AllowedTypes) {
			return result, errors.New(commonModel.FILE_TYPE_NOT_ALLOWED)
		}
	default:
		return result, errors.New(commonModel.FILE_TYPE_NOT_ALLOWED)
	}

	// 填充返回结果
	result.FileName = s3Dto.FileName
	result.ContentType = contentType

	presignResponse, err := commonService.storagePort.PresignUpload(context.Background(), storageDomain.PresignRequest{
		UserID:      userid,
		FileName:    s3Dto.FileName,
		ContentType: contentType,
		Method:      method,
		Expiry:      24 * time.Hour,
		Category:    storageDomain.CategoryImage,
	})
	if err != nil {
		return result, err
	}
	result.ObjectKey = presignResponse.ObjectKey
	result.PresignURL = presignResponse.PresignURL
	result.FileURL = presignResponse.FileURL

	// 保存到临时文件表
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

// GetS3Client 获取 S3 客户端和配置信息（支持 R2 / AWS / MinIO / 其他）

// CleanupTempFiles 清理过期的临时文件
func (commonService *CommonService) CleanupTempFiles() error {
	// 获取所有未删除的临时文件
	files, err := commonService.commonRepository.GetAllTempFiles()
	if err != nil {
		return err
	}

	// 当前时间戳
	now := time.Now().UTC().Unix()

	for _, file := range files {
		// 如果最后访问时间超过24小时，则删除
		if now-file.LastAccessedAt > 24*3600 {
			// 删除文件
			switch file.Storage {
			case string(commonModel.LOCAL_FILE):
				// TODO: 删除本地文件

			case string(commonModel.S3_FILE):
				if file.ObjectKey == "" {
					// 如果没有传入 object_key，则无法删除,忽略
					continue
				}
				if err := commonService.storagePort.Delete(context.Background(), storageDomain.DeleteRequest{
					Source:    storageDomain.SourceS3,
					ObjectKey: file.ObjectKey,
					Category:  storageDomain.CategoryFile,
				}); err != nil {
					// 记录日志，继续处理下一个文件
					fmt.Printf("删除S3临时文件失败: %s, 错误: %v\n", file.ObjectKey, err)
				}
			default:
				// 未知存储类型，忽略
			}

			// 从数据库中删除记录(开启事务)
			_ = commonService.txManager.Run(func(ctx context.Context) error {
				return commonService.commonRepository.DeleteTempFilePermanently(ctx, file.ID)
			})
		}
	}

	return nil
}

func (commonService *CommonService) RefreshEchoImageURL(echo *echoModel.Echo) {
	// 用 channel 或 waitGroup 并发刷新 URL
	var wg sync.WaitGroup
	mu := sync.Mutex{}

	for i := range echo.Images {
		if echo.Images[i].ImageSource == echoModel.ImageSourceS3 && echo.Images[i].ObjectKey != "" {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				if newURL, err := commonService.storagePort.ResolveURL(context.Background(), echo.Images[i].ObjectKey); err == nil {
					mu.Lock()
					echo.Images[i].ImageURL = newURL
					mu.Unlock()
				}
			}(i)
		}
	}

	wg.Wait()

	// 所有 URL 都拿到了，再一次性更新 DB
	_ = commonService.txManager.Run(func(ctx context.Context) error {
		return commonService.echoRepository.UpdateEcho(ctx, echo)
	})
}

// GetWebsiteTitle 获取网站标题
func (commonService *CommonService) GetWebsiteTitle(websiteURL string) (string, error) {
	websiteURL = httpUtil.TrimURL(websiteURL)

	body, err := httpUtil.SendRequest(websiteURL, "GET", httpUtil.Header{}, 10*time.Second)
	if err != nil {
		return "", err
	}

	// 解析 HTML 并提取标题
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

// extractTitle 从 HTML 节点中提取 title 标签的内容
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
