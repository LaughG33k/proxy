package proxy

type LevelLog struct{}

var (
	Error LevelLog = LevelLog{}
	Info  LevelLog = LevelLog{}
	Log   LevelLog = LevelLog{}
	Panic LevelLog = LevelLog{}
)

type Logger interface {
	Log(l LevelLog, args ...interface{})
}
