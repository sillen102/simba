package simbaContext

type LoggerContextKey string
type RequestContextKey string
type TraceIDContextKey string
type ConnectionIDContextKey string

const (
	LoggerKey          LoggerContextKey       = "logger"
	TraceIDKey         TraceIDContextKey      = "traceId"
	TraceIDHeader      string                 = "X-Trace-Id"
	RequestSettingsKey RequestContextKey      = "requestSettings"
	ConnectionIDKey    ConnectionIDContextKey = "connectionId"
)
