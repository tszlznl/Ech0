package event

import (
	"testing"

	contracts "github.com/lin-snow/ech0/internal/event/contracts"
)

func TestTopicConstants_AreUniqueAndStable(t *testing.T) {
	topics := []string{
		contracts.TopicUserCreated,
		contracts.TopicUserUpdated,
		contracts.TopicUserDeleted,
		contracts.TopicEchoCreated,
		contracts.TopicEchoUpdated,
		contracts.TopicEchoDeleted,
		contracts.TopicResourceUploaded,
		contracts.TopicSystemBackup,
		contracts.TopicSystemRestore,
		contracts.TopicSystemExport,
		contracts.TopicBackupScheduleUpdate,
		contracts.TopicDeadLetterRetried,
		contracts.TopicInboxClear,
		contracts.TopicEch0UpdateCheck,
	}

	seen := map[string]struct{}{}
	for _, topic := range topics {
		if topic == "" {
			t.Fatalf("topic should not be empty")
		}
		if _, ok := seen[topic]; ok {
			t.Fatalf("duplicated topic found: %s", topic)
		}
		seen[topic] = struct{}{}
	}
}
