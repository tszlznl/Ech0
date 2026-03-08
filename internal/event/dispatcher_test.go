package event

import (
	"encoding/json"
	"testing"

	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
)

func TestWebhookReplayPayload_RoundTripKeepsPayloadRaw(t *testing.T) {
	raw := json.RawMessage(`{"id":1,"name":"echo"}`)
	in := WebhookReplayPayload{
		Webhook: webhookModel.Webhook{
			Name: "hook",
			URL:  "https://example.com/hook",
		},
		Event: WebhookObservation{
			Topic:      TopicEchoCreated,
			EventName:  "EchoCreatedEvent",
			Payload:    raw,
			Metadata:   map[string]string{MetaKeySource: "test"},
			OccurredAt: 123,
		},
	}

	buf, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal payload failed: %v", err)
	}

	var out WebhookReplayPayload
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
