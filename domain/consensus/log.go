package consensus

import (
	"github.com/sedracoin/sedrad/infrastructure/logger"
	"github.com/sedracoin/sedrad/util/panics"
)

var log = logger.RegisterSubSystem("BDAG")
var spawn = panics.GoroutineWrapperFunc(log)
