package event

import (
	"encoding/json"
	"testing"

	bus "github.com/lin-snow/ech0/internal/event/bus"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
)

func TestWebhookReplayPayload_RoundTripKeepsPayloadRaw(t *testing.T) {
	raw := json.RawMessage(`{"id":1,"name":"echo"}`)
	in := contracts.WebhookReplayPayload{
		Webhook: webhookModel.Webhook{
			Name: "hook",
			URL:  "https://example.com/hook",
		},
		Event: contracts.WebhookObservation{
			Topic:      contracts.TopicEchoCreated,
			EventName:  "EchoCreatedEvent",
			Payload:    raw,
			Metadata:   map[string]string{bus.MetaKeySource: "test"},
			OccurredAt: 123,
		},
	}

	buf, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal payload failed: %v", err)
	}

	var out contracts.WebhookReplayPayload
	if err := json.Unmarshal(buf, &out); err != nil {
		t.Fatalf("unmarshal payload failed: %v", err)
	}

	if out.Event.Topic != in.Event.Topic {
		t.Fatalf("topic mismatch: got %s want %s", out.Event.Topic, in.Event.Topic)
	}
	if string(out.Event.Payload) != string(raw) {
		t.Fatalf("payload raw mismatch: got %s want %s", string(out.Event.Payload), string(raw))
	}
	if out.Event.EventName != in.Event.EventName {
		t.Fatalf("event name mismatch: got %s want %s", out.Event.EventName, in.Event.EventName)
	}
}
