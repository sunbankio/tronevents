package daemon

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sunbankio/tronevents/pkg/worker"
)

// Shutdown handles the graceful shutdown of the daemon.
func Shutdown(workerManager *worker.Manager, logger *log.Logger) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	<-ch

	logger.Println("Shutting down...")
	workerManager.Stop()
}
