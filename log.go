package smq

// Log 日志接口
type Log interface {
	Info(msg string)
	Error(msg string)
}
