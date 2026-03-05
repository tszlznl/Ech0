package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"log"
	"os"
	"strconv"
	"sync"
)

var (
	cfg  *AppConfig
	once sync.Once
)

type AppConfig struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Upload   UploadConfig
	Setting  SettingConfig
	Comment  CommentConfig
	Security SecurityConfig
}

type ServerConfig struct {
	Port string // 服务器端口
	Host string // 服务器主机地址
	Mode string // 运行模式，可能的值为 "debug" 或 "release"
}

type DatabaseConfig struct {
	Type    string // 数据库类型
	Path    string // 数据库文件路径
	LogMode string // 数据库日志模式
}

type AuthConfig struct {
	Jwt JWTConfig
}

type JWTConfig struct {
	Expires  int    // JWT的过期时间，单位为秒
	Issuer   string // JWT的发行者
	Audience string // JWT的受众
}

type UploadConfig struct {
	ImageMaxSize int      // 图片文件的最大上传大小，单位为字节
	AudioMaxSize int      // 音频文件的最大上传大小，单位为字节
	AllowedTypes []string // 允许上传的文件类型
	ImagePath    string   // 图片文件存储路径
	AudioPath    string   // 音频文件存储路径
}

type SettingConfig struct {
	SiteTitle     string // 网站标题
	ServerLogo    string // 服务器Logo
	Servername    string // 服务器名称
	Serverurl     string // 服务器 URL
	AllowRegister bool   // 是否允许注册
	Icpnumber     string // ICP 备案号
	MetingAPI     string // Meting API 地址
	CustomCSS     string // 自定义 CSS 样式
	CustomJS      string // 自定义 JS 脚本
}

type CommentConfig struct {
	EnableComment bool   // 是否启用评论
	Provider      string // 评论提供者
	CommentAPI    string // 评论 API 地址
}

type SecurityConfig struct {
	JWTSecret     []byte
	RSAPrivate    *rsa.PrivateKey
	RSAPrivateKey []byte
	RSAPublic     *rsa.PublicKey
	RSAPublicKey  []byte
}

// Config 返回全局配置中心
func Config() *AppConfig {
	once.Do(func() {
		cfg = defaultConfig()
		applyEnvOverrides(cfg)
		cfg.Security.JWTSecret = getJWTSecret()
		genSecretKey(cfg)
	})
	return cfg
}

func defaultConfig() *AppConfig {
	return &AppConfig{
		Server: ServerConfig{
			Port: "6277",
			Host: "0.0.0.0",
			Mode: "release",
		},
		Database: DatabaseConfig{
			Type:    "sqlite",
			Path:    "data/ech0.db",
			LogMode: "release",
		},
		Auth: AuthConfig{
			Jwt: JWTConfig{
				Expires:  2592000,
				Issuer:   "ech0",
				Audience: "ech0",
			},
		},
		Upload: UploadConfig{
			ImageMaxSize: 20971520,
			AudioMaxSize: 20971520,
			ImagePath:    "data/images/",
			AudioPath:    "data/audios/",
			AllowedTypes: []string{
				"image/jpeg",
				"image/png",
				"image/gif",
				"image/webp",
				"image/svg+xml",
				"image/avif",
				"audio/mpeg",
				"audio/flac",
				"audio/wav",
				"audio/mp4",
			},
		},
		Setting: SettingConfig{
			SiteTitle:     "Ech0",
			ServerLogo:    "/Ech0.svg",
			Servername:    "Ech0",
			Serverurl:     "https://ech0.example.com",
			AllowRegister: true,
			Icpnumber:     "",
			MetingAPI:     "",
			CustomCSS:     "",
			CustomJS:      "",
		},
		Comment: CommentConfig{
			EnableComment: false,
			Provider:      "twikoo",
			CommentAPI:    "",
		},
	}
}

func applyEnvOverrides(cfg *AppConfig) {
	setStringEnv("ECH0_SERVER_PORT", &cfg.Server.Port)
	setStringEnv("ECH0_SERVER_HOST", &cfg.Server.Host)
	setStringEnv("ECH0_SERVER_MODE", &cfg.Server.Mode)
	setStringEnv("ECH0_DB_TYPE", &cfg.Database.Type)
	setStringEnv("ECH0_DB_PATH", &cfg.Database.Path)
	setStringEnv("ECH0_DB_LOGMODE", &cfg.Database.LogMode)
	setStringEnv("ECH0_UPLOAD_IMAGE_PATH", &cfg.Upload.ImagePath)
	setStringEnv("ECH0_UPLOAD_AUDIO_PATH", &cfg.Upload.AudioPath)
	setStringEnv("ECH0_SERVER_URL", &cfg.Setting.Serverurl)
	setIntEnv("ECH0_JWT_EXPIRES", &cfg.Auth.Jwt.Expires)
}

func setStringEnv(key string, target *string) {
	if value := os.Getenv(key); value != "" {
		*target = value
	}
}

func setIntEnv(key string, target *int) {
	value := os.Getenv(key)
	if value == "" {
		return
	}
	parsed, err := strconv.Atoi(value)
	if err == nil {
		*target = parsed
	}
}

// getJWTSecret 加载JWT密钥
func getJWTSecret() []byte {
	// 从环境变量中获取JWT密钥
	secret := os.Getenv("JWT_SECRET")
	if secret == "" { // 如果没有设置环境变量，则使用UUID生成默认密钥
		b := make([]byte, 16)
		_, err := rand.Read(b)
		if err != nil {
			log.Fatal("failed to generate random JWT secret:", err)
		}
		secret = hex.EncodeToString(b)
	}

	return []byte(secret)
}

// genSecretKey 生成用于联邦架构的密钥对，并保存到本地文件
func genSecretKey(cfg *AppConfig) {
	const (
		keyDir     = "data/keys"
		privateKey = "private.pem"
		publicKey  = "public.pem"
	)
	// 检查密钥文件是否已经存在
	if _, err := os.Stat(keyDir); os.IsNotExist(err) {
		// 创建存放密钥的目录
		if err := os.Mkdir(keyDir, 0o700); err != nil {
			log.Fatalf("Failed to create key directory: %v", err)
		}
	}

	genFlag := false
	if _, err := os.Stat(keyDir + "/" + privateKey); err != nil {
		log.Println("Private key not found, generating new key pair.")
		genFlag = true
	}

	if _, err := os.Stat(keyDir + "/" + publicKey); err != nil {
		log.Println("Public key not found, generating new key pair.")
		genFlag = true
	}

	if genFlag {
		//  2048 位 RSA 私钥
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Fatalf("Failed to generate private key: %v", err)
		}

		// 保存私钥到文件
		privBytes := x509.MarshalPKCS1PrivateKey(priv)
		privPem := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privBytes,
		})
		if err := os.WriteFile(keyDir+"/"+privateKey, privPem, 0o600); err != nil {
			log.Fatalf("Failed to write private key: %v", err)
		}

		// 保存公钥到文件
		pub := &priv.PublicKey
		pubBytes, err := x509.MarshalPKIXPublicKey(pub)
		if err != nil {
			log.Fatalf("Failed to marshal public key: %v", err)
		}
		pubPem := pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubBytes,
		})
		if err := os.WriteFile(keyDir+"/"+publicKey, pubPem, 0o644); err != nil {
			log.Fatalf("Failed to write public key: %v", err)
		}

		log.Println("Generated RSA key pair and saved to private.pem and public.pem")
		cfg.Security.RSAPrivateKey = privPem
		cfg.Security.RSAPrivate = priv
		cfg.Security.RSAPublicKey = pubPem
		cfg.Security.RSAPublic = pub
	} else {
		// 读取现有的密钥文件
		privPem, err := os.ReadFile(keyDir + "/" + privateKey)
		if err == nil {
			block, _ := pem.Decode(privPem)
			if block == nil || block.Type != "RSA PRIVATE KEY" {
				log.Fatal("Failed to decode PEM block containing private key")
			}
			priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				log.Fatalf("Failed to parse private key: %v", err)
			}
			cfg.Security.RSAPrivate = priv
			cfg.Security.RSAPrivateKey = privPem
		} else {
			log.Println("Private key not found, generating new key pair.")
		}
		// 读取公钥文件
		pubPem, err := os.ReadFile(keyDir + "/" + publicKey)
		if err == nil {
			block, _ := pem.Decode(pubPem)
			if block == nil || block.Type != "PUBLIC KEY" {
				log.Fatal("Failed to decode PEM block containing public key")
			}
			pub, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				log.Fatalf("Failed to parse public key: %v", err)
			}
			rsaPub, ok := pub.(*rsa.PublicKey)
			if !ok {
				log.Fatal("Public key is not an RSA public key")
			}
			cfg.Security.RSAPublic = rsaPub
			cfg.Security.RSAPublicKey = pubPem
		} else {
			log.Println("Public key not found, generating new key pair.")
		}
	}
}
