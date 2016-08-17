package log

// Standard log field names.
const (
	FnTopic          = "topic"
	FnLoggedAt       = "logged_at"
	FnSeverity       = "severity"
	FnUtsname        = "utsname"
	FnMessage        = "message"
	FnSecret         = "secret"
	FnRequestID      = "request_id"
	FnOperationID    = "operation_id"
	FnResponseTime   = "response_time"
	FnRemoteAddress  = "remote_ipaddr"
	FnURL            = "url"
	FnHTTPMethod     = "http_method"
	FnHTTPVersion    = "http_version"
	FnHTTPStatusCode = "http_status_code"
	FnHTTPReferer    = "http_referer"
	FnHTTPUserAgent  = "http_user_agent"
	FnRequestSize    = "request_size"
	FnResponseSize   = "response_size"
	FnDomain         = "domain"
	FnService        = "service"
	FnTrackingCookie = "tracking_cookie"
	FnBrowser        = "browser"
	FnServiceSet     = "serviceset"
	FnStartAt        = "start_at"
)

// Severities a.k.a log levels.
const (
	LvCritical = 2
	LvError    = 3
	LvWarn     = 4
	LvInfo     = 6
	LvDebug    = 7
)

const (
	// The maximum length of a formatted log message.
	maxLogSize = 1 << 20

	// The maximum length of topic.
	maxTopicLength = 128
)

const (
	// RFC3339Micro is for time.Time.Format().
	RFC3339Micro = "2006-01-02T15:04:05.000000Z07:00"
)