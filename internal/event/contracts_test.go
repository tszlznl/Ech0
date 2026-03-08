package event

import "testing"

func TestTopicConstants_AreUniqueAndStable(t *testing.T) {
	topics := []string{
		TopicUserCreated,
		TopicUserUpdated,
		TopicUserDeleted,
		TopicEchoCreated,
		TopicEchoUpdated,
		TopicEchoDeleted,
		TopicResourceUploaded,
		TopicSystemBackup,
		TopicSystemRestore,
		TopicSystemExport,
		TopicBackupScheduleUpdate,
		TopicDeadLetterRetried,
		TopicInboxClear,
		TopicEch0UpdateCheck,
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
