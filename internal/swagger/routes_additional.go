package swagger

// GetAgentRecent godoc
//
//	@Summary		获取近期动态
//	@Description	获取 Agent 近期动态列表
//	@Tags			系统设置
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/agent/recent [get]
func GetAgentRecent() {}

// GetInitStatus godoc
//
//	@Summary		获取初始化状态
//	@Description	获取系统初始化状态
//	@Tags			初始化
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/init/status [get]
func GetInitStatus() {}

// InitOwner godoc
//
//	@Summary		初始化 Owner
//	@Description	创建首个 Owner 账号
//	@Tags			初始化
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/init/owner [post]
func InitOwner() {}

// ListFiles godoc
//
//	@Summary		文件列表
//	@Description	分页获取文件列表
//	@Tags			文件管理
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/files [get]
func ListFiles() {}

// ListFileTree godoc
//
//	@Summary		文件树
//	@Description	获取文件树结构
//	@Tags			文件管理
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/file/tree [get]
func ListFileTree() {}

// StreamFileByPath godoc
//
//	@Summary		按路径流式读取文件
//	@Description	按路径读取并返回文件流
//	@Tags			文件管理
//	@Produce		octet-stream
//	@Success		200	{file}	file
//	@Router			/file/stream [get]
func StreamFileByPath() {}

// GetFileByID godoc
//
//	@Summary		按 ID 获取文件
//	@Description	通过 ID 查询文件元信息
//	@Tags			文件管理
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/file/{id} [get]
func GetFileByID() {}

// StreamFileByID godoc
//
//	@Summary		按 ID 流式读取文件
//	@Description	通过 ID 读取并返回文件流
//	@Tags			文件管理
//	@Produce		octet-stream
//	@Success		200	{file}	file
//	@Router			/file/{id}/stream [get]
func StreamFileByID() {}

// UploadFile godoc
//
//	@Summary		上传文件
//	@Description	上传文件到存储系统
//	@Tags			文件管理
//	@Accept			multipart/form-data
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/files/upload [post]
func UploadFile() {}

// CreateExternalFile godoc
//
//	@Summary		创建外链文件
//	@Description	创建外部 URL 文件记录
//	@Tags			文件管理
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/files/external [post]
func CreateExternalFile() {}

// DeleteFile godoc
//
//	@Summary		删除文件
//	@Description	根据 ID 删除文件
//	@Tags			文件管理
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/file/{id} [delete]
func DeleteFile() {}

// GetFilePresignURL godoc
//
//	@Summary		获取文件预签名 URL
//	@Description	获取文件直传预签名 URL
//	@Tags			文件管理
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/files/presign [put]
func GetFilePresignURL() {}

// GetCommentsForm godoc
//
//	@Summary		获取评论表单配置
//	@Description	获取评论发布所需表单元数据
//	@Tags			评论
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/comments/form [get]
func GetCommentsForm() {}

// ListComments godoc
//
//	@Summary		获取评论列表
//	@Description	按动态 ID 获取评论列表
//	@Tags			评论
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/comments [get]
func ListComments() {}

// ListPublicComments godoc
//
//	@Summary		获取公开评论
//	@Description	获取公开评论流
//	@Tags			评论
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/comments/public [get]
func ListPublicComments() {}

// CreateComment godoc
//
//	@Summary		创建评论
//	@Description	提交新评论
//	@Tags			评论
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/comments [post]
func CreateComment() {}

// ListPanelComments godoc
//
//	@Summary		评论管理列表
//	@Description	管理后台获取评论列表
//	@Tags			评论管理
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/panel/comments [get]
func ListPanelComments() {}

// GetPanelCommentByID godoc
//
//	@Summary		获取评论详情
//	@Description	管理后台按 ID 获取评论详情
//	@Tags			评论管理
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/panel/comments/{id} [get]
func GetPanelCommentByID() {}

// UpdatePanelCommentStatus godoc
//
//	@Summary		更新评论状态
//	@Description	管理后台更新评论状态
//	@Tags			评论管理
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/panel/comments/{id}/status [patch]
func UpdatePanelCommentStatus() {}

// UpdatePanelCommentHot godoc
//
//	@Summary		更新评论置顶
//	@Description	管理后台更新评论热度/置顶状态
//	@Tags			评论管理
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/panel/comments/{id}/hot [patch]
func UpdatePanelCommentHot() {}

// DeletePanelComment godoc
//
//	@Summary		删除评论
//	@Description	管理后台删除评论
//	@Tags			评论管理
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/panel/comments/{id} [delete]
func DeletePanelComment() {}

// BatchPanelComments godoc
//
//	@Summary		批量操作评论
//	@Description	管理后台批量操作评论
//	@Tags			评论管理
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/panel/comments/batch [post]
func BatchPanelComments() {}

// GetPanelCommentSettings godoc
//
//	@Summary		获取评论设置
//	@Description	管理后台获取评论设置
//	@Tags			评论管理
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/panel/comments/settings [get]
func GetPanelCommentSettings() {}

// UpdatePanelCommentSettings godoc
//
//	@Summary		更新评论设置
//	@Description	管理后台更新评论设置
//	@Tags			评论管理
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/panel/comments/settings [put]
func UpdatePanelCommentSettings() {}

// TestPanelCommentEmail godoc
//
//	@Summary		测试评论邮件
//	@Description	管理后台发送评论邮件测试
//	@Tags			评论管理
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/panel/comments/settings/test-email [post]
func TestPanelCommentEmail() {}

// GetWebsiteTitle godoc
//
//	@Summary		获取网站标题
//	@Description	根据 URL 获取网站标题
//	@Tags			通用功能
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/website/title [get]
func GetWebsiteTitle() {}

// GetRss godoc
//
//	@Summary		获取 RSS
//	@Description	获取 RSS/Atom 订阅源
//	@Tags			通用功能
//	@Produce		application/rss+xml
//	@Success		200	{string}	string
//	@Router			/rss [get]
func GetRss() {}

// Healthz godoc
//
//	@Summary		健康检查
//	@Description	服务健康检查接口
//	@Tags			系统
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/healthz [get]
func Healthz() {}

// GetSystemLogs godoc
//
//	@Summary		系统日志
//	@Description	获取系统日志列表
//	@Tags			系统
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/system/logs [get]
func GetSystemLogs() {}

// StreamSystemLogs godoc
//
//	@Summary		系统日志流
//	@Description	SSE 方式订阅系统日志
//	@Tags			系统
//	@Produce		text/event-stream
//	@Success		200	{string}	string
//	@Router			/system/logs/stream [get]
func StreamSystemLogs() {}

// MigrationUpload godoc
//
//	@Summary		上传迁移包
//	@Description	上传迁移源文件
//	@Tags			迁移
//	@Accept			multipart/form-data
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/migration/upload [post]
func MigrationUpload() {}

// MigrationStart godoc
//
//	@Summary		开始迁移
//	@Description	启动迁移任务
//	@Tags			迁移
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/migration/start [post]
func MigrationStart() {}

// MigrationStatus godoc
//
//	@Summary		迁移状态
//	@Description	查询当前迁移任务状态
//	@Tags			迁移
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/migration/status [get]
func MigrationStatus() {}

// MigrationCancel godoc
//
//	@Summary		取消迁移
//	@Description	取消当前迁移任务
//	@Tags			迁移
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/migration/cancel [post]
func MigrationCancel() {}

// MigrationCleanup godoc
//
//	@Summary		清理迁移
//	@Description	清理迁移临时文件与状态
//	@Tags			迁移
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/migration/cleanup [post]
func MigrationCleanup() {}
