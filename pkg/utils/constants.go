package utils

// gRPC constants
const (
	DefaultGRPCPort = "9090"
)

// http constants
const (
	OrganizationIDPathPrefix   = "orgid"
	RegionPathKey              = "regionID"
	LocationPathKey            = "locationID"
	ChannelPathKey             = "channelID"
	UserPathKey                = "userID"
	FeaturePathKey             = "feature"
	IdentityUserPathKey        = "identityUserID"
	MachineAccountIDPathPrefix = "machineAccountID"
	EntityNameKey              = "entityName"
	TrackingIDPrefix           = "helloworld"
	ApplicationNameKey         = "helloworld"
	DatabaseName               = "helloworld"

	DefaultAppName     = "sre-go-helloworld"
	DefaultHTTPPort    = "8080"
	ConfigTypeUser     = "user_config"
	ConfigTypeLogin    = "user_login"
	ConfigTypeLogout   = "user_logout"
	ConfigTypeLocation = "location_config"
	MetricTag          = "metric"
)

const (
	//DefaultInfluxDBName holds default DB name for Influx DB used by Data Service
	DefaultInfluxDBName = "sreInfluxDB"
)
