package daemon

import (
	"context"
	"encoding/json"
	"time"

	goRedis "github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"github.com/sunbankio/tronevents/pkg/config"
	"github.com/sunbankio/tronevents/pkg/logging"
	"github.com/sunbankio/tronevents/pkg/publisher"
	redisPkg "github.com/sunbankio/tronevents/pkg/redis"

	tronScanner "github.com/sunbankio/tronevents/pkg/scanner"
	"github.com/sunbankio/tronevents/pkg/storage"
	"github.com/sunbankio/tronevents/pkg/worker"
)

const (
	WaitInterval = 3100 * time.Millisecond // 3100ms wait interval
)

// Service orchestrates all the components of the daemon.
type Service struct {
	config                *config.Config
	redisClient           *goRedis.Client
	asynqClient           *asynq.Client
	asynqServer           *asynq.Server
	tronScanner           *tronScanner.Scanner
	lastSyncedBlock       *storage.LastSyncedStorage
	publisher             *publisher.EventPublisher
	workerManager         *worker.Manager
	logger                *logging.Logger
	blockProcessedStorage *storage.BlockProcessedStorage
}

// WorkerManager returns the worker manager for shutdown handling.
func (s *Service) WorkerManager() *worker.Manager {
	return s.workerManager
}

// Logger returns the logger.
func (s *Service) Logger() *logging.Logger {
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
		cfg.Queue.ToAsynqConfig(),
	)

	// Initialize TRON scanner - using the node address and Tron settings from config
	nodeURL := cfg.Tron.NodeURL
	if nodeURL == "" {
		nodeURL = "localhost:50051" // Default address
	}
	tronScannerInstance, err := tronScanner.NewScanner(nodeURL, cfg.Tron.Timeout, cfg.Tron.PoolSize, cfg.Tron.MaxPoolSize)
	if err != nil {
		panic(err)
	}

	// Use configurable Redis prefix
	redisPrefix := cfg.Redis.Prefix
	if redisPrefix == "" {
		redisPrefix = "tron" // Default prefix
	}
	lastSyncedBlockStorage := storage.NewLastSyncedStorage(goRedisClient, redisPrefix+":last_synced_block")
	blockProcessedStorage := storage.NewBlockProcessedStorage(goRedisClient, redisPrefix+":processed_blocks")
	publisher := publisher.NewEventPublisher(goRedisClient)
	workerManager := worker.NewManager(asynqServer, logging.NewLogger(cfg.LogLevel))

	return &Service{
		config:                cfg,
		redisClient:           goRedisClient,
		asynqClient:           asynqClient,
		asynqServer:           asynqServer,
		tronScanner:           tronScannerInstance,
		lastSyncedBlock:       lastSyncedBlockStorage,
		publisher:             publisher,
		workerManager:         workerManager,
		logger:                logging.NewLogger(cfg.LogLevel),
		blockProcessedStorage: blockProcessedStorage,
	}
}

// RunWithContext starts the daemon service with a context for cancellation
func (s *Service) RunWithContext(ctx context.Context) {
	// Start cleanup process to remove entries older than 7 days, running every hour
	s.blockProcessedStorage.StartCleanup(ctx, 1*time.Hour, 7*24*time.Hour)

	// Create and register the proper task handler for the worker
	handler := worker.NewHandler(s.tronScanner, s.publisher, s.blockProcessedStorage, s.logger)
	mux := asynq.NewServeMux()
	worker.RegisterHandlers(mux, handler)

	// Start the worker manager with the proper handler
	if err := s.workerManager.StartWithMux(mux); err != nil {
		s.logger.Fatal("Failed to start worker manager: ", err)
	}

	// Main processing loop
	s.runLoop(ctx)
}

// batchEnqueueBlocks enqueues multiple blocks in batch to reduce Redis operations
func (s *Service) batchEnqueueBlocks(blockNumbers []int64, queueName string) {
	if len(blockNumbers) == 0 {
		return
	}

	// Process in smaller batches to avoid overwhelming Redis
	batchSize := 100 // Process up to 100 blocks at a time
	for i := 0; i < len(blockNumbers); i += batchSize {
		end := i + batchSize
		if end > len(blockNumbers) {
			end = len(blockNumbers)
		}

		batch := blockNumbers[i:end]

		// Create all tasks first, then enqueue them using Asynq's client
		for _, blockNum := range batch {
			payload, err := json.Marshal(worker.BlockProcessPayload{BlockNumber: blockNum})
			if err != nil {
				s.logger.Printf("Error marshaling payload for block %d: %v", blockNum, err)
				continue
			}
			task := asynq.NewTask("block:process", payload)
			if _, err := s.asynqClient.Enqueue(task, asynq.Queue(queueName), asynq.MaxRetry(5)); err != nil {
				s.logger.Printf("Error enqueuing block %d: %v", blockNum, err)
			} else {
				s.logger.Debugf("Successfully enqueued block %d to %s queue", blockNum, queueName)
			}
		}
	}
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
		lastSyncedBlock, err := s.lastSyncedBlock.Load(ctx)
		if err != nil {
			s.logger.Println("Error loading last synced block: ", err)
			time.Sleep(1 * time.Second)
			continue
		}
		s.logger.Debugf("Loaded last_synced_block = %d", lastSyncedBlock)

		// Use scanner.Scan(0) to get current block
		s.logger.Debugf("Scanning current block with Scan(ctx, 0)")
		returnedBlockNum, returnedBlockTime, transactions, err := s.tronScanner.Scan(ctx, 0)
		if err != nil {
			s.logger.Println("Error scanning current block: ", err)
			time.Sleep(1 * time.Second)
			continue
		}
		s.logger.Debugf("Scan completed - returned block: %d, transactions count: %d", returnedBlockNum, len(transactions))

		// Check if lastSyncedBlock == returnedBlockNum
		if lastSyncedBlock == returnedBlockNum {
			s.logger.Debugf("lastSyncedBlock (%d) == returnedBlockNum (%d), waiting 1 second", lastSyncedBlock, returnedBlockNum)
			// Wait one second and continue
			waitUntil(returnedBlockTime.Add(WaitInterval))
			continue
		}

		// Publish result to Redis stream (transactions from the current block) in batch
		// Convert slice of values to slice of pointers for batch publishing
		transactionPointers := make([]*tronScanner.Transaction, len(transactions))
		for i := range transactions {
			transactionPointers[i] = &transactions[i]
		}
		if err := s.publisher.PublishBatch(context.Background(), transactionPointers); err != nil {
			s.logger.Printf("Error publishing batch of transactions: %v", err)
		}

		// Mark the current block as processed to prevent duplicate processing
		if err := s.blockProcessedStorage.MarkProcessed(ctx, returnedBlockNum); err != nil {
			s.logger.Errorf("[MAIN]    Failed to mark block %d as processed: %v", returnedBlockNum, err)
		}

		s.logger.Infof("[MAIN]   Block %d scanned, published %d transactions.", returnedBlockNum, len(transactions))

		// Program first run, or we are in sync, no backlog
		// if last synced block not exists or zero or returned block number = last_synced_block+1
		if lastSyncedBlock == 0 || returnedBlockNum == lastSyncedBlock+1 {
			s.logger.Debugf("First run or in sync - lastSyncedBlock: %d, returnedBlockNum: %d", lastSyncedBlock, returnedBlockNum)
			s.updateLastSyncedBlock(ctx, returnedBlockNum)
			waitUntil(returnedBlockTime.Add(WaitInterval))
			continue
		}

		// Slight backlog: we are lagging, but at most 20 blocks
		// blocks from last_synced+1 to returned_block-1 (inclusive) are missing
		if returnedBlockNum <= lastSyncedBlock+20 {
			s.logger.Debugf("Slight backlog detected - lastSynced: %d, returned: %d, gap: %d blocks", lastSyncedBlock, returnedBlockNum, returnedBlockNum-lastSyncedBlock-1)

			// Batch enqueue blocks to reduce Redis operations
			gap := returnedBlockNum - lastSyncedBlock - 1
			if gap <= 0 {
				gap = 0
			}
			blockNumbers := make([]int64, 0, gap)
			for blockNum := lastSyncedBlock + 1; blockNum < returnedBlockNum; blockNum++ {
				blockNumbers = append(blockNumbers, blockNum)
			}

			s.batchEnqueueBlocks(blockNumbers, "priority")
			s.updateLastSyncedBlock(ctx, returnedBlockNum)
			waitUntil(returnedBlockTime.Add(WaitInterval))
			continue
		}

		// Large backlog: we are lagging more than 20 blocks
		// last_synced+1 to returned_block-1 are all missing
		// however, we will backlog at most 7 days backlog (201600 blocks)
		s.logger.Debugf("Large backlog detected - lastSynced: %d, returned: %d, gap: %d blocks", lastSyncedBlock, returnedBlockNum, returnedBlockNum-lastSyncedBlock-1)

		// push the gapped blocks' block number range from max(returned_block-201600, last_synced+1) to returned_block-1, into backlog queue
		startBlock := lastSyncedBlock + 1
		maxStartBlock := returnedBlockNum - 201600
		if maxStartBlock > startBlock {
			startBlock = maxStartBlock
		}

		s.logger.Debugf("Large backlog - queuing range: %d to %d", startBlock, returnedBlockNum-1)

		// Batch enqueue blocks to reduce Redis operations
		gap := returnedBlockNum - startBlock
		if gap <= 0 {
			gap = 0
		}
		blockNumbers := make([]int64, 0, gap)
		for blockNum := startBlock; blockNum < returnedBlockNum; blockNum++ {
			blockNumbers = append(blockNumbers, blockNum)
		}

		s.batchEnqueueBlocks(blockNumbers, "backlog")
		s.updateLastSyncedBlock(ctx, returnedBlockNum)
		waitUntil(returnedBlockTime.Add(WaitInterval))
	}
}

func (s *Service) updateLastSyncedBlock(ctx context.Context, blockNumber int64) {
	if err := s.lastSyncedBlock.Save(ctx, blockNumber); err != nil {
		s.logger.Printf("Error saving last synced block: %v", err)
	}
	s.logger.Debugf("Updated last synced block to %d", blockNumber)
}

func waitUntil(until time.Time) {
	now := time.Now()
	if now.After(until) {
		return
	}
	time.Sleep(until.Sub(now))
}
