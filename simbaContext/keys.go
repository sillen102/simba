package simbaContext

type AuthContextKey string
type LoggerContextKey string
type RequestContextKey string
type RequestIdContextKey string

const (
	AuthFuncKey        AuthContextKey      = "authFunc"
	LoggerKey          LoggerContextKey    = "logger"
	RequestIDKey       RequestIdContextKey = "requestId"
	RequestIDHeader    string              = "X-Request-Id"
	RequestSettingsKey RequestContextKey   = "requestSettings"
)
