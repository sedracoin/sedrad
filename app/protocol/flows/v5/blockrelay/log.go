package blockrelay

import (
	"github.com/sedracoin/sedrad/infrastructure/logger"
	"github.com/sedracoin/sedrad/util/panics"
)

var log = logger.RegisterSubSystem("PROT")
var spawn = panics.GoroutineWrapperFunc(log)
