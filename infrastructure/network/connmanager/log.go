package connmanager

import (
	"github.com/sedracoin/sedrad/infrastructure/logger"
	"github.com/sedracoin/sedrad/util/panics"
)

var log = logger.RegisterSubSystem("CMGR")
var spawn = panics.GoroutineWrapperFunc(log)
