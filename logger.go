package level

type logger interface {
	Error(string, ...interface{})
	Warning(string, ...interface{})
	Info(string, ...interface{})
	Debug(string, ...interface{})
}

type loggable interface {
	Logger(logger)
}

type nilLogger struct{}

func (self *nilLogger) Error(_ string, _ ...interface{})   {}
func (self *nilLogger) Warning(_ string, _ ...interface{}) {}
func (self *nilLogger) Info(_ string, _ ...interface{})    {}
func (self *nilLogger) Debug(_ string, _ ...interface{})   {}
