package daemon

import (
	"log"

	goRedis "github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"github.com/sunbankio/tronevents/pkg/config"
	"github.com/sunbankio/tronevents/pkg/monitoring"
	"github.com/sunbankio/tronevents/pkg/processor"
	"github.com/sunbankio/tronevents/pkg/publisher"
	redisPkg "github.com/sunbankio/tronevents/pkg/redis"
	tronScanner "github.com/sunbankio/tronevents/pkg/scanner"
	"github.com/sunbankio/tronevents/pkg/storage"
	"github.com/sunbankio/tronevents/pkg/worker"
)

// Service orchestrates all the components of the daemon.
type Service struct {
	config         *config.Config
	redisClient    *goRedis.Client
	asynqClient    *asynq.Client
	asynqServer    *asynq.Server
	tronScanner    *tronScanner.Scanner
	storage        *storage.BlockStorage
	publisher      *publisher.EventPublisher
	workerManager  *worker.Manager
	monitoring     *monitoring.Metrics
	logger         *log.Logger
}

// WorkerManager returns the worker manager for shutdown handling.
func (s *Service) WorkerManager() *worker.Manager {
	return s.workerManager
}

// Logger returns the logger.
func (s *Service) Logger() *log.Logger {
	return s.logger
}

// NewService creates a new daemon Service.
func NewService(cfg *config.Config) *Service {
	// Initialize components
	goRedisClient, err := redisPkg.NewClient(cfg.Redis)
	if err != nil {
		panic(err)
	}

	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.Redis.Addr})
	asynqServer := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Redis.Addr},
		worker.DefaultConfig(),
	)

	// Initialize TRON scanner - using the node address from config
	nodeURL := cfg.TronNodeURL
	if nodeURL == "" {
		nodeURL = "localhost:50051" // Default address
	}
	tronScannerInstance, err := tronScanner.NewScanner(nodeURL)
	if err != nil {
		panic(err)
	}

	storage := storage.NewBlockStorage("last_synced_block.json")
	publisher := publisher.NewEventPublisher(goRedisClient)
	workerManager := worker.NewManager(asynqServer, log.Default())
	monitoring := monitoring.NewMetrics()

	return &Service{
		config:        cfg,
		redisClient:   goRedisClient,
		asynqClient:   asynqClient,
		asynqServer:   asynqServer,
		tronScanner:   tronScannerInstance,
		storage:       storage,
		publisher:     publisher,
		workerManager: workerManager,
		monitoring:    monitoring,
		logger:        log.Default(),
	}
}

// Run starts the daemon service.
func (s *Service) Run() {
	// Start the worker manager
	if err := s.workerManager.Start(); err != nil {
		s.logger.Fatal("Failed to start worker manager: ", err)
	}

	// Main processing loop
	for {
		// Fetch the current block number
		currentBlock, err := s.tronScanner.GetCurrentBlockNumber()
		if err != nil {
			s.logger.Println("Error getting current block number: ", err)
			continue
		}

		// Load the last synced block number
		lastSyncedBlock, err := s.storage.Load()
		if err != nil {
			s.logger.Println("Error loading last synced block: ", err)
			continue
		}

		// Determine processing mode based on block numbers
		if lastSyncedBlock == 0 || currentBlock == lastSyncedBlock+1 {
			// First run or in sync
			processor.SyncLogic(s.tronScanner, s.publisher, s.storage, s.logger)
		} else if currentBlock-lastSyncedBlock <= 20 {
			// Slight backlog
			processor.SlightBacklog(s.tronScanner, s.asynqClient, s.storage, s.logger)
		} else {
			// Large backlog
			processor.LargeBacklog(s.tronScanner, s.asynqClient, s.storage, s.logger)
		}
	}
}