package registry

import (
	"context"

	busen "github.com/lin-snow/Busen"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
)

func (er *EventRegistrar) registerDeadLetter() error {
	unsub, err := busen.Subscribe(er.bus,
		func(ctx context.Context, e busen.Event[contracts.DeadLetterRetriedEvent]) error {
			return er.eh.dlr.Handle(ctx, e.Value)
		},
		er.deadLetterOptions()...,
	)
	if err != nil {
		return err
	}
	er.unsub = append(er.unsub, unsub)
	return nil
}

func (er *EventRegistrar) registerSystem() error {
	unsub, err := busen.Subscribe(er.bus,
		func(ctx context.Context, e busen.Event[contracts.UpdateBackupScheduleEvent]) error {
			return er.eh.bs.HandleBackupScheduleUpdated(ctx, e.Value)
		},
		er.systemOptions()...,
	)
	if err != nil {
		return err
	}
	er.unsub = append(er.unsub, unsub)
	return nil
}

func (er *EventRegistrar) registerAgent() error {
	unsub, err := busen.Subscribe(er.bus,
		func(ctx context.Context, e busen.Event[contracts.EchoCreatedEvent]) error {
			return er.eh.ap.HandleEchoCreated(ctx, e.Value)
		},
		er.agentOptions()...,
	)
	if err != nil {
		return err
	}
	er.unsub = append(er.unsub, unsub)

	unsub, err = busen.Subscribe(er.bus,
		func(ctx context.Context, e busen.Event[contracts.EchoUpdatedEvent]) error {
			return er.eh.ap.HandleEchoUpdated(ctx, e.Value)
		},
		er.agentOptions()...,
	)
	if err != nil {
		return err
	}
	er.unsub = append(er.unsub, unsub)

	unsub, err = busen.Subscribe(er.bus,
		func(ctx context.Context, e busen.Event[contracts.UserDeletedEvent]) error {
			return er.eh.ap.HandleUserDeleted(ctx, e.Value)
		},
		er.agentOptions()...,
	)
	if err != nil {
		return err
	}
	er.unsub = append(er.unsub, unsub)
	return nil
}

func (er *EventRegistrar) registerInbox() error {
	unsub, err := busen.Subscribe(er.bus,
		func(ctx context.Context, e busen.Event[contracts.Ech0UpdateCheckEvent]) error {
			return er.eh.id.HandleEch0UpdateCheck(ctx, e.Value)
		},
		er.inboxOptions()...,
	)
	if err != nil {
		return err
	}
	er.unsub = append(er.unsub, unsub)

	unsub, err = busen.Subscribe(er.bus,
		func(ctx context.Context, e busen.Event[contracts.InboxClearEvent]) error {
			return er.eh.id.HandleInboxClear(ctx, e.Value)
		},
		er.inboxOptions()...,
	)
	if err != nil {
		return err
	}
	er.unsub = append(er.unsub, unsub)
	return nil
}
