package common

import (
	"fmt"
	"github.com/sedracoin/sedrad/domain/dagconfig"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
)

// RunSedradForTesting runs sedrad for testing purposes
func RunSedradForTesting(t *testing.T, testName string, rpcAddress string) func() {
	appDir, err := TempDir(testName)
	if err != nil {
		t.Fatalf("TempDir: %s", err)
	}

	sedradRunCommand, err := StartCmd("SEDRAD",
		"sedrad",
		NetworkCliArgumentFromNetParams(&dagconfig.DevnetParams),
		"--appdir", appDir,
		"--rpclisten", rpcAddress,
		"--loglevel", "debug",
	)
	if err != nil {
		t.Fatalf("StartCmd: %s", err)
	}
	t.Logf("sedrad started with --appdir=%s", appDir)

	isShutdown := uint64(0)
	go func() {
		err := sedradRunCommand.Wait()
		if err != nil {
			if atomic.LoadUint64(&isShutdown) == 0 {
				panic(fmt.Sprintf("sedrad closed unexpectedly: %s. See logs at: %s", err, appDir))
			}
		}
	}()

	return func() {
		err := sedradRunCommand.Process.Signal(syscall.SIGTERM)
		if err != nil {
			t.Fatalf("Signal: %s", err)
		}
		err = os.RemoveAll(appDir)
		if err != nil {
			t.Fatalf("RemoveAll: %s", err)
		}
		atomic.StoreUint64(&isShutdown, 1)
		t.Logf("sedrad stopped")
	}
}
