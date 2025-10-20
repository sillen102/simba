package simbaContext

type LoggerContextKey string
type RequestContextKey string
type TraceIDContextKey string

const (
	LoggerKey          LoggerContextKey  = "logger"
	TraceIDKey         TraceIDContextKey = "requestId"
	TraceIDHeader      string            = "X-Trace-Id"
	RequestSettingsKey RequestContextKey = "requestSettings"
)
