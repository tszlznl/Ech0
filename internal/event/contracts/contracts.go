package contracts

import (
	"encoding/json"
	"reflect"
	"time"

	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	queueModel "github.com/lin-snow/ech0/internal/model/queue"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
)

const (
	TopicUserCreated          = "user.created"
	TopicUserUpdated          = "user.updated"
	TopicUserDeleted          = "user.deleted"
	TopicEchoCreated          = "echo.created"
	TopicEchoUpdated          = "echo.updated"
	TopicEchoDeleted          = "echo.deleted"
	TopicCommentCreated       = "comment.created"
	TopicCommentStatusUpdated = "comment.status.updated"
	TopicCommentDeleted       = "comment.deleted"
	TopicResourceUploaded     = "resource.uploaded"
	TopicSystemBackup         = "system.backup"
	TopicSystemExport         = "system.export"
	TopicBackupScheduleUpdate = "system.backup_schedule.updated"
	TopicDeadLetterRetried    = "deadletter.retried"
	TopicEch0UpdateCheck      = "ech0.update.check"
)

var webhookTopicWhitelist = map[string]struct{}{
	TopicUserCreated:          {},
	TopicUserUpdated:          {},
	TopicUserDeleted:          {},
	TopicEchoCreated:          {},
	TopicEchoUpdated:          {},
	TopicEchoDeleted:          {},
	TopicCommentCreated:       {},
	TopicCommentStatusUpdated: {},
	TopicCommentDeleted:       {},
	TopicResourceUploaded:     {},
	TopicSystemBackup:         {},
	TopicSystemExport:         {},
	TopicBackupScheduleUpdate: {},
	TopicEch0UpdateCheck:      {},
}

type (
	UserCreatedEvent struct{ User userModel.User }
	UserUpdatedEvent struct{ User userModel.User }
	UserDeletedEvent struct{ User userModel.User }

	EchoCreatedEvent struct {
		Echo echoModel.Echo
		User userModel.User
	}
	EchoUpdatedEvent struct {
		Echo echoModel.Echo
		User userModel.User
	}
	EchoDeletedEvent struct {
		Echo echoModel.Echo
		User userModel.User
	}
	CommentCreatedEvent struct {
		Comment commentModel.Comment
	}
	CommentStatusUpdatedEvent struct {
		Comment commentModel.Comment
	}
	CommentDeletedEvent struct {
		Comment commentModel.Comment
	}

	ResourceUploadedEvent struct {
		User     userModel.User
		FileName string
		URL      string
		Size     int64
		Type     string
	}

	SystemBackupEvent struct {
		Info string
		Size int64
	}
	SystemExportEvent struct {
		Info string
		Size int64
	}

	UpdateBackupScheduleEvent struct {
		Schedule settingModel.BackupSchedule
	}

	DeadLetterRetriedEvent struct {
		DeadLetter queueModel.DeadLetter
	}
	Ech0UpdateCheckEvent struct{ Info string }

	WebhookObservation struct {
		Topic      string            `json:"topic"`
		EventName  string            `json:"event_name"`
		Payload    json.RawMessage   `json:"payload"`
		Metadata   map[string]string `json:"metadata,omitempty"`
		OccurredAt int64             `json:"occurred_at"`
	}

	WebhookReplayPayload struct {
		Webhook webhookModel.Webhook `json:"webhook"`
		Event   WebhookObservation   `json:"event"`
	}
)

func NewWebhookObservation(topic string, payload any, metadata map[string]string) (WebhookObservation, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return WebhookObservation{}, err
	}
	return WebhookObservation{
		Topic:      topic,
		EventName:  eventNameOf(payload),
		Payload:    raw,
		Metadata:   metadata,
		OccurredAt: time.Now().UTC().Unix(),
	}, nil
}

func eventNameOf(payload any) string {
	if payload == nil {
		return ""
	}
	t := reflect.TypeOf(payload)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Name() != "" {
		return t.Name()
	}
	return t.String()
}

func IsWebhookTopicAllowed(topic string) bool {
	_, ok := webhookTopicWhitelist[topic]
	return ok
}
