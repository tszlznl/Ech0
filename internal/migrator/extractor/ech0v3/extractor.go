package ech0v3

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/migrator/spec"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/lin-snow/ech0/internal/storage"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Extractor struct{}

const (
	echoLogSampleStep = int64(50)
	dbBatchSize       = 200
)

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Extract(_ context.Context, req spec.ExtractRequest) (spec.ExtractResult, error) {
	items, ok := req.SourcePayload["items"].([]any)
	if !ok || len(items) == 0 {
		return spec.ExtractResult{
			Records:        []spec.RawRecord{},
			NextCheckpoint: req.Checkpoint,
			HasMore:        false,
			TotalHint:      0,
		}, nil
	}

	start := int(req.Checkpoint)
	if start < 0 {
		start = 0
	}
	if start >= len(items) {
		return spec.ExtractResult{
			Records:        []spec.RawRecord{},
			NextCheckpoint: int64(len(items)),
			HasMore:        false,
			TotalHint:      int64(len(items)),
		}, nil
	}

	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}
	end := start + batchSize
	if end > len(items) {
		end = len(items)
	}

	records := make([]spec.RawRecord, 0, end-start)
	for i := start; i < end; i++ {
		obj, ok := items[i].(map[string]any)
		if !ok {
			return spec.ExtractResult{}, fmt.Errorf("ech0 v3 item at index %d is not object", i)
		}
		if _, exists := obj["content"]; !exists {
			if text, ok := obj["text"]; ok {
				obj["content"] = text
			}
		}
		sourceID := fmt.Sprintf("%v", obj["id"])
		if sourceID == "<nil>" {
			sourceID = fmt.Sprintf("ech0v3-%d", i)
		}
		records = append(records, spec.RawRecord{
			SourceID: sourceID,
			Data:     obj,
		})
	}

	return spec.ExtractResult{
		Records:        records,
		NextCheckpoint: int64(end),
		HasMore:        end < len(items),
		TotalHint:      int64(len(items)),
	}, nil
}

func (e *Extractor) Migrate(ctx context.Context, req spec.MigrateRequest) (spec.MigrateResult, error) {
	options := parseOptions(req.SourcePayload)
	sourceDBPath, sourceRoot, err := resolveSourceDBPath(req.SourcePayload)
	if err != nil {
		return spec.MigrateResult{}, err
	}

	sourceDB, err := gorm.Open(sqlite.Open(sourceDBPath), &gorm.Config{})
	if err != nil {
		return spec.MigrateResult{}, fmt.Errorf("open source sqlite: %w", err)
	}

	var total int64
	if err := sourceDB.Table("echos").Count(&total).Error; err != nil {
		return spec.MigrateResult{}, fmt.Errorf("count source echos: %w", err)
	}

	jobID := uuidUtil.MustNewV7()
	report := map[string]any{
		"job_id":            jobID,
		"source_db":         sourceDBPath,
		"include_tags":      true,
		"include_images":    true,
		"failure_threshold": options.FailureThreshold,
		"processed":         int64(0),
		"success_count":     int64(0),
		"fail_count":        int64(0),
		"failed_items":      []spec.FailedItem{},
		"echo_id_map":       map[string]string{},
	}
	logUtil.GetLogger().Info("migration ech0_v3 started",
		zap.String("module", "migration"),
		zap.String("source_db", sourceDBPath),
		zap.String("job_id", jobID),
		zap.Int64("total", total),
	)

	if req.UpdateProgress != nil {
		req.UpdateProgress(spec.MigrateProgress{
			CurrentPhase: "extracting",
			Processed:    0,
			Total:        total,
		})
	}

	migrateErr := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var sourceEchos []v3Echo
		if err := sourceDB.WithContext(ctx).Table("echos").Order("id ASC").Find(&sourceEchos).Error; err != nil {
			return fmt.Errorf("query source echos: %w", err)
		}
		echoIDsWithImages, err := loadEchoIDsWithImages(sourceDB)
		if err != nil {
			return fmt.Errorf("load image-linked echo ids: %w", err)
		}

		idMap := make(map[int64]string, len(sourceEchos))
		failed := make([]spec.FailedItem, 0)
		successCount := int64(0)
		failCount := int64(0)

		for i := range sourceEchos {
			if err := ctx.Err(); err != nil {
				return err
			}
			row := sourceEchos[i]
			content := strings.TrimSpace(row.Content)
			hasContent := content != ""
			hasExtension := strings.TrimSpace(row.Extension) != ""
			_, hasFiles := echoIDsWithImages[row.ID]

			if !hasContent && hasExtension {
				content = "[迁移占位] 原记录仅包含扩展内容"
			}
			if !hasContent && !hasExtension && hasFiles {
				content = "[迁移占位] 原记录仅包含图片/附件内容"
			}
			if !hasContent && !hasExtension && !hasFiles {
				failCount++
				failed = append(failed, spec.FailedItem{
					SourceID: strconv.FormatInt(row.ID, 10),
					Reason:   "empty content, files and extension",
				})
				logUtil.GetLogger().Warn("migration echo skipped",
					zap.String("module", "migration"),
					zap.String("job_id", jobID),
					zap.Int64("source_echo_id", row.ID),
					zap.String("reason", "empty content, files and extension"),
				)
				if exceedsFailureThreshold(successCount+failCount, failCount, options.FailureThreshold) {
					return fmt.Errorf("failure rate exceeded %.2f%%, rollback job %s", options.FailureThreshold*100, jobID)
				}
				continue
			}

			echo := echoModel.Echo{
				Content:   content,
				Username:  normalizeUsername(row.Username),
				Layout:    normalizeLayout(row.Layout),
				Private:   row.Private,
				UserID:    options.CreatedBy,
				FavCount:  row.FavCount,
				CreatedAt: normalizeTime(row.CreatedAt),
			}
			if err := tx.Create(&echo).Error; err != nil {
				failCount++
				failed = append(failed, spec.FailedItem{
					SourceID: strconv.FormatInt(row.ID, 10),
					Reason:   err.Error(),
				})
				logUtil.GetLogger().Warn("migration echo create failed",
					zap.String("module", "migration"),
					zap.String("job_id", jobID),
					zap.Int64("source_echo_id", row.ID),
					zap.String("reason", err.Error()),
				)
				if exceedsFailureThreshold(successCount+failCount, failCount, options.FailureThreshold) {
					return fmt.Errorf("failure rate exceeded %.2f%%, rollback job %s", options.FailureThreshold*100, jobID)
				}
				continue
			}

			if err := persistExtension(tx, echo.ID, row.ExtensionType, row.Extension); err != nil {
				failCount++
				failed = append(failed, spec.FailedItem{
					SourceID: strconv.FormatInt(row.ID, 10),
					Reason:   fmt.Sprintf("save extension: %v", err),
				})
				logUtil.GetLogger().Warn("migration echo extension failed",
					zap.String("module", "migration"),
					zap.String("job_id", jobID),
					zap.Int64("source_echo_id", row.ID),
					zap.String("target_echo_id", echo.ID),
					zap.String("reason", err.Error()),
				)
				if exceedsFailureThreshold(successCount+failCount, failCount, options.FailureThreshold) {
					return fmt.Errorf("failure rate exceeded %.2f%%, rollback job %s", options.FailureThreshold*100, jobID)
				}
				continue
			}

			successCount++
			idMap[row.ID] = echo.ID
			if successCount%echoLogSampleStep == 0 {
				logUtil.GetLogger().Info("migration echo progress",
					zap.String("module", "migration"),
					zap.String("job_id", jobID),
					zap.Int64("processed", successCount+failCount),
					zap.Int64("success_count", successCount),
					zap.Int64("fail_count", failCount),
					zap.Int64("total", total),
				)
			}

			if req.UpdateProgress != nil {
				req.UpdateProgress(spec.MigrateProgress{
					CurrentPhase: "loading",
					Processed:    successCount + failCount,
					Total:        total,
					SuccessCount: successCount,
					FailCount:    failCount,
				})
			}
		}

		if err := ctx.Err(); err != nil {
			return err
		}
		if err := migrateTags(tx, sourceDB, idMap); err != nil {
			return fmt.Errorf("migrate tags: %w", err)
		}

		if err := ctx.Err(); err != nil {
			return err
		}
		imageSummary, err := migrateImages(tx, sourceDB, sourceRoot, options.CreatedBy, idMap)
		if err != nil {
			return fmt.Errorf("migrate images: %w", err)
		}
		if imageSummary.SourceS3Setting != nil {
			report["source_s3_setting"] = s3SettingToMap(*imageSummary.SourceS3Setting)
		}
		if len(imageSummary.FailedItems) > 0 {
			failed = append(failed, imageSummary.FailedItems...)
			failCount += int64(len(imageSummary.FailedItems))
			if exceedsFailureThreshold(successCount+failCount, failCount, options.FailureThreshold) {
				return fmt.Errorf("failure rate exceeded %.2f%%, rollback job %s", options.FailureThreshold*100, jobID)
			}
		}

		report["processed"] = successCount + failCount
		report["success_count"] = successCount
		report["fail_count"] = failCount
		report["failed_items"] = failed
		report["echo_id_map"] = stringifyIDMap(idMap)
		return nil
	})
	if migrateErr != nil {
		return spec.MigrateResult{}, migrateErr
	}

	if req.UpdateProgress != nil {
		req.UpdateProgress(spec.MigrateProgress{
			CurrentPhase: "completed",
			Processed:    toInt64(report["processed"]),
			Total:        total,
			SuccessCount: toInt64(report["success_count"]),
			FailCount:    toInt64(report["fail_count"]),
		})
	}

	if req.UpdateProgress != nil {
		req.UpdateProgress(spec.MigrateProgress{
			CurrentPhase: "reporting",
			Processed:    toInt64(report["processed"]),
			Total:        total,
			SuccessCount: toInt64(report["success_count"]),
			FailCount:    toInt64(report["fail_count"]),
		})
	}
	logUtil.GetLogger().Info("migration ech0_v3 finished",
		zap.String("module", "migration"),
		zap.String("job_id", jobID),
		zap.Int64("processed", toInt64(report["processed"])),
		zap.Int64("success_count", toInt64(report["success_count"])),
		zap.Int64("fail_count", toInt64(report["fail_count"])),
	)

	return spec.MigrateResult{
		Processed:    toInt64(report["processed"]),
		Total:        total,
		SuccessCount: toInt64(report["success_count"]),
		FailCount:    toInt64(report["fail_count"]),
		ErrorSummary: fmt.Sprintf("迁移完成: success=%d fail=%d", toInt64(report["success_count"]), toInt64(report["fail_count"])),
		JobID:        jobID,
		Report:       report,
	}, nil
}

type v3Echo struct {
	ID            int64     `gorm:"column:id"`
	Content       string    `gorm:"column:content"`
	Username      string    `gorm:"column:username"`
	Private       bool      `gorm:"column:private"`
	Extension     string    `gorm:"column:extension"`
	ExtensionType string    `gorm:"column:extension_type"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	FavCount      int       `gorm:"column:fav_count"`
	Layout        string    `gorm:"column:layout"`
}

type v3Tag struct {
	ID         int64     `gorm:"column:id"`
	Name       string    `gorm:"column:name"`
	UsageCount int       `gorm:"column:usage_count"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

type v3EchoTag struct {
	EchoID int64 `gorm:"column:echo_id"`
	TagID  int64 `gorm:"column:tag_id"`
}

type v3Image struct {
	ID         int64  `gorm:"column:id"`
	MessageID  int64  `gorm:"column:message_id"`
	ImageURL   string `gorm:"column:image_url"`
	ImageSrc   string `gorm:"column:image_source"`
	ObjectKey  string `gorm:"column:object_key"`
	Width      int    `gorm:"column:width"`
	Height     int    `gorm:"column:height"`
}

type migrationOptions struct {
	FailureThreshold float64
	CreatedBy        string
}

func parseOptions(payload map[string]any) migrationOptions {
	options := migrationOptions{
		FailureThreshold: 0.02,
		CreatedBy:        "migrator",
	}
	if v, ok := payload["failure_threshold"].(float64); ok && v > 0 && v < 1 {
		options.FailureThreshold = v
	}
	if v, ok := payload["created_by"].(string); ok && strings.TrimSpace(v) != "" {
		options.CreatedBy = strings.TrimSpace(v)
	}
	return options
}

func resolveSourceDBPath(payload map[string]any) (string, string, error) {
	tmpDir, ok := payload["tmp_dir"].(string)
	if !ok || strings.TrimSpace(tmpDir) == "" {
		return "", "", errors.New("source_payload.tmp_dir is required")
	}
	root := filepath.Join("data", filepath.FromSlash(strings.TrimSpace(tmpDir)))
	dbPath := filepath.Join(root, "ech0.db")
	if _, err := os.Stat(dbPath); err != nil {
		return "", "", fmt.Errorf("source db not found: %w", err)
	}
	return dbPath, root, nil
}

func normalizeUsername(v string) string {
	if strings.TrimSpace(v) == "" {
		return "migrator"
	}
	return strings.TrimSpace(v)
}

func normalizeLayout(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case echoModel.LayoutGrid, echoModel.LayoutHorizontal, echoModel.LayoutCarousel, echoModel.LayoutWaterfall:
		return strings.TrimSpace(strings.ToLower(v))
	default:
		return echoModel.LayoutWaterfall
	}
}

func normalizeTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t
}

func persistExtension(tx *gorm.DB, echoID string, extensionType string, raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	payload := map[string]any{}
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		payload["raw"] = trimmed
	}
	if len(payload) == 0 {
		return nil
	}
	extType := strings.TrimSpace(extensionType)
	if extType == "" {
		extType = "LEGACY"
	}
	return tx.Create(&echoModel.EchoExtension{
		EchoID:  echoID,
		Type:    extType,
		Payload: payload,
	}).Error
}

func migrateTags(tx *gorm.DB, sourceDB *gorm.DB, idMap map[int64]string) error {
	var sourceTags []v3Tag
	if err := sourceDB.Table("tags").Find(&sourceTags).Error; err != nil {
		return err
	}
	if len(sourceTags) == 0 {
		return nil
	}

	tagNames := make([]string, 0, len(sourceTags))
	for i := range sourceTags {
		name := strings.TrimSpace(sourceTags[i].Name)
		if name != "" {
			tagNames = append(tagNames, name)
		}
	}
	existingTags, err := loadTagsByNames(tx, tagNames)
	if err != nil {
		return err
	}

	tagIDMap := make(map[int64]string, len(sourceTags))
	newTags := make([]echoModel.Tag, 0)
	for i := range sourceTags {
		tagRow := sourceTags[i]
		name := strings.TrimSpace(tagRow.Name)
		if name == "" {
			continue
		}
		if tagID, ok := existingTags[name]; ok {
			tagIDMap[tagRow.ID] = tagID
			continue
		}
		newTagID := uuidUtil.MustNewV7()
		newTags = append(newTags, echoModel.Tag{
			ID:         newTagID,
			Name:       name,
			UsageCount: tagRow.UsageCount,
			CreatedAt:  normalizeTime(tagRow.CreatedAt),
		})
		existingTags[name] = newTagID
		tagIDMap[tagRow.ID] = newTagID
	}
	if len(newTags) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(newTags, dbBatchSize).Error; err != nil {
			return err
		}
	}

	var sourceEchoTags []v3EchoTag
	if err := sourceDB.Table("echo_tags").Find(&sourceEchoTags).Error; err != nil {
		return err
	}
	relations := make([]echoModel.EchoTag, 0, len(sourceEchoTags))
	for i := range sourceEchoTags {
		row := sourceEchoTags[i]
		targetEchoID, ok := idMap[row.EchoID]
		if !ok {
			continue
		}
		targetTagID, ok := tagIDMap[row.TagID]
		if !ok {
			continue
		}
		relations = append(relations, echoModel.EchoTag{
			EchoID: targetEchoID,
			TagID:  targetTagID,
		})
	}
	if len(relations) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(relations, dbBatchSize).Error; err != nil {
			return err
		}
	}
	return nil
}

type imageMigrationSummary struct {
	FailedItems    []spec.FailedItem
	SourceS3Setting *settingModel.S3Setting
}

func migrateImages(tx *gorm.DB, sourceDB *gorm.DB, sourceRoot string, createdBy string, idMap map[int64]string) (imageMigrationSummary, error) {
	summary := imageMigrationSummary{
		FailedItems: make([]spec.FailedItem, 0),
	}
	var sourceImages []v3Image
	if err := sourceDB.Table("images").Order("id ASC").Find(&sourceImages).Error; err != nil {
		return summary, err
	}
	if len(sourceImages) == 0 {
		return summary, nil
	}
	sourceS3Setting, err := loadSourceS3Setting(sourceDB)
	if err != nil {
		return summary, fmt.Errorf("load source s3 setting: %w", err)
	}
	summary.SourceS3Setting = sourceS3Setting
	mappedS3Setting, s3SettingValid := mapSourceS3SettingToV4(sourceS3Setting)
	objectProvider := strings.ToLower(strings.TrimSpace(mappedS3Setting.Provider))
	objectBucket := strings.TrimSpace(mappedS3Setting.BucketName)

	type imageCandidate struct {
		sourceImageID int64
		sourceMessageID int64
		targetEchoID string
		key          string
		imageURL     string
		imageSrc     string
		storageType  string
		provider     string
		bucket       string
		width        int
		height       int
	}
	candidates := make([]imageCandidate, 0, len(sourceImages))
	uniqueLocalKeys := make(map[string]struct{})
	uniqueExternalKeys := make(map[string]struct{})
	uniqueObjectKeys := make(map[string]struct{})
	for i := range sourceImages {
		row := sourceImages[i]
		targetEchoID, ok := idMap[row.MessageID]
		if !ok {
			continue
		}

		imageSrc := normalizeImageSource(row.ImageSrc, row.ImageURL)
		switch imageSrc {
		case "local":
			key := strings.TrimSpace(row.ObjectKey)
			if key == "" {
				key = normalizeKeyFromURL(row.ImageURL)
			}
			if key == "" {
				continue
			}
			candidates = append(candidates, imageCandidate{
				sourceImageID: row.ID,
				sourceMessageID: row.MessageID,
				targetEchoID: targetEchoID,
				key:          key,
				imageURL:     strings.TrimSpace(row.ImageURL),
				imageSrc:     imageSrc,
				storageType:  "local",
				provider:     "",
				bucket:       "",
				width:        row.Width,
				height:       row.Height,
			})
			uniqueLocalKeys[key] = struct{}{}
		case "s3":
			if !s3SettingValid {
				summary.FailedItems = append(summary.FailedItems, spec.FailedItem{
					SourceID: strconv.FormatInt(row.MessageID, 10),
					Reason:   "invalid source s3 setting",
				})
				continue
			}
			objectKey := strings.TrimSpace(row.ObjectKey)
			migratedKey := normalizeObjectKeyForV4(objectKey, mappedS3Setting.PathPrefix)
			if migratedKey == "" {
				migratedKey = normalizeObjectKeyForV4(normalizeKeyFromURL(row.ImageURL), mappedS3Setting.PathPrefix)
			}
			if migratedKey == "" {
				summary.FailedItems = append(summary.FailedItems, spec.FailedItem{
					SourceID: strconv.FormatInt(row.MessageID, 10),
					Reason:   "invalid s3 object key",
				})
				continue
			}
			finalURL := buildObjectURLFromSetting(mappedS3Setting, migratedKey)
			candidates = append(candidates, imageCandidate{
				sourceImageID: row.ID,
				sourceMessageID: row.MessageID,
				targetEchoID: targetEchoID,
				key:          migratedKey,
				imageURL:     strings.TrimSpace(finalURL),
				imageSrc:     imageSrc,
				storageType:  "object",
				provider:     objectProvider,
				bucket:       objectBucket,
				width:        row.Width,
				height:       row.Height,
			})
			uniqueObjectKeys[migratedKey] = struct{}{}
		default:
			finalURL := strings.TrimSpace(row.ImageURL)
			if !isAbsoluteURL(finalURL) {
				summary.FailedItems = append(summary.FailedItems, spec.FailedItem{
					SourceID: strconv.FormatInt(row.MessageID, 10),
					Reason:   "invalid external image url",
				})
				continue
			}
			externalKey := buildExternalFileKey(finalURL, "image")
			candidates = append(candidates, imageCandidate{
				sourceImageID: row.ID,
				sourceMessageID: row.MessageID,
				targetEchoID: targetEchoID,
				key:          externalKey,
				imageURL:     finalURL,
				imageSrc:     imageSrc,
				storageType:  "external",
				provider:     "external",
				bucket:       "",
				width:        row.Width,
				height:       row.Height,
			})
			uniqueExternalKeys[externalKey] = struct{}{}
		}
	}
	if len(candidates) == 0 {
		return summary, nil
	}

	existingFileIDs := make(map[string]string)
	localKeys := mapKeys(uniqueLocalKeys)
	localMap, err := loadExistingFileIDsByRoute(tx, "local", "", "", localKeys)
	if err != nil {
		return summary, err
	}
	for k, v := range localMap {
		existingFileIDs[fileRouteKey("local", "", "", k)] = v
	}
	externalKeys := mapKeys(uniqueExternalKeys)
	externalMap, err := loadExistingFileIDsByRoute(tx, "external", "external", "", externalKeys)
	if err != nil {
		return summary, err
	}
	for k, v := range externalMap {
		existingFileIDs[fileRouteKey("external", "external", "", k)] = v
	}
	objectKeys := mapKeys(uniqueObjectKeys)
	objectMap, err := loadExistingFileIDsByRoute(tx, "object", objectProvider, objectBucket, objectKeys)
	if err != nil {
		return summary, err
	}
	for k, v := range objectMap {
		existingFileIDs[fileRouteKey("object", objectProvider, objectBucket, k)] = v
	}

	metaCache := make(map[string]fileMeta, len(candidates))
	newFileRecords := make([]fileModel.File, 0)
	for i := range candidates {
		candidate := candidates[i]
		routeKey := fileRouteKey(candidate.storageType, candidate.provider, candidate.bucket, candidate.key)
		if _, ok := existingFileIDs[routeKey]; ok {
			continue
		}
		meta, ok := metaCache[candidate.key]
		if !ok {
			if candidate.storageType == "local" {
				localPath := resolveMigratedLocalPath(candidate.key)
				meta.size, meta.contentType = inferFileMeta(sourceRoot, candidate.key, candidate.imageURL)
				if copyErr := copySourceImageToLocalStorage(sourceRoot, candidate.key, candidate.imageURL, localPath); copyErr != nil {
					summary.FailedItems = append(summary.FailedItems, spec.FailedItem{
						SourceID: strconv.FormatInt(candidate.sourceMessageID, 10),
						Reason:   "copy local image failed: " + copyErr.Error(),
					})
					logUtil.GetLogger().Warn("copy source image to local storage failed",
						zap.String("module", "migration"),
						zap.String("source_key", candidate.key),
						zap.String("target_path", localPath),
						zap.Error(copyErr),
					)
					continue
				}
			} else {
				meta.size = 0
				meta.contentType = inferContentType(candidate.imageURL, candidate.key)
			}
			metaCache[candidate.key] = meta
		}
		fileID := uuidUtil.MustNewV7()
		fileURL := strings.TrimSpace(candidate.imageURL)
		if candidate.storageType == "local" {
			fileURL = buildMigratedFileURL(candidate.key, candidate.imageURL, candidate.imageSrc)
		}
		newFileRecords = append(newFileRecords, fileModel.File{
			ID:          fileID,
			Key:         candidate.key,
			StorageType: candidate.storageType,
			Provider:    candidate.provider,
			Bucket:      candidate.bucket,
			URL:         fileURL,
			Name:        inferFileName(candidate.key, candidate.imageURL),
			ContentType: meta.contentType,
			Size:        meta.size,
			Width:       candidate.width,
			Height:      candidate.height,
			Category:    inferCategory(meta.contentType, candidate.key),
			UserID:      createdBy,
		})
		existingFileIDs[routeKey] = fileID
	}
	if len(newFileRecords) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(newFileRecords, dbBatchSize).Error; err != nil {
			return summary, err
		}
	}

	echoFiles := make([]fileModel.EchoFile, 0, len(candidates))
	for i := range candidates {
		candidate := candidates[i]
		routeKey := fileRouteKey(candidate.storageType, candidate.provider, candidate.bucket, candidate.key)
		fileID, ok := existingFileIDs[routeKey]
		if !ok || fileID == "" {
			continue
		}
		echoFiles = append(echoFiles, fileModel.EchoFile{
			EchoID: candidate.targetEchoID,
			FileID: fileID,
		})
	}
	if len(echoFiles) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(echoFiles, dbBatchSize).Error; err != nil {
			return summary, err
		}
	}
	return summary, nil
}

type fileMeta struct {
	size        int64
	contentType string
}

func inferFileMeta(sourceRoot string, key string, fallbackURL string) (int64, string) {
	contentType := inferContentType(fallbackURL, key)

	sourcePath := filepath.Join(sourceRoot, filepath.FromSlash(key))
	info, err := os.Stat(sourcePath)
	if err != nil {
		return 0, contentType
	}
	return info.Size(), contentType
}

func inferCategory(contentType string, key string) string {
	lowerType := strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case strings.HasPrefix(lowerType, "image/"):
		return "image"
	case strings.HasPrefix(lowerType, "audio/"):
		return "audio"
	case strings.HasPrefix(lowerType, "video/"):
		return "video"
	}
	ext := strings.ToLower(filepath.Ext(key))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".svg":
		return "image"
	case ".mp3", ".wav", ".ogg", ".m4a":
		return "audio"
	case ".mp4", ".mov", ".webm":
		return "video"
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx":
		return "document"
	default:
		return "file"
	}
}

func normalizeKeyFromURL(url string) string {
	trimmed := strings.TrimSpace(url)
	if trimmed == "" {
		return ""
	}
	base := filepath.Base(trimmed)
	base = strings.Split(base, "?")[0]
	if base == "." || base == "/" {
		return ""
	}
	return base
}

func buildMigratedFileURL(key string, imageURL string, _ string) string {
	cleanKey := cleanMigratedFileKey(key)
	if cleanKey == "" {
		cleanKey = cleanMigratedFileKey(normalizeKeyFromURL(imageURL))
	}
	if cleanKey == "" {
		return strings.TrimSpace(imageURL)
	}
	resolvedPath := resolveMigratedLocalPath(cleanKey)
	if resolvedPath == "" {
		return strings.TrimSpace(imageURL)
	}
	return "/api/files/" + strings.TrimLeft(resolvedPath, "/")
}

func normalizeImageSource(source string, imageURL string) string {
	s := strings.ToLower(strings.TrimSpace(source))
	switch {
	case strings.Contains(s, "s3"), strings.Contains(s, "r2"), strings.Contains(s, "object"):
		return "s3"
	case strings.Contains(s, "url"), strings.Contains(s, "http"), strings.Contains(s, "external"), strings.Contains(s, "link"):
		return "url"
	case strings.Contains(s, "local"):
		return "local"
	}
	if isAbsoluteURL(strings.TrimSpace(imageURL)) {
		return "url"
	}
	return "local"
}

func loadSourceS3Setting(sourceDB *gorm.DB) (*settingModel.S3Setting, error) {
	type v3KeyValue struct {
		Key   string `gorm:"column:key"`
		Value string `gorm:"column:value"`
	}
	var kv v3KeyValue
	if err := sourceDB.Table("key_values").Where("key = ?", "s3_setting").Take(&kv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	raw := strings.TrimSpace(kv.Value)
	if raw == "" {
		return nil, nil
	}
	var setting settingModel.S3Setting
	if err := json.Unmarshal([]byte(raw), &setting); err != nil {
		return nil, err
	}
	return &setting, nil
}

func buildExternalFileKey(rawURL string, category string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(rawURL)))
	return "external/" + strings.TrimSpace(category) + "/" + fmt.Sprintf("%x", sum[:])
}

func mapSourceS3SettingToV4(setting *settingModel.S3Setting) (settingModel.S3Setting, bool) {
	if setting == nil {
		return settingModel.S3Setting{}, false
	}
	mapped := settingModel.S3Setting{
		Enable:     true,
		Provider:   strings.ToLower(strings.TrimSpace(setting.Provider)),
		Endpoint:   strings.TrimSpace(setting.Endpoint),
		AccessKey:  strings.TrimSpace(setting.AccessKey),
		SecretKey:  strings.TrimSpace(setting.SecretKey),
		BucketName: strings.TrimSpace(setting.BucketName),
		Region:     strings.TrimSpace(setting.Region),
		UseSSL:     setting.UseSSL,
		CDNURL:     strings.TrimRight(strings.TrimSpace(setting.CDNURL), "/"),
		PathPrefix: strings.Trim(strings.TrimSpace(setting.PathPrefix), "/"),
		PublicRead: setting.PublicRead,
	}
	mapped.Endpoint = strings.TrimPrefix(mapped.Endpoint, "http://")
	mapped.Endpoint = strings.TrimPrefix(mapped.Endpoint, "https://")
	if mapped.Provider == "" || mapped.Endpoint == "" || mapped.BucketName == "" || mapped.AccessKey == "" || mapped.SecretKey == "" {
		return mapped, false
	}
	if mapped.Region == "" {
		mapped.Region = "auto"
	}
	return mapped, true
}

func normalizeObjectKeyForV4(objectKey string, pathPrefix string) string {
	clean := cleanMigratedFileKey(objectKey)
	if clean == "" {
		return ""
	}
	prefix := strings.Trim(strings.TrimSpace(pathPrefix), "/")
	if prefix != "" {
		prefixWithSlash := prefix + "/"
		clean = strings.TrimPrefix(clean, prefixWithSlash)
	}
	for _, route := range []string{"images/", "audios/", "videos/", "documents/", "files/"} {
		if strings.HasPrefix(clean, route) {
			clean = strings.TrimPrefix(clean, route)
			break
		}
	}
	return cleanMigratedFileKey(clean)
}

func buildObjectURLFromSetting(setting settingModel.S3Setting, key string) string {
	resolvedKey := strings.Trim(strings.TrimSpace(storage.NewFileSchema().Resolve(cleanMigratedFileKey(key))), "/")
	if resolvedKey == "" {
		return ""
	}
	if prefix := strings.Trim(strings.TrimSpace(setting.PathPrefix), "/"); prefix != "" {
		resolvedKey = prefix + "/" + resolvedKey
	}
	cdn := strings.TrimRight(strings.TrimSpace(setting.CDNURL), "/")
	if cdn != "" {
		return cdn + "/" + resolvedKey
	}
	endpoint := strings.TrimSpace(setting.Endpoint)
	if endpoint == "" {
		return ""
	}
	if !strings.HasPrefix(strings.ToLower(endpoint), "http://") && !strings.HasPrefix(strings.ToLower(endpoint), "https://") {
		scheme := "https"
		if !setting.UseSSL {
			scheme = "http"
		}
		endpoint = scheme + "://" + endpoint
	}
	endpoint = strings.TrimRight(endpoint, "/")
	bucket := strings.Trim(strings.TrimSpace(setting.BucketName), "/")
	if bucket == "" {
		return ""
	}
	return endpoint + "/" + bucket + "/" + resolvedKey
}

func s3SettingToMap(setting settingModel.S3Setting) map[string]any {
	return map[string]any{
		"enable":      setting.Enable,
		"provider":    setting.Provider,
		"endpoint":    setting.Endpoint,
		"access_key":  setting.AccessKey,
		"secret_key":  setting.SecretKey,
		"bucket_name": setting.BucketName,
		"region":      setting.Region,
		"use_ssl":     setting.UseSSL,
		"cdn_url":     setting.CDNURL,
		"path_prefix": setting.PathPrefix,
		"public_read": setting.PublicRead,
	}
}

func cleanMigratedFileKey(key string) string {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return ""
	}
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	trimmed = strings.Split(trimmed, "?")[0]
	return strings.Trim(trimmed, "/")
}

func resolveMigratedLocalPath(key string) string {
	clean := cleanMigratedFileKey(key)
	if clean == "" {
		return ""
	}
	for _, prefix := range []string{"images/", "audios/", "videos/", "documents/", "files/"} {
		if strings.HasPrefix(clean, prefix) {
			return clean
		}
	}
	return storage.NewFileSchema().Resolve(clean)
}

func inferFileName(key string, fileURL string) string {
	if base := path.Base(strings.Trim(strings.TrimSpace(key), "/")); base != "" && base != "." && base != "/" {
		return base
	}
	if base := path.Base(strings.Split(strings.TrimSpace(fileURL), "?")[0]); base != "" && base != "." && base != "/" {
		return base
	}
	return "file"
}

func inferContentType(primary string, secondary string) string {
	if ext := strings.ToLower(filepath.Ext(strings.TrimSpace(primary))); ext != "" {
		if ct := mime.TypeByExtension(ext); ct != "" {
			return ct
		}
	}
	if ext := strings.ToLower(filepath.Ext(strings.TrimSpace(secondary))); ext != "" {
		if ct := mime.TypeByExtension(ext); ct != "" {
			return ct
		}
	}
	return "application/octet-stream"
}

func isAbsoluteURL(raw string) bool {
	lower := strings.ToLower(strings.TrimSpace(raw))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func resolveSourceImagePathCandidates(sourceRoot string, key string, imageURL string, targetPath string) []string {
	candidates := make([]string, 0, 8)
	appendIf := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" {
			return
		}
		candidates = append(candidates, filepath.Clean(p))
	}
	cleanKey := cleanMigratedFileKey(key)
	if cleanKey != "" {
		appendIf(filepath.Join(sourceRoot, filepath.FromSlash(cleanKey)))
		appendIf(filepath.Join(sourceRoot, filepath.FromSlash(path.Base(cleanKey))))
	}
	rawURL := strings.TrimSpace(imageURL)
	rawURL = strings.Split(rawURL, "?")[0]
	trimmedURLPath := strings.Trim(strings.TrimPrefix(rawURL, "/"), "/")
	if trimmedURLPath != "" {
		appendIf(filepath.Join(sourceRoot, filepath.FromSlash(trimmedURLPath)))
		appendIf(filepath.Join(sourceRoot, filepath.FromSlash(path.Base(trimmedURLPath))))
	}
	targetPath = strings.Trim(strings.TrimSpace(targetPath), "/")
	if targetPath != "" {
		appendIf(filepath.Join(sourceRoot, filepath.FromSlash(targetPath)))
		appendIf(filepath.Join(sourceRoot, filepath.FromSlash(path.Base(targetPath))))
	}
	return dedupePaths(candidates)
}

func copySourceImageToLocalStorage(sourceRoot string, key string, imageURL string, targetPath string) error {
	targetPath = strings.Trim(strings.TrimSpace(targetPath), "/")
	if targetPath == "" {
		return errors.New("empty target path")
	}
	targetRoot := strings.TrimSpace(config.Config().Storage.DataRoot)
	if targetRoot == "" {
		targetRoot = "data/files"
	}
	targetFullPath := filepath.Join(targetRoot, filepath.FromSlash(targetPath))
	if info, err := os.Stat(targetFullPath); err == nil && !info.IsDir() {
		return nil
	}
	candidates := resolveSourceImagePathCandidates(sourceRoot, key, imageURL, targetPath)
	var lastErr error
	for _, src := range candidates {
		info, err := os.Stat(src)
		if err != nil || info.IsDir() {
			lastErr = err
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetFullPath), 0o755); err != nil {
			return err
		}
		if err := copyFileContents(src, targetFullPath); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	if lastErr != nil {
		return lastErr
	}
	return errors.New("source file not found")
}

func dedupePaths(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	out := make([]string, 0, len(input))
	for _, p := range input {
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func copyFileContents(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func loadTagsByNames(tx *gorm.DB, names []string) (map[string]string, error) {
	result := make(map[string]string)
	if len(names) == 0 {
		return result, nil
	}
	var tags []echoModel.Tag
	if err := tx.Where("name IN ?", names).Find(&tags).Error; err != nil {
		return nil, err
	}
	for i := range tags {
		result[tags[i].Name] = tags[i].ID
	}
	return result, nil
}

func fileRouteKey(storageType string, provider string, bucket string, key string) string {
	return strings.TrimSpace(storageType) + "|" + strings.TrimSpace(provider) + "|" + strings.TrimSpace(bucket) + "|" + strings.TrimSpace(key)
}

func loadExistingFileIDsByRoute(tx *gorm.DB, storageType string, provider string, bucket string, keys []string) (map[string]string, error) {
	result := make(map[string]string)
	if len(keys) == 0 {
		return result, nil
	}
	for start := 0; start < len(keys); start += dbBatchSize {
		end := start + dbBatchSize
		if end > len(keys) {
			end = len(keys)
		}
		var files []fileModel.File
		if err := tx.Where(
			"storage_type = ? AND provider = ? AND bucket = ? AND key IN ?",
			strings.TrimSpace(storageType),
			strings.TrimSpace(provider),
			strings.TrimSpace(bucket),
			keys[start:end],
		).Find(&files).Error; err != nil {
			return nil, err
		}
		for i := range files {
			result[files[i].Key] = files[i].ID
		}
	}
	return result, nil
}

func mapKeys(set map[string]struct{}) []string {
	if len(set) == 0 {
		return nil
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	return keys
}

func stringifyIDMap(idMap map[int64]string) map[string]string {
	out := make(map[string]string, len(idMap))
	for oldID, newID := range idMap {
		out[strconv.FormatInt(oldID, 10)] = newID
	}
	return out
}

func loadEchoIDsWithImages(sourceDB *gorm.DB) (map[int64]struct{}, error) {
	var images []v3Image
	if err := sourceDB.Table("images").Select("message_id").Find(&images).Error; err != nil {
		return nil, err
	}
	out := make(map[int64]struct{}, len(images))
	for i := range images {
		out[images[i].MessageID] = struct{}{}
	}
	return out, nil
}

func exceedsFailureThreshold(processed int64, failed int64, threshold float64) bool {
	if processed == 0 {
		return false
	}
	return float64(failed)/float64(processed) > threshold
}

func toInt64(v any) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case int:
		return int64(x)
	case float64:
		return int64(x)
	default:
		return 0
	}
}
