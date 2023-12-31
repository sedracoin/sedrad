package main

import (
	"github.com/sedracoin/sedrad/infrastructure/logger"
	"github.com/sedracoin/sedrad/util/panics"
)

var (
	backendLog = logger.NewBackend()
	log        = backendLog.Logger("APLG")
	spawn      = panics.GoroutineWrapperFunc(log)
)
