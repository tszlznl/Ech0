declare namespace App {
  /**
   * Namespace Api
   */
  namespace Api {
    type Response<T> = {
      code: number
      msg: string
      error_code?: string
      data: T
    }

    namespace Auth {
      type LoginParams = {
        username: string
        password: string
      }

      type LoginResponse = {
        token: string
      }

      type SignupParams = {
        username: string
        password: string
      }

      // Passkey / WebAuthn
      type PasskeyRegisterBeginResp = {
        nonce: string
        publicKey: unknown
      }

      type PasskeyLoginBeginResp = {
        nonce: string
        publicKey: unknown
      }

      type PasskeyDevice = {
        id: string
        device_name: string
        aaguid: string
        last_used_at: string
        created_at: string
      }
    }

    namespace User {
      type User = {
        id: string
        username: string
        password?: string
        is_admin: boolean
        is_owner?: boolean
        avatar?: string
      }

      type UserInfo = {
        username: string
        password: string
        is_admin: boolean
        is_owner?: boolean
        avatar: string
      }

      type UserStatus = {
        user_id: string
        username: string
        is_admin: boolean
        is_owner?: boolean
      }
    }

    namespace File {
      type Category = import('@/constants/file').FileCategory
      type StorageType = import('@/constants/file').FileStorageType

      type FileDto = {
        id: string
        key: string
        url: string
        content_type?: string
        category?: Category
        storage_type?: StorageType
        size?: number
        width?: number
        height?: number
      }
      type FileDeleteDto = {
        id: string
      }
      type CreateExternalFileDto = {
        url: string
        content_type?: string
        category?: Category
        width?: number
        height?: number
        name?: string
      }
      type UpdateFileMetaDto = {
        size: number
        width?: number
        height?: number
        content_type?: string
      }
    }

    namespace Ech0 {
      type ParamsByPagination = {
        page: number
        pageSize: number
        search?: string
      }

      type Echo = {
        id: string
        content: string
        username: string
        echo_files?: EchoFile[]
        layout?: string
        private: boolean
        user_id: string
        extension?: string
        extension_type?: string
        tags?: Tag[]
        fav_count: number
        created_at: string
      }

      type FileObject = {
        id: string
        echo_id: string
        url: string
        storage_type: File.StorageType
        category?: File.Category
        content_type?: string
        key?: string // 对应后端 file.key
        size?: number // 文件大小（字节）
        width?: number // 图片宽度
        height?: number // 图片高度
      }

      type Tag = {
        id: string
        name: string
        usage_count: number
        created_at: string
      }

      type EchoFile = {
        id: string
        echo_id: string
        file_id: string
        sort_order: number
        file?: {
          id: string
          key: string
          storage_type: File.StorageType
          provider?: string
          bucket?: string
          url: string
          name?: string
          content_type?: string
          size?: number
          category?: File.Category
          user_id?: string
          width?: number
          height?: number
          created_at?: string
        }
      }

      type FileToAdd = {
        id?: string
        url: string
        storage_type: File.StorageType
        category?: File.Category
        content_type?: string
        key?: string // 对应后端 file.key
        size?: number // 文件大小（字节）
        width?: number // 图片宽度
        height?: number // 图片高度
      }

      type TagToAdd = {
        id?: string
        name: string
        usage_count?: number
        created_at?: string
      }

      type EchoToAdd = {
        content: string
        echo_files?: Array<{ file_id: string; sort_order: number }> | null
        tags?: TagToAdd[] | null
        layout?: string | null
        extension?: string | null
        extension_type?: string | null
        private: boolean
      }

      type EchoToUpdate = {
        id: string
        content: string
        username: string
        echo_files?: Array<{ file_id: string; sort_order: number }> | null
        tags?: TagToAdd[] | null
        layout?: string | null
        private: boolean
        user_id: string
        extension?: string | null
        extension_type?: string | null
        created_at: string
      }

      type PaginationResult = {
        items: Echo[]
        total: number
      }

      type Status = {
        owner_id: string // Owner ID
        username: string // 系统管理员用户名
        logo: string // 系统管理员Logo
        users: App.Api.User.UserStatus[] // 用户列表
        total_echos: number // Echo总数
      }

      type HeatMap = {
        date: string
        count: number
      }[]

      type FileToDelete = {
        id: string
      }

      type GithubCardData = {
        name: string
        stargazers_count: number
        forks_count: number
        description: string
        owner: {
          avatar_url: string
        }
      }

      type HelloEch0 = {
        hello: string
        version: string
        github: string
      }

      type PresignResult = {
        id: string
        file_name: string
        content_type: string
        key: string
        presign_url: string
        file_url: string
      }
    }

    namespace Setting {
      type SystemSetting = {
        site_title: string
        server_logo: string
        server_name: string
        server_url: string
        allow_register: boolean
        ICP_number: string
        meting_api: string
        custom_css: string
        custom_js: string
      }

      type CommentSetting = {
        enable_comment: boolean
        provider: string // 评论提供者
        comment_api: string // 评论 API 地址
      }

      type S3Setting = {
        enable: boolean
        provider: string
        endpoint: string
        access_key: string
        secret_key: string
        bucket_name: string
        region: string
        use_ssl: boolean
        cdn_url: string
        path_prefix: string
        public_read: boolean
      }

      type OAuth2Setting = {
        enable: boolean
        provider: string
        client_id: string
        client_secret: string
        redirect_uri: string
        scopes: string[]
        auth_url: string
        token_url: string
        user_info_url: string

        is_oidc: boolean
        issuer: string
        jwks_url: string
      }

      type OAuth2Status = {
        enabled: boolean
        provider: string
      }

      type OAuthInfo = {
        provider: string
        user_id: string
        oauth_id: string
        issuer: string
        auth_type: string
      }

      type Webhook = {
        id: string
        name: string
        url: string
        is_active: boolean
        last_status: string
        last_trigger: string
        created_at: string
        updated_at: string
      }

      type WebhookDto = {
        name: string
        url: string
        secret?: string
        is_active: boolean
      }

      type AccessToken = {
        id: string
        user_id: string
        name: string
        expiry: string | null
        created_at: string
      }

      type AccessTokenDto = {
        name: string
        expiry: string
      }

      type BackupSchedule = {
        enable: boolean
        cron_expression: string
      }

      type BackupScheduleDto = {
        enable: boolean
        cron_expression: string
      }

      type AgentSetting = {
        enable: boolean
        provider: string
        model: string
        api_key: string
        prompt: string
        base_url: string
      }

      type AgentSettingDto = {
        enable: boolean
        provider: string
        model: string
        api_key: string
        prompt: string
        base_url: string
      }
    }

    namespace Init {
      type Status = {
        initialized: boolean
        owner_exists: boolean
      }
    }

    namespace Connect {
      type Connect = {
        server_name: string
        server_url: string
        logo: string
        total_echos: number
        today_echos: number
        sys_username: string
      }

      type Connected = {
        id: string
        connect_url: string
      }
    }

    namespace Todo {
      type Todo = {
        id: string
        content: string
        user_id: string
        username: string
        status: number
        created_at: string
      }

      type TodoToAdd = {
        content: string
      }
    }

    namespace Dashboard {
      // CpuMetric cpu监控指标
      type CpuMetric = {
        UsagePercent: number // CPU 使用率百分比
        Cores: number // CPU 核心数
        FrequencyMHz: number // CPU 主频，单位 MHz
      }
      // MemoryMetric 内存监控指标
      type MemoryMetric = {
        Total: number // 总内存大小
        Used: number // 已使用内存大小
        Available: number // 可用内存大小
        Percentage: number // 内存使用率百分比
      }
      // DiskMetric 磁盘监控指标
      type DiskMetric = {
        Total: number // 磁盘总大小
        Used: number // 已使用磁盘大小
        Available: number // 可用磁盘大小
        Percentage: number // 磁盘使用率百分比
      }

      // NetworkMetric 网络监控指标
      type NetworkMetric = {
        TotalBytesSent: number // 总发送字节数
        TotalBytesReceived: number // 总接收字节数
        BytesSentPerSecond: number // 每秒发送字节数 (B/s)
        BytesReceivedPerSecond: number // 每秒接收字节数 (B/s)
      }

      // SystemMetric 系统监控指标
      type SystemMetric = {
        Hostname: string // 主机名
        OsName: string // 操作系统名称
        Uptime: number // 系统运行时长
        KernelVersion: string // 内核版本
        KernelArch: string // 内核架构
        Time: string // 采样时间
        TimeZone: string // 采样时区
        ProcessCount: number // 当前进程数
        ThreadCount: number // 当前线程数
        GolangVersion: string // Golang 版本
        GoRoutineCount: number // 当前 Goroutine 数量
      }

      // Metrics 综合监控指标
      type Metrics = {
        CPU: CpuMetric // CPU 监控指标
        Memory: MemoryMetric // 内存监控指标
        Disk: DiskMetric // 磁盘监控指标
        Network: NetworkMetric // 网络监控指标
        System: SystemMetric // 系统监控指标
      }
    }

    namespace SystemLog {
      type Entry = {
        time: string
        level: string
        msg: string
        module?: string
        caller?: string
        error?: string
        raw?: string
        fields?: Record<string, unknown>
      }

      type QueryParams = {
        tail?: number
        level?: string
        keyword?: string
      }
    }

    namespace Hub {
      type HubItem = string | { id: string; connect_url: string }
      type HubList = HubItem[]
      type HubItemInfo = Connect.Connect
      type HubInfoList = HubItemInfo[]

      type Echo = {
        id: string
        content: string
        username: string
        echo_files?: Ech0.EchoFile[]
        tags?: Tag[]
        layout?: string
        private: boolean
        user_id: string
        extension: string
        extension_type: string
        fav_count: number
        created_at: string
        createdTs: number
        virtual_key: string
        server_name: string
        server_url: string
        logo: string
      }
    }

    namespace Inbox {
      type Inbox = {
        id: string
        source: string
        content: string
        type: string
        read: boolean
        read_count: number
        meta?: string
        read_at?: number // Unix时间戳（秒）
        created_at: number // Unix时间戳（秒）
      }

      type InboxListResult = {
        items: Inbox[]
        total: number
      }

      type InboxListParams = {
        page: number
        pageSize: number
        search?: string
      }
    }
  }
}
