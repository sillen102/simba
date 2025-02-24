package simbaContext

type LoggerContextKey string
type RequestContextKey string
type RequestIdContextKey string

const (
	LoggerKey          LoggerContextKey    = "logger"
	RequestIDKey       RequestIdContextKey = "requestId"
	RequestIDHeader    string              = "X-Request-Id"
	RequestSettingsKey RequestContextKey   = "requestSettings"
)
