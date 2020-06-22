package event

// EntityType defines a type of entity ot be used in an event
type EntityType string

const (
	// TeamEntity is a team entity
	TeamEntity EntityType = "team"
	// TrackerEntity is a tracker entity
	TrackerEntity EntityType = "tracker"
	// CollectionEntity is a collection entity
	CollectionEntity EntityType = "collection"
	// SubscriptionEntity is a subscription entity
	SubscriptionEntity EntityType = "subscription"
	// ShapeCollectionEntity is a shape collection entity
	ShapeCollectionEntity EntityType = "shapecollection"
)
