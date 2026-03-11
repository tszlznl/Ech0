package load

import (
	"context"
	"strings"

	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/migrator/spec"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
)

// EchoLoader 将迁移结果以 Echo 形式写入数据库。
type EchoLoader struct {
	createdBy string
}

func NewEchoLoader(createdBy string) *EchoLoader {
	return &EchoLoader{createdBy: createdBy}
}

func (l *EchoLoader) Load(_ context.Context, records []spec.CanonicalRecord) (spec.LoadResult, error) {
	failed := make([]spec.FailedItem, 0)
	var loaded int64
	for _, record := range records {
		content := strings.TrimSpace(record.Content)
		if content == "" {
			content = strings.TrimSpace(record.Title)
		}
		if content == "" {
			failed = append(failed, spec.FailedItem{
				SourceID: record.SourceID,
				Reason:   "empty content after transform",
			})
			continue
		}

		echo := &echoModel.Echo{
			Content:  content,
			UserID:   l.createdBy,
			Username: "migrator",
			Layout:   echoModel.LayoutWaterfall,
			Private:  false,
		}

		if err := database.GetDB().Create(echo).Error; err != nil {
			failed = append(failed, spec.FailedItem{
				SourceID: record.SourceID,
				Reason:   err.Error(),
			})
			continue
		}
		loaded += 1
	}

	return spec.LoadResult{
		Loaded: loaded,
		Failed: failed,
	}, nil
}
