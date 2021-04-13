package nfs

import (
	"time"

	"github.com/micro-plat/hydra/context"
	"github.com/micro-plat/lib4go/logger"
)

type reqLog struct {
	start time.Time
	ctx   context.IContext
}

type log struct {
	start time.Time
	log   *logger.Logger
}

func start(input ...interface{}) log {
	l := log{
		start: time.Now(),
		log:   logger.New("nfs"),
	}
	p := make([]interface{}, 0, len(input)+1)
	p = append(p, "event.start > ")
	p = append(p, input...)
	l.log.Debug(p...)
	return l
}
func (l log) end(input ...interface{}) {
	p := make([]interface{}, 0, len(input)+2)
	p = append(p, "event.end > ")
	p = append(p, input...)
	p = append(p, time.Since(l.start))
	l.log.Debug(p...)
}