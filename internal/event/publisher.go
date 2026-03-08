package event

import (
	"context"
	"strconv"

	busen "github.com/lin-snow/Busen"
)

type Publisher struct {
	bus *busen.Bus
}

func NewPublisher(busProvider func() *busen.Bus) *Publisher {
	return &Publisher{bus: busProvider()}
}

func (p *Publisher) UserCreated(ctx context.Context, evt UserCreatedEvent) error {
	return busen.Publish(ctx, p.bus, evt,
		busen.WithTopic(TopicUserCreated),
		busen.WithKey(strconv.FormatUint(uint64(evt.User.ID), 10)))
}

func (p *Publisher) UserUpdated(ctx context.Context, evt UserUpdatedEvent) error {
	return busen.Publish(ctx, p.bus, evt,
		busen.WithTopic(TopicUserUpdated),
		busen.WithKey(strconv.FormatUint(uint64(evt.User.ID), 10)))
}

func (p *Publisher) UserDeleted(ctx context.Context, evt UserDeletedEvent) error {
	return busen.Publish(ctx, p.bus, evt,
		busen.WithTopic(TopicUserDeleted),
		busen.WithKey(strconv.FormatUint(uint64(evt.User.ID), 10)))
}

func (p *Publisher) EchoCreated(ctx context.Context, evt EchoCreatedEvent) error {
	return busen.Publish(ctx, p.bus, evt,
		busen.WithTopic(TopicEchoCreated),
		busen.WithKey(strconv.FormatUint(uint64(evt.Echo.ID), 10)))
}

func (p *Publisher) EchoUpdated(ctx context.Context, evt EchoUpdatedEvent) error {
	return busen.Publish(ctx, p.bus, evt,
		busen.WithTopic(TopicEchoUpdated),
		busen.WithKey(strconv.FormatUint(uint64(evt.Echo.ID), 10)))
}

func (p *Publisher) EchoDeleted(ctx context.Context, evt EchoDeletedEvent) error {
	return busen.Publish(ctx, p.bus, evt,
		busen.WithTopic(TopicEchoDeleted),
		busen.WithKey(strconv.FormatUint(uint64(evt.Echo.ID), 10)))
}

func (p *Publisher) ResourceUploaded(ctx context.Context, evt ResourceUploadedEvent, key string) error {
	return busen.Publish(ctx, p.bus, evt,
		busen.WithTopic(TopicResourceUploaded),
		busen.WithKey(key))
}

func (p *Publisher) SystemBackup(ctx context.Context, evt SystemBackupEvent) error {
	return busen.Publish(ctx, p.bus, evt, busen.WithTopic(TopicSystemBackup))
}

func (p *Publisher) SystemRestore(ctx context.Context, evt SystemRestoreEvent) error {
	return busen.Publish(ctx, p.bus, evt, busen.WithTopic(TopicSystemRestore))
}

func (p *Publisher) SystemExport(ctx context.Context, evt SystemExportEvent) error {
	return busen.Publish(ctx, p.bus, evt, busen.WithTopic(TopicSystemExport))
}

func (p *Publisher) BackupScheduleUpdated(ctx context.Context, evt UpdateBackupScheduleEvent) error {
	return busen.Publish(ctx, p.bus, evt, busen.WithTopic(TopicBackupScheduleUpdate))
}

func (p *Publisher) DeadLetterRetried(ctx context.Context, evt DeadLetterRetriedEvent) error {
	return busen.Publish(ctx, p.bus, evt,
		busen.WithTopic(TopicDeadLetterRetried),
		busen.WithKey(strconv.FormatUint(uint64(evt.DeadLetter.ID), 10)))
}

func (p *Publisher) Ech0UpdateChecked(ctx context.Context, evt Ech0UpdateCheckEvent) error {
	return busen.Publish(ctx, p.bus, evt, busen.WithTopic(TopicEch0UpdateCheck))
}

func (p *Publisher) InboxCleared(ctx context.Context, evt InboxClearEvent) error {
	return busen.Publish(ctx, p.bus, evt, busen.WithTopic(TopicInboxClear))
}
