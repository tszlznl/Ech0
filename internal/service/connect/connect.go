package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/connect"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/lin-snow/ech0/internal/transaction"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
	"github.com/lin-snow/ech0/pkg/viewer"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

const (
	connectsInfoCacheTTL        = 30 * time.Minute
	connectFanoutMaxConcurrency = 8
	connectsInfoSingleflightKey = "connects_info"
)

type ConnectService struct {
	transactor        transaction.Transactor
	connectRepository Repository
	echoRepository    EchoRepository
	commonService     CommonService
	settingService    SettingService

	connectsInfoCacheMu      sync.RWMutex
	connectsInfoCache        []model.Connect
	connectsInfoCacheExpires time.Time
	connectsInfoCacheValid   bool
	connectsInfoFetcher      singleflight.Group
}

func NewConnectService(
	tx transaction.Transactor,
	connectRepository Repository,
	echoRepository EchoRepository,
	commonService CommonService,
	settingService SettingService,
) *ConnectService {
	return &ConnectService{
		transactor:        tx,
		connectRepository: connectRepository,
		echoRepository:    echoRepository,
		commonService:     commonService,
		settingService:    settingService,
	}
}

// AddConnect 添加连接
func (connectService *ConnectService) AddConnect(ctx context.Context, connected model.Connected) error {
	userid := viewer.MustFromContext(ctx).UserID()
	if err := connectService.transactor.Run(ctx, func(txCtx context.Context) error {
		user, err := connectService.commonService.CommonGetUserByUserId(txCtx, userid)
		if err != nil {
			return err
		}

		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		// 检查连接地址是否为空
		if connected.ConnectURL == "" {
			return errors.New(commonModel.INVALID_CONNECTION_URL)
		}

		// 去除连接地址前后的空格和斜杠
		connected.ConnectURL = httpUtil.TrimURL(connected.ConnectURL)

		// 检查连接地址是否已存在
		connectedList, err := connectService.connectRepository.GetAllConnects(txCtx)
		if err != nil {
			return err
		}

		// 检查连接地址是否已存在
		for _, conn := range connectedList {
			if conn.ConnectURL == connected.ConnectURL {
				return errors.New(commonModel.CONNECT_HAS_EXISTS)
			}
		}

		// 添加连接地址
		if err := connectService.connectRepository.CreateConnect(txCtx, &connected); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	connectService.invalidateConnectsInfoCache()
	return nil
}

// DeleteConnect 删除连接
func (connectService *ConnectService) DeleteConnect(ctx context.Context, id string) error {
	userid := viewer.MustFromContext(ctx).UserID()
	if err := connectService.transactor.Run(ctx, func(txCtx context.Context) error {
		user, err := connectService.commonService.CommonGetUserByUserId(txCtx, userid)
		if err != nil {
			return err
		}

		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		// 删除连接地址
		if err := connectService.connectRepository.DeleteConnect(txCtx, id); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	connectService.invalidateConnectsInfoCache()
	return nil
}

// GetConnect 提供当前实例的连接信息
func (connectService *ConnectService) GetConnect() (model.Connect, error) {
	var connect model.Connect

	// 获取系统设置
	var setting settingModel.SystemSetting
	if err := connectService.settingService.GetSetting(&setting); err != nil {
		return connect, err
	}

	// 获取 owner 信息
	owner, err := connectService.commonService.GetOwner()
	if err != nil {
		return connect, err
	}

	// 统计当天发布的数量（优先复用进程本地时区，未识别时回退 UTC）
	siteTimezone := "UTC"
	if time.Local != nil {
		if localName := strings.TrimSpace(time.Local.String()); localName != "" {
			siteTimezone = timezoneUtil.NormalizeTimezone(localName)
		}
	}
	todayEchos := connectService.echoRepository.GetTodayEchos(true, siteTimezone)
	// 统计总发布数量
	_, totalEchos := connectService.echoRepository.GetEchosByPage(1, 1, "", true)

	// 设置 Connect 信息
	connect.ServerName = setting.ServerName
	connect.ServerURL = setting.ServerURL
	connect.TotalEchos = int(totalEchos)
	connect.TodayEchos = len(todayEchos)
	connect.SysUsername = owner.Username
	connect.Version = commonModel.Version

	trimmedServerURL := strings.TrimRight(setting.ServerURL, "/")
	logoPath := strings.TrimSpace(setting.ServerLogo)

	if logoPath == "" || logoPath == "Ech0.svg" || logoPath == "/Ech0.svg" {
		connect.Logo = fmt.Sprintf("%s/Ech0.svg", trimmedServerURL)
	} else if strings.HasPrefix(logoPath, "http://") || strings.HasPrefix(logoPath, "https://") {
		connect.Logo = logoPath
	} else if strings.HasPrefix(logoPath, "/") {
		connect.Logo = fmt.Sprintf("%s%s", trimmedServerURL, logoPath)
	} else {
		connect.Logo = fmt.Sprintf("%s/%s", trimmedServerURL, logoPath)
	}

	return connect, nil
}

// GetConnectsInfo 获取实例获取到的其它实例的连接信息
func (connectService *ConnectService) GetConnectsInfo() ([]model.Connect, error) {
	if cached, ok := connectService.getCachedConnectsInfo(); ok {
		return cached, nil
	}

	result, err, _ := connectService.connectsInfoFetcher.Do(connectsInfoSingleflightKey, func() (any, error) {
		// double-check，避免在等待 singleflight 期间其它请求已回填缓存
		if cached, ok := connectService.getCachedConnectsInfo(); ok {
			return cached, nil
		}

		connectList, fetchErr := connectService.fetchConnectsInfo()
		if fetchErr != nil {
			return nil, fetchErr
		}

		connectService.setCachedConnectsInfo(connectList)
		return cloneConnects(connectList), nil
	})
	if err != nil {
		return nil, err
	}

	connects, ok := result.([]model.Connect)
	if !ok {
		return nil, fmt.Errorf("invalid cache result type")
	}

	return cloneConnects(connects), nil
}

func (connectService *ConnectService) fetchConnectsInfo() ([]model.Connect, error) {
	// 总超时时间：给予足够的缓冲，避免单个慢连接导致整体超时
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 获取所有连接地址
	connects, err := connectService.connectRepository.GetAllConnects(context.Background())
	if err != nil {
		return nil, err
	}

	if len(connects) == 0 {
		return []model.Connect{}, nil
	}

	var connectList []model.Connect
	connectList = make([]model.Connect, 0, len(connects))

	var wg sync.WaitGroup
	connectChan := make(chan model.Connect, len(connects))
	semaphore := make(chan struct{}, connectFanoutMaxConcurrency)

	seenURLs := make(map[string]struct{})
	var seenMutex sync.Mutex

	// 重试配置：平衡速度和可靠性
	const maxRetries = 3
	const baseDelay = 1 * time.Second
	const requestTimeout = 3 * time.Second // 单个请求超时时间（降低到3秒，加快失败检测）

	for _, conn := range connects {
		wg.Add(1)
		go func(conn model.Connected) {
			defer wg.Done()
			select {
			case semaphore <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-semaphore }()

			url := httpUtil.TrimURL(conn.ConnectURL) + "/api/connect"

			var lastErr error
			for attempt := 0; attempt < maxRetries; attempt++ {
				select {
				case <-ctx.Done():
					logUtil.GetLogger().
						Info(
							"fetch connection info cancelled",
							zap.String("module", "connect"),
							zap.String("connect_url", conn.ConnectURL),
							zap.Error(ctx.Err()),
						)
					return // 总体超时直接退出
				default:
				}

				// 计算当前重试的延迟时间（指数退避）
				if attempt > 0 {
					delay := baseDelay * time.Duration(1<<(attempt-1)) // 1s, 2s, 4s...
					select {
					case <-time.After(delay):
					case <-ctx.Done():
						return
					}
				}

				resp, err := httpUtil.SendRequest(url, "GET", struct {
					Header  string
					Content string
				}{
					Header:  "Ech0_URL",
					Content: conn.ConnectURL,
				}, requestTimeout) // 传入自定义超时时间
				if err != nil {
					lastErr = err
					logUtil.GetLogger().Error("fetch connection info failed",
						zap.String("module", "connect"),
						zap.String("connect_url", conn.ConnectURL),
						zap.Int("attempt", attempt+1),
						zap.Error(err),
					)

					// 如果是最后一次重试，记录最终失败
					if attempt == maxRetries-1 {
						logUtil.GetLogger().Error("fetch connection info exhausted retries",
							zap.String("module", "connect"),
							zap.String("connect_url", conn.ConnectURL),
							zap.Int("retries", maxRetries),
							zap.Error(lastErr),
						)
					}
					continue
				}

				var connectInfo commonModel.Result[model.Connect]
				if err := json.Unmarshal(resp, &connectInfo); err != nil {
					lastErr = fmt.Errorf("JSON解析失败: %w", err)
					logUtil.GetLogger().Error("parse connection info failed",
						zap.String("module", "connect"),
						zap.String("connect_url", conn.ConnectURL),
						zap.Int("attempt", attempt+1),
						zap.Error(lastErr),
					)

					if attempt == maxRetries-1 {
						logUtil.GetLogger().Error("fetch connection info exhausted retries",
							zap.String("module", "connect"),
							zap.String("connect_url", conn.ConnectURL),
							zap.Int("retries", maxRetries),
							zap.Error(lastErr),
						)
					}
					continue
				}

				// 验证响应数据
				if connectInfo.Code != 1 {
					lastErr = fmt.Errorf("响应码无效: %d, 消息: %s", connectInfo.Code, connectInfo.Message)
					logUtil.GetLogger().Error("validate connection info failed",
						zap.String("module", "connect"),
						zap.String("connect_url", conn.ConnectURL),
						zap.Int("attempt", attempt+1),
						zap.Error(lastErr),
					)

					if attempt == maxRetries-1 {
						logUtil.GetLogger().Error("fetch connection info exhausted retries",
							zap.String("module", "connect"),
							zap.String("connect_url", conn.ConnectURL),
							zap.Int("retries", maxRetries),
							zap.Error(lastErr),
						)
					}
					continue
				}

				if connectInfo.Data.ServerURL == "" {
					lastErr = fmt.Errorf("服务器URL为空")
					logUtil.GetLogger().Error("validate connection info failed",
						zap.String("module", "connect"),
						zap.String("connect_url", conn.ConnectURL),
						zap.Int("attempt", attempt+1),
						zap.Error(lastErr),
					)

					if attempt == maxRetries-1 {
						logUtil.GetLogger().Error("fetch connection info exhausted retries",
							zap.String("module", "connect"),
							zap.String("connect_url", conn.ConnectURL),
							zap.Int("retries", maxRetries),
							zap.Error(lastErr),
						)
					}
					continue
				}

				// 成功获取有效数据，检查重复并发送
				seenMutex.Lock()
				if _, exists := seenURLs[connectInfo.Data.ServerURL]; exists {
					seenMutex.Unlock()
					logUtil.GetLogger().Info("connection info duplicated",
						zap.String("module", "connect"),
						zap.String("connect_url", conn.ConnectURL),
						zap.String("server_url", connectInfo.Data.ServerURL),
					)
					return // 重复数据，直接返回
				}
				seenURLs[connectInfo.Data.ServerURL] = struct{}{}
				seenMutex.Unlock()

				logUtil.GetLogger().Info("fetch connection info succeeded",
					zap.String("module", "connect"),
					zap.String("connect_url", conn.ConnectURL),
					zap.String("server_name", connectInfo.Data.ServerName),
				)
				connectChan <- connectInfo.Data
				return // 成功处理，退出重试循环
			}
		}(conn)
	}

	// 使用带缓冲的通道来避免goroutine泄漏
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(connectChan)
		close(done)
	}()

	// 在单独的 goroutine 中收集结果，使用 mutex 保护并发写入
	var mu sync.Mutex
	collectDone := make(chan struct{})
	go func() {
		for connect := range connectChan {
			if connect.ServerURL != "" {
				mu.Lock()
				connectList = append(connectList, connect)
				mu.Unlock()
			}
		}
		close(collectDone)
	}()

	// 等待完成或超时
	select {
	case <-done:
		// 所有 goroutine 完成，等待收集完成
		<-collectDone
		mu.Lock()
		count := len(connectList)
		mu.Unlock()
		logUtil.GetLogger().Info("collect connection info completed", zap.String("module", "connect"), zap.Int("valid_count", count))
	case <-ctx.Done():
		// 超时，等待收集器完成或超时
		logUtil.GetLogger().Info("collect connection info timeout, waiting collector", zap.String("module", "connect"))
		select {
		case <-collectDone:
			// 收集器已完成
			logUtil.GetLogger().Info("collector completed", zap.String("module", "connect"))
		case <-time.After(200 * time.Millisecond):
			// 给收集器额外的时间处理缓冲区中的数据
			logUtil.GetLogger().Info("collector timeout", zap.String("module", "connect"))
		}
		mu.Lock()
		count := len(connectList)
		mu.Unlock()
		logUtil.GetLogger().Info("collect connection info timeout completed", zap.String("module", "connect"), zap.Int("valid_count", count))
	}

	// 安全地返回结果
	mu.Lock()
	defer mu.Unlock()
	return connectList, nil
}

func (connectService *ConnectService) invalidateConnectsInfoCache() {
	connectService.connectsInfoCacheMu.Lock()
	defer connectService.connectsInfoCacheMu.Unlock()
	connectService.connectsInfoCache = nil
	connectService.connectsInfoCacheExpires = time.Time{}
	connectService.connectsInfoCacheValid = false
}

func (connectService *ConnectService) getCachedConnectsInfo() ([]model.Connect, bool) {
	connectService.connectsInfoCacheMu.RLock()
	defer connectService.connectsInfoCacheMu.RUnlock()

	if !connectService.connectsInfoCacheValid {
		return nil, false
	}
	if time.Now().After(connectService.connectsInfoCacheExpires) {
		return nil, false
	}
	return cloneConnects(connectService.connectsInfoCache), true
}

func (connectService *ConnectService) setCachedConnectsInfo(connects []model.Connect) {
	connectService.connectsInfoCacheMu.Lock()
	defer connectService.connectsInfoCacheMu.Unlock()
	connectService.connectsInfoCache = cloneConnects(connects)
	connectService.connectsInfoCacheExpires = time.Now().Add(connectsInfoCacheTTL)
	connectService.connectsInfoCacheValid = true
}

func cloneConnects(connects []model.Connect) []model.Connect {
	if len(connects) == 0 {
		return []model.Connect{}
	}
	cloned := make([]model.Connect, len(connects))
	copy(cloned, connects)
	return cloned
}

// GetConnects 获取当前实例添加的所有连接
func (connectService *ConnectService) GetConnects() ([]model.Connected, error) {
	// 获取所有连接地址
	connects, err := connectService.connectRepository.GetAllConnects(context.Background())
	if err != nil {
		return nil, err
	}

	// 如果没有找到，返回空切片
	if len(connects) == 0 {
		return []model.Connected{}, nil
	}

	// 返回查询到的 connects
	return connects, nil
}
