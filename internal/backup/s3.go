package backup

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	virefs "github.com/lin-snow/VireFS"
	"github.com/lin-snow/ech0/internal/config"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

const (
	s3BackupPrefix   = "backups/"
	s3BackupKeepCount = 3
)

// BuildBackupS3FS creates a VireFS ObjectFS dedicated to backup operations.
// It respects the user's PathPrefix but omits Schema (no file-type classification),
// so backups are stored at <prefix>/backups/ech0_backup_xxx.zip.
func BuildBackupS3FS(cfg config.StorageConfig) (virefs.FS, error) {
	if !cfg.ObjectEnabled {
		return nil, fmt.Errorf("object storage is not enabled")
	}

	provider := mapS3Provider(cfg.Provider)
	region := resolveS3Region(cfg.Provider, cfg.Region)
	endpoint := normalizeS3Endpoint(cfg.Endpoint, cfg.UseSSL)

	var opts []virefs.ObjectOption
	if cfg.PathPrefix != "" {
		opts = append(opts, virefs.WithPrefix(strings.Trim(cfg.PathPrefix, "/")+"/"))
	}

	fs, err := virefs.NewObjectFSFromConfig(context.Background(), &virefs.S3Config{
		Provider:  provider,
		Endpoint:  endpoint,
		Region:    region,
		Bucket:    cfg.BucketName,
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
	}, opts...)
	if err != nil {
		return nil, fmt.Errorf("create backup s3 fs: %w", err)
	}
	return fs, nil
}

// UploadToS3 uploads the local backup ZIP to S3 at backups/<fileName>,
// then cleans up old backups keeping only the most recent ones.
func UploadToS3(ctx context.Context, backupFilePath, fileName string, s3FS virefs.FS) error {
	f, err := os.Open(backupFilePath)
	if err != nil {
		return fmt.Errorf("open backup file for s3 upload: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			logUtil.GetLogger().Warn("Failed to close backup file after s3 upload",
				zap.String("path", backupFilePath), zap.Error(closeErr))
		}
	}()

	s3Key := s3BackupPrefix + fileName
	if err := s3FS.Put(ctx, s3Key, f); err != nil {
		return fmt.Errorf("put backup to s3: %w", err)
	}

	logUtil.GetLogger().Info("Backup uploaded to S3",
		zap.String("key", s3Key))

	if err := cleanupOldS3Backups(ctx, s3FS, s3BackupKeepCount); err != nil {
		logUtil.GetLogger().Warn("Failed to cleanup old S3 backups",
			zap.Error(err))
	}

	return nil
}

// cleanupOldS3Backups lists files under the backups/ prefix and removes
// all but the most recent keepCount files (sorted by name, which embeds
// a UTC timestamp).
func cleanupOldS3Backups(ctx context.Context, s3FS virefs.FS, keepCount int) error {
	result, err := s3FS.List(ctx, s3BackupPrefix)
	if err != nil {
		return fmt.Errorf("list s3 backups: %w", err)
	}

	var backupFiles []string
	for _, item := range result.Files {
		key := strings.Trim(item.Key, "/")
		if item.IsDir || key == "" {
			continue
		}
		if strings.HasPrefix(key, s3BackupPrefix) && strings.HasSuffix(key, ".zip") {
			backupFiles = append(backupFiles, key)
		}
	}

	if len(backupFiles) <= keepCount {
		return nil
	}

	sort.Strings(backupFiles)

	toDelete := backupFiles[:len(backupFiles)-keepCount]
	for _, key := range toDelete {
		if err := s3FS.Delete(ctx, key); err != nil {
			logUtil.GetLogger().Warn("Failed to delete old S3 backup",
				zap.String("key", key), zap.Error(err))
		} else {
			logUtil.GetLogger().Info("Deleted old S3 backup",
				zap.String("key", key))
		}
	}

	return nil
}

func mapS3Provider(raw string) virefs.Provider {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "minio":
		return virefs.ProviderMinIO
	case "r2":
		return virefs.ProviderR2
	default:
		return virefs.ProviderAWS
	}
}

func resolveS3Region(providerRaw string, regionRaw string) string {
	region := strings.TrimSpace(regionRaw)
	if region != "" {
		return region
	}
	switch strings.ToLower(strings.TrimSpace(providerRaw)) {
	case "r2", "other":
		return "auto"
	default:
		return "us-east-1"
	}
}

func normalizeS3Endpoint(endpoint string, useSSL bool) string {
	if endpoint == "" {
		return endpoint
	}
	lower := strings.ToLower(endpoint)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return endpoint
	}
	if useSSL {
		return "https://" + endpoint
	}
	return "http://" + endpoint
}
