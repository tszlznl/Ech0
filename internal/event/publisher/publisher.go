package publisher

import (
	"context"
	"fmt"

	busen "github.com/lin-snow/Busen"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
)

type Publisher struct {
	bus *busen.Bus
}

func New(busProvider func() *busen.Bus) *Publisher {
	return &Publisher{bus: busProvider()}
}

func (p *Publisher) UserCreated(ctx context.Context, evt contracts.UserCreatedEvent) error {
	return busen.Publish(ctx, p.bus, evt,
		busen.WithTopic(contracts.TopicUserCreated),
		busen.WithKey(evt.User.ID))
}

func (p *Publisher) UserUpdated(ctx context.Context, evt contracts.UserUpdatedEvent) error {
	return busen.Publish(ctx, p.bus, evt,
		busen.WithTopic(contracts.TopicUserUpdated),
		busen.WithKey(evt.User.ID))
}

func (p *Publisher) UserDeleted(ctx context.Context, evt contracts.UserDeletedEvent) error {
	return busen.Publish(ctx, p.bus, evt,
		busen.WithTopic(contracts.TopicUserDeleted),
		busen.WithKey(evt.User.ID))
}

func (p *Publisher) EchoCreated(ctx context.Context, evt contracts.EchoCreatedEvent) error {
	return busen.Publish(ctx, p.bus, evt,
		busen.WithTopic(contracts.TopicEchoCreated),
		busen.WithKey(evt.Echo.ID))
}

func (p *Publisher) EchoUpdated(ctx context.Context, evt contracts.EchoUpdatedEvent) error {
	return busen.Publish(ctx, p.bus, evt,
		busen.WithTopic(contracts.TopicEchoUpdated),
		busen.WithKey(evt.Echo.ID))
}

func (p *Publisher) EchoDeleted(ctx context.Context, evt contracts.EchoDeletedEvent) error {
	return busen.Publish(ctx, p.bus, evt,
		busen.WithTopic(contracts.TopicEchoDeleted),
		busen.WithKey(evt.Echo.ID))
}

func (p *Publisher) ResourceUploaded(ctx context.Context, evt contracts.ResourceUploadedEvent, key string) error {
	return busen.Publish(ctx, p.bus, evt,
		busen.WithTopic(contracts.TopicResourceUploaded),
		busen.WithKey(key))
}

func (p *Publisher) SystemBackup(ctx context.Context, evt contracts.SystemBackupEvent) error {
	return busen.Publish(ctx, p.bus, evt, busen.WithTopic(contracts.TopicSystemBackup))
}

func (p *Publisher) SystemRestore(ctx context.Context, evt contracts.SystemRestoreEvent) error {
	return busen.Publish(ctx, p.bus, evt, busen.WithTopic(contracts.TopicSystemRestore))
}

func (p *Publisher) SystemExport(ctx context.Context, evt contracts.SystemExportEvent) error {
	return busen.Publish(ctx, p.bus, evt, busen.WithTopic(contracts.TopicSystemExport))
}

func (p *Publisher) BackupScheduleUpdated(ctx context.Context, evt contracts.UpdateBackupScheduleEvent) error {
	return busen.Publish(ctx, p.bus, evt, busen.WithTopic(contracts.TopicBackupScheduleUpdate))
}

func (p *Publisher) DeadLetterRetried(ctx context.Context, evt contracts.DeadLetterRetriedEvent) error {
	return busen.Publish(ctx, p.bus, evt,
		busen.WithTopic(contracts.TopicDeadLetterRetried),
		busen.WithKey(fmt.Sprint(evt.DeadLetter.ID)))
}

func (p *Publisher) Ech0UpdateChecked(ctx context.Context, evt contracts.Ech0UpdateCheckEvent) error {
	return busen.Publish(ctx, p.bus, evt, busen.WithTopic(contracts.TopicEch0UpdateCheck))
}

func (p *Publisher) InboxCleared(ctx context.Context, evt contracts.InboxClearEvent) error {
	return busen.Publish(ctx, p.bus, evt, busen.WithTopic(contracts.TopicInboxClear))
}
