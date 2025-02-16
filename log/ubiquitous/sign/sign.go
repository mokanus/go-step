package sign

type Sign string

func (s Sign) String() string {
	return string(s)
}

const (
	LOGGER         Sign = "logger"
	TRACE_ID       Sign = "trace_id"
	SPAN_ID        Sign = "span_id"
	PARENT_SPAN_ID Sign = "parent_span_id"
	REQUEST        Sign = "request"
	RESPONSE       Sign = "response"
	SESSION        Sign = "session"
	Error          Sign = "error"
	TRANSACTION    Sign = "trans"
	DIS_LOCK       Sign = "distribute_lock"
	IS_GLOBAL_LOCK Sign = "is_global_lock"
)
