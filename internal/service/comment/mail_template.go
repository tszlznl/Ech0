package service

import (
	"fmt"
	stdhtml "html"
	"strings"
	"time"

	model "github.com/lin-snow/ech0/internal/model/comment"
)

type notifyContent struct {
	Subject  string
	TextBody string
	HTMLBody string
}

func buildNotifyContent(kind string, comment model.Comment) notifyContent {
	subject := notifySubject(kind, comment.Status)
	eventTitle := notifyEventTitle(kind, comment.Status)
	statusLabel, statusColor, statusBg := notifyStatusStyle(kind, comment.Status)
	createdAt := notifyTime(comment.CreatedAt)
	contentText := strings.TrimSpace(comment.Content)
	if contentText == "" && kind == "test" {
		contentText = "这是一封来自 Ech0 的评论通知测试邮件。"
	}
	contentHTML := strings.ReplaceAll(stdhtml.EscapeString(contentText), "\n", "<br/>")
	text := fmt.Sprintf(
		"Ech0 评论通知\n\n事件: %s\n评论ID: %s\n昵称: %s\n状态: %s\n来源: %s\n时间: %s\n\n内容:\n%s",
		eventTitle,
		strings.TrimSpace(comment.ID),
		strings.TrimSpace(comment.Nickname),
		statusLabel,
		strings.TrimSpace(string(comment.Source)),
		createdAt,
		contentText,
	)
	htmlBody := fmt.Sprintf(
		`<!doctype html><html><body style="margin:0;padding:0;background:#f3f6fb;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Arial,sans-serif;color:#1f2937;">
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="padding:24px 12px;">
  <tr><td align="center">
    <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="max-width:640px;background:#ffffff;border:1px solid #e5e7eb;border-radius:14px;overflow:hidden;">
      <tr><td style="padding:18px 20px;background:#0f172a;color:#f8fafc;font-size:16px;font-weight:700;">Ech0 评论通知</td></tr>
      <tr><td style="padding:18px 20px;">
        <div style="font-size:18px;font-weight:700;margin-bottom:10px;color:#111827;">%s</div>
        <div style="display:inline-block;padding:3px 10px;border-radius:999px;font-size:12px;font-weight:600;color:%s;background:%s;border:1px solid %s;">%s</div>
        <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="margin-top:14px;border-collapse:collapse;background:#f8fafc;border:1px solid #e5e7eb;border-radius:10px;">
          <tr><td style="padding:10px 12px;font-size:13px;color:#475569;width:100px;">评论ID</td><td style="padding:10px 12px;font-size:13px;color:#111827;">%s</td></tr>
          <tr><td style="padding:10px 12px;font-size:13px;color:#475569;">昵称</td><td style="padding:10px 12px;font-size:13px;color:#111827;">%s</td></tr>
          <tr><td style="padding:10px 12px;font-size:13px;color:#475569;">来源</td><td style="padding:10px 12px;font-size:13px;color:#111827;">%s</td></tr>
          <tr><td style="padding:10px 12px;font-size:13px;color:#475569;">时间</td><td style="padding:10px 12px;font-size:13px;color:#111827;">%s</td></tr>
        </table>
        <div style="margin-top:14px;font-size:13px;color:#475569;margin-bottom:6px;">评论内容</div>
        <div style="padding:12px;border:1px solid #e5e7eb;border-radius:10px;background:#ffffff;line-height:1.65;font-size:14px;color:#111827;word-break:break-word;">%s</div>
      </td></tr>
    </table>
  </td></tr>
</table>
</body></html>`,
		stdhtml.EscapeString(eventTitle),
		statusColor,
		statusBg,
		statusBg,
		stdhtml.EscapeString(statusLabel),
		stdhtml.EscapeString(strings.TrimSpace(comment.ID)),
		stdhtml.EscapeString(strings.TrimSpace(comment.Nickname)),
		stdhtml.EscapeString(strings.TrimSpace(string(comment.Source))),
		stdhtml.EscapeString(createdAt),
		contentHTML,
	)
	return notifyContent{
		Subject:  subject,
		TextBody: text,
		HTMLBody: htmlBody,
	}
}

func notifySubject(kind string, status model.Status) string {
	prefix := "[Ech0评论通知]"
	switch kind {
	case "created":
		return prefix + " 新评论待处理"
	case "status":
		if status == model.StatusApproved {
			return prefix + " 评论审核通过"
		}
		if status == model.StatusRejected {
			return prefix + " 评论审核拒绝"
		}
		return prefix + " 评论状态变更"
	case "hot":
		return prefix + " 评论被设为Hot"
	default:
		return prefix + " 测试邮件"
	}
}

func notifyEventTitle(kind string, status model.Status) string {
	switch kind {
	case "created":
		return "有新评论待处理"
	case "status":
		if status == model.StatusApproved {
			return "评论已审核通过"
		}
		if status == model.StatusRejected {
			return "评论已审核拒绝"
		}
		return "评论状态已更新"
	case "hot":
		return "评论已被设为 Hot"
	default:
		return "评论通知测试邮件"
	}
}

func notifyStatusStyle(kind string, status model.Status) (label, color, bg string) {
	switch kind {
	case "hot":
		return "HOT", "#7c3aed", "#ede9fe"
	case "status":
		if status == model.StatusApproved {
			return "已通过", "#047857", "#d1fae5"
		}
		if status == model.StatusRejected {
			return "已拒绝", "#b45309", "#fef3c7"
		}
	}
	if status == model.StatusPending {
		return "待审核", "#0369a1", "#e0f2fe"
	}
	if status == model.StatusApproved {
		return "已通过", "#047857", "#d1fae5"
	}
	if status == model.StatusRejected {
		return "已拒绝", "#b45309", "#fef3c7"
	}
	return "通知", "#334155", "#e2e8f0"
}

func notifyTime(ts time.Time) string {
	if ts.IsZero() {
		ts = time.Now()
	}
	return ts.Local().Format("2006-01-02 15:04:05")
}
