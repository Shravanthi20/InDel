package kafka

// Topic constants for Kafka message bus
const (
	TopicClaimsGenerated     = "indel.claims.generated"
	TopicClaimsScored        = "indel.claims.scored"
	TopicPayoutsQueued       = "indel.payouts.queued"
	TopicPayoutsFailed       = "indel.payouts.failed"
	TopicWeatherAlerts       = "indel.weather.alerts"
	TopicAQIAlerts           = "indel.aqi.alerts"
	TopicDisruptionConfirmed = "indel.disruption.confirmed"
	TopicOrderDrop           = "indel.zone.order-drop"
	TopicEarningsSettled     = "indel.earnings.settled"
	TopicClaimReviewed       = "indel.claims.reviewed"

	// Disruption event topics
	TopicClaimDisruptionCreated = "claim.disruption.created"
	TopicClaimDisruptionUpdated = "claim.disruption.updated"

	// Policy lock-in lifecycle topics
	TopicPolicyCreated      = "policy.created"
	TopicPolicyLocked       = "policy.locked"
	TopicPolicyActivated    = "policy.activated"
	TopicPurchaseBlocked    = "policy.purchase_blocked"

	// Claim rejection topic (used when policy is in lock-in period)
	TopicClaimRequested = "claim.requested"
	TopicClaimRejected  = "claim.rejected"

	// Disruption signal topics (used for purchase blocking)
	TopicDisruptionActive    = "disruption.active"
	TopicDisruptionPredicted = "disruption.predicted"
)
