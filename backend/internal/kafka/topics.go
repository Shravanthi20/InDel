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
)
