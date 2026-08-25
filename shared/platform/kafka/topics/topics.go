package topics

// All returns the complete topic catalog owned by the Notegic runtimes.
// Every entry is constructed by its topic owner with explicit broker settings.
func All() []TopicSpec {
	return []TopicSpec{
		CoreLifecycleTopicSpec(),
		CoreYjsMaintenanceHintTopicSpec(),
		CoreNotificationTopicSpec(),
		DurableJobRealtimeGatewayRoutineTaskLifecycleTopicSpec(),
		CoreEmailRequestTopicSpec(),
		NotificationTopicSpec(),
		YjsWorkerCoreCommandTopicSpec(),
		CoreYjsWorkerReplyTopicSpec(),
		YjsWorkerCoreMaintenanceCommandTopicSpec(),
		CoreYjsWorkerMaintenanceResultTopicSpec(),
	}
}
