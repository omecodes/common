package prompt

import (
	"os"
	"os/signal"
	"syscall"
)

func QuitSignal() chan os.Signal {
	stopSig := make(chan os.Signal, 1)
	signal.Notify(stopSig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	return stopSig
}
