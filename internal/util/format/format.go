package util

import (
	"fmt"
	"strings"

	"github.com/robfig/cron/v3"
)

func ValidateCrontabExpression(expr string) error {
	// 去掉多余空格，防止 "  * * * * *  " 之类情况
	expr = strings.TrimSpace(expr)
	fields := strings.Fields(expr)

	switch len(fields) {
	case 5:
		// 标准 cron（分钟、小时、日、月、星期）
		_, err := cron.ParseStandard(expr)
		if err != nil {
			return fmt.Errorf("invalid 5-field cron expression: %w", err)
		}
	case 6:
		// 含秒字段（秒、分钟、小时、日、月、星期）
		parser := cron.NewParser(
			cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
		)
		_, err := parser.Parse(expr)
		if err != nil {
			return fmt.Errorf("invalid 6-field cron expression: %w", err)
		}
	default:
		return fmt.Errorf("cron expression must have 5 or 6 fields, got %d", len(fields))
	}

	return nil
}
