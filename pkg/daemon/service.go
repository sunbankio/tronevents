package daemon

import (
	"context"
	"encoding/json"
	"log"
	"time"

	goRedis "github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"github.com/sunbankio/tronevents/pkg/config"
	"github.com/sunbankio/tronevents/pkg/monitoring"
	"github.com/sunbankio/tronevents/pkg/publisher"
	redisPkg "github.com/sunbankio/tronevents/pkg/redis"
	tronScanner "github.com/sunbankio/tronevents/pkg/scanner"
	"github.com/sunbankio/tronevents/pkg/storage"
	"github.com/sunbankio/tronevents/pkg/worker"
)

// Service orchestrates all the components of the daemon.
type Service struct {
	config        *config.Config
	redisClient   *goRedis.Client
	asynqClient   *asynq.Client
	asynqServer   *asynq.Server
	tronScanner   *tronScanner.Scanner
	storage       *storage.RedisStorage
	publisher     *publisher.EventPublisher
	workerManager *worker.Manager
	monitoring    *monitoring.Metrics
	logger        *log.Logger
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

	storage := storage.NewRedisStorage(goRedisClient, "last_synced_block")
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
	// Create a background context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the worker manager
	if err := s.workerManager.Start(); err != nil {
		s.logger.Fatal("Failed to start worker manager: ", err)
	}

	// Main processing loop
	s.runLoop(ctx)
}

// RunWithContext starts the daemon service with a context for cancellation
func (s *Service) RunWithContext(ctx context.Context) {
	// Start the worker manager
	if err := s.workerManager.Start(); err != nil {
		s.logger.Fatal("Failed to start worker manager: ", err)
	}

	// Main processing loop
	s.runLoop(ctx)
}

// runLoop contains the main processing logic
func (s *Service) runLoop(ctx context.Context) {
	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			s.logger.Println("Context cancelled, shutting down...")
			return
		default:
			// Continue with normal processing
		}

		// Read last_synced_block
		lastSyncedBlock, err := s.storage.Load(ctx)
		if err != nil {
			s.logger.Println("Error loading last synced block: ", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Use scanner.Scan(0) to get current block
		returnedBlockNum, _, transactions, err := s.tronScanner.Scan(ctx, 0)
		if err != nil {
			s.logger.Println("Error scanning current block: ", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Check if lastSyncedBlock == returnedBlockNum
		if lastSyncedBlock == returnedBlockNum {
			// Wait one second and continue
			time.Sleep(1 * time.Second)
			continue
		}

		// Publish result to Redis stream (transactions from the current block)
		for _, tx := range transactions {
			if err := s.publisher.Publish(context.Background(), &tx); err != nil {
				s.logger.Printf("Error publishing transaction: %v", err)
			}
		}

		// Program first run, or we are in sync, no backlog
		// if last synced block not exists or zero or returned block number = last_synced_block+1
		if lastSyncedBlock == 0 || returnedBlockNum == lastSyncedBlock+1 {
			// Update last synced_block
			if err := s.storage.Save(ctx, returnedBlockNum); err != nil {
				s.logger.Printf("Error saving last synced block: %v", err)
			}
			// Wait 3 seconds and continue
			time.Sleep(3 * time.Second)
			continue
		}

		// Slight backlog: we are lagging, but at most 20 blocks
		// blocks from last_synced+1 to returned_block-1 (inclusive) are missing
		if returnedBlockNum <= lastSyncedBlock+20 {
			// Push the gapped blocks' block numbers into a priority queue
			for blockNum := lastSyncedBlock + 1; blockNum < returnedBlockNum; blockNum++ {
				payload, err := json.Marshal(map[string]interface{}{"block_number": blockNum})
				if err != nil {
					s.logger.Printf("Error marshaling payload for block %d: %v", blockNum, err)
					continue
				}
				task := asynq.NewTask("block:process", payload)
				if _, err := s.asynqClient.Enqueue(task, asynq.Queue("priority")); err != nil {
					s.logger.Printf("Error enqueuing block %d: %v", blockNum, err)
				}
			}

			// Update last synced_block
			if err := s.storage.Save(ctx, returnedBlockNum); err != nil {
				s.logger.Printf("Error saving last synced block: %v", err)
			}
			// Wait 3 seconds and continue
			time.Sleep(3 * time.Second)
			continue
		}

		// Large backlog: we are lagging more than 20 blocks
		// last_synced+1 to returned_block-1 are all missing
		// however, we will backlog at most 7 days backlog (201600 blocks)
		// push the gapped blocks' block number range from max(returned_block-201600, last_synced+1) to returned_block-1, into backlog queue
		startBlock := lastSyncedBlock + 1
		maxStartBlock := returnedBlockNum - 201600
		if maxStartBlock > startBlock {
			startBlock = maxStartBlock
		}

		for blockNum := startBlock; blockNum < returnedBlockNum; blockNum++ {
			payload, err := json.Marshal(map[string]interface{}{"block_number": blockNum})
			if err != nil {
				s.logger.Printf("Error marshaling payload for block %d: %v", blockNum, err)
				continue
			}
			task := asynq.NewTask("block:process", payload)
			if _, err := s.asynqClient.Enqueue(task, asynq.Queue("backlog")); err != nil {
				s.logger.Printf("Error enqueuing block %d: %v", blockNum, err)
			}
		}

		// Update last synced_block
		if err := s.storage.Save(ctx, returnedBlockNum); err != nil {
			s.logger.Printf("Error saving last synced block: %v", err)
		}
		// Continue (no wait in large backlog case)
	}
}
