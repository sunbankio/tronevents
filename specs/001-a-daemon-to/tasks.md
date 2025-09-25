# Tasks: TRON Events Redis Stream Daemon

This task list implements the TRON Events Redis Stream Daemon feature according to the specification and implementation plan. The daemon continuously monitors the TRON blockchain for new events and publishes them to a Redis stream with 7-day retention, handling interruptions gracefully by processing missed blocks upon restart.

## T001: Project Setup and Dependencies
**Task**: Initialize Go project structure and install dependencies
**Files**: go.mod, go.sum
**Instructions**:
1. Create go.mod file with module name
2. Add required dependencies: github.com/hibiken/asynq, redis client library
3. Include existing TRON blockchain scanner from /scanner/ package
4. Run `go mod tidy` to resolve dependencies
**Dependencies**: None
**Status**: Pending

## T002: Create Core Data Models
**Task**: Implement data models for block numbers and adapt scanner.Transaction for Redis publishing [P]
**Files**: pkg/models/block.go, pkg/models/event_adapter.go
**Instructions**:
1. Create Block Number model as sequential identifier
2. Create adapter/wrapper for scanner.Transaction struct to optimize for Redis stream publishing
3. Define any additional metadata needed for Redis stream processing
4. Add appropriate JSON tags for serialization
**Dependencies**: T001
**Status**: Pending

## T003: Setup Redis Configuration and Connection
**Task**: Implement Redis client setup and configuration [P]
**Files**: pkg/redis/config.go, pkg/redis/client.go
**Instructions**:
1. Create Redis configuration struct with connection parameters
2. Implement Redis client initialization with connection pooling
3. Add connection validation and error handling
4. Set up Redis streams with 7-day expiration TTL
**Dependencies**: T001
**Status**: Pending

## T004: Setup Asynq Queue System
**Task**: Configure asynq queue system with priority, backlog, retry, and dead letter queues [P]
**Files**: pkg/queue/worker.go, pkg/queue/config.go
**Instructions**:
1. Configure asynq server with Redis connection
2. Create 4 queues: priority, backlog, retry, deadletter
3. Set up queue priorities and worker pool configuration
4. Configure retry policies with specified backoff times: 5, 10, 30, 60, 180, 300, 600, 1800, 3600 seconds
**Dependencies**: T001, T003
**Status**: Pending

## T005: Integrate with Existing Block Scanner Service
**Task**: Integrate with existing TRON blockchain scanner in /scanner/ package [P]
**Files**: pkg/scanner/client.go
**Instructions**:
1. Create wrapper/adapter for existing scanner.Scan(0) to read current block number
2. Implement error handling for scanner operations
3. Add connection retry mechanism with backoff
4. Create interface abstraction for scanner to enable testing
**Dependencies**: T001
**Status**: Pending

## T006: Create Block Persistence Mechanism
**Task**: Implement storage for last processed block number [P]
**Files**: pkg/storage/block_storage.go
**Instructions**:
1. Create persistence mechanism for last_synced_block tracking
2. Implement save/load functions for block number
3. Add error handling for persistence failures
4. Ensure atomic updates to prevent corruption
**Dependencies**: T001
**Status**: Pending

## T007: Implement Normal Operation Logic
**Task**: Create logic for normal operation when last_synced_block matches returned block
**Files**: pkg/processor/normal_operation.go
**Instructions**:
1. Implement logic: when last_synced_block matches returned block, wait 1 second and continue
2. Process the scanner.Transaction structs returned by the scanner
3. Add event publishing of Transaction structs to Redis stream via publisher service
4. Include appropriate logging
**Dependencies**: T002, T003, T005, T006
**Status**: Pending

## T008: Implement First Run and Sync Logic
**Task**: Create logic for first run or when in-sync (no backlog)
**Files**: pkg/processor/sync_logic.go
**Instructions**:
1. Implement logic: when last_synced_block doesn't exist or is zero, or when returned_block_num=last_synced_block+1
2. Process the scanner.Transaction structs returned by the scanner
3. Update last_synced_block and wait 3 seconds
4. Add appropriate state tracking
**Dependencies**: T002, T003, T005, T006, T007
**Status**: Pending

## T009: Implement Slight Backlog Processing
**Task**: Create logic for processing slight backlog (up to 20 blocks)
**Files**: pkg/processor/slight_backlog.go
**Instructions**:
1. Implement logic: when lagging at most 20 blocks
2. Process scanner.Transaction structs for the gapped blocks
3. Push gapped blocks' block numbers to priority queue
4. Update last_synced_block and wait 3 seconds
5. Add logging for backlog processing
**Dependencies**: T002, T003, T004, T005, T006, T008
**Status**: Pending

## T010: Implement Large Backlog Processing
**Task**: Create logic for processing large backlog (more than 20 blocks)
**Files**: pkg/processor/large_backlog.go
**Instructions**:
1. Implement logic: when lagging more than 20 blocks
2. Process scanner.Transaction structs for the gapped blocks
3. Push block number range (max(returned_block-201600, last_synced+1) to returned_block-1) to backlog queue
4. Update last_synced_block
5. Implement LIFO processing (recent backlogs first)
**Dependencies**: T002, T003, T004, T005, T006, T009
**Status**: Pending

## T011: Create Worker Configuration
**Task**: Implement worker configuration with 15 concurrent workers
**Files**: pkg/worker/manager.go, pkg/worker/config.go
**Instructions**:
1. Configure max 15 workers total
2. Assign 1 worker for new blocks (generated every 3 seconds)
3. Assign 1 worker for priority queue
4. Assign up to 13 workers for backlog queue
5. Implement worker lifecycle management
**Dependencies**: T004
**Status**: Pending

## T012: Implement Daemon Shutdown Handling
**Task**: Create robust handling of server shutdown during backlog processing
**Files**: pkg/daemon/shutdown.go
**Instructions**:
1. Implement graceful shutdown for all worker types
2. Ensure no data loss during shutdown
3. Save current progress to persistent storage
4. Handle interruption during backlog processing
**Dependencies**: T010, T011
**Status**: Pending

## T013: Implement Event Publishing Service
**Task**: Create service to publish scanner.Transaction structs to Redis stream
**Files**: pkg/publisher/event_publisher.go
**Instructions**:
1. Implement function to format and publish scanner.Transaction structs to Redis stream
2. Use the adapter from T002 to convert/prepare transactions for Redis publishing
3. Ensure 7-day retention with automatic expiration
4. Add rate limiting for 500 events per 3 seconds
5. Include error handling and retry logic
**Dependencies**: T002, T003
**Status**: Pending

## T014: Create Main Daemon Service
**Task**: Integrate all components into main daemon service
**Files**: cmd/daemon/main.go, pkg/daemon/service.go
**Instructions**:
1. Create main daemon service that orchestrates all components
2. Implement the main processing loop with all states (normal, first run, slight backlog, large backlog)
3. Integrate existing block scanner (returns scanner.Transaction structs), event publisher, and worker manager
4. Process scanner.Transaction structs through the appropriate workflow based on system state
5. Add signal handling for graceful shutdown
**Dependencies**: T007, T008, T009, T010, T011, T012, T013
**Status**: Pending

## T015: Implement Logging and Monitoring
**Task**: Add structured logging and monitoring capabilities
**Files**: pkg/logging/logger.go, pkg/monitoring/metrics.go
**Instructions**:
1. Implement structured logging using a standard format
2. Add metrics collection for events processed, blocks scanned, queue status
3. Log important events for monitoring and debugging
4. Add error and warning level logging
**Dependencies**: T001
**Status**: Pending

## T016: Create Configuration Management
**Task**: Implement configuration system for daemon settings [P]
**Files**: pkg/config/config.go
**Instructions**:
1. Create configuration struct with all daemon settings
2. Add environment variable and file-based configuration loading
3. Include Redis connection settings, worker counts, timeouts
4. Add validation for configuration values
**Dependencies**: T001
**Status**: Pending

## T017: Unit Tests for Data Models
**Task**: Write unit tests for data models [P]
**Files**: pkg/models/event_adapter_test.go, pkg/models/block_test.go
**Instructions**:
1. Create unit tests for scanner.Transaction adapter/wrapper
2. Test Block Number model functionality
3. Test Transaction struct serialization/deserialization for Redis
4. Verify all data model methods work correctly
**Dependencies**: T002
**Status**: Pending

## T018: Unit Tests for Redis Operations
**Task**: Write unit tests for Redis operations [P]
**Files**: pkg/redis/client_test.go
**Instructions**:
1. Test Redis connection functionality
2. Test Redis stream operations
3. Test TTL and expiration functionality
4. Test error handling for Redis operations
**Dependencies**: T003
**Status**: Pending

## T019: Unit Tests for Queue System
**Task**: Write unit tests for asynq queue system [P]
**Files**: pkg/queue/worker_test.go
**Instructions**:
1. Test queue configuration and setup
2. Test task enqueuing for different queues
3. Test retry logic with specified backoff times
4. Test dead letter queue functionality
**Dependencies**: T004
**Status**: Pending

## T020: Unit Tests for Scanner Integration
**Task**: Write unit tests for scanner integration [P]
**Files**: pkg/scanner/client_test.go
**Instructions**:
1. Test scanner wrapper functionality
2. Test error handling for network issues
3. Test connection retry mechanisms
4. Verify block number reading accuracy from existing scanner
**Dependencies**: T005
**Status**: Pending

## T021: Integration Test for Normal Operation
**Task**: Create integration test for normal operation scenario [P]
**Files**: tests/integration/normal_operation_test.go
**Instructions**:
1. Test scenario: Given daemon is running normally, When new TRON events occur, Then scanner.Transaction structs are published to Redis stream in real-time
2. Connect to localhost TRON node to retrieve real blockchain data
3. Verify Redis stream contains expected Transaction structs from real blockchain
**Dependencies**: T007, T013
**Status**: Pending

## T022: Integration Test for 7-Day Retention
**Task**: Create integration test for 7-day retention [P]
**Files**: tests/integration/retention_test.go
**Instructions**:
1. Test scenario: Given daemon has been running for several days, When daemon continues to operate, Then events older than 7 days are automatically removed from Redis stream
2. Use real blockchain data from localhost TRON node to populate Redis stream
3. Verify TTL functionality works correctly with real Transaction structs
**Dependencies**: T003, T013
**Status**: Pending

## T023: Integration Test for Restart with Missed Blocks
**Task**: Create integration test for restart with missed blocks [P]
**Files**: tests/integration/restart_test.go
**Instructions**:
1. Test scenario: Given daemon is stopped unexpectedly, When daemon is restarted, Then it processes missed blocks and continues normal operation
2. Connect to localhost TRON node to retrieve real blockchain data for missed blocks
3. Verify scanner.Transaction structs from missed blocks are processed using real data
**Dependencies**: T006, T008, T014
**Status**: Pending

## T024: Integration Test for Simultaneous Processing
**Task**: Create integration test for simultaneous new and backlog processing [P]
**Files**: tests/integration/simultaneous_test.go
**Instructions**:
1. Test scenario: Given daemon processes new blocks while catching up, When new blocks are added, Then daemon processes both missed and new blocks simultaneously
2. Use real blockchain data from localhost TRON node for concurrent processing scenarios
3. Verify both queues are handled properly and Transaction structs from real blockchain are published correctly
**Dependencies**: T009, T010, T011
**Status**: Pending

## T025: Performance Test for Backlog Processing
**Task**: Create performance test for handling large backlog of scanner.Transaction structs
**Files**: tests/performance/backlog_test.go
**Instructions**:
1. Test the system's ability to robustly handle interruptions during the processing of a large backlog.
2. Connect to a localhost TRON node to retrieve a substantial amount of historical blockchain data to simulate a large backlog scenario.
3. While the system is processing the backlog, simulate an unexpected shutdown or interruption.
4. Restart the daemon and verify that it correctly resumes processing from the point of interruption without any data loss or duplication.
5. Confirm that the last processed block number was correctly persisted and reloaded.
6. Measure overall throughput and stability during the interruption and resumption cycle.
**Dependencies**: T013, T014
**Status**: Pending

## T026: Backlog Processing Test for 7 Days
**Task**: Create test for handling up to 201600 blocks backlog
**Files**: tests/integration/backlog_test.go
**Instructions**:
1. Test system's ability to handle up to 201600 blocks as backlog
2. Verify LIFO processing (recent blocks prioritized) of scanner.Transaction structs
3. Test worker distribution for large backlogs of Transaction structs
4. Verify system stability during large backlog processing
**Dependencies**: T010, T011, T014
**Status**: Pending

## T027: End-to-End Test for Daemon Functionality
**Task**: Create comprehensive end-to-end test [P]
**Files**: tests/e2e/daemon_e2e_test.go
**Instructions**:
1. Test complete daemon workflow from start to finish
2. Use real blockchain data from localhost TRON node for various scenarios: normal operation, restart, backlog processing
3. Verify all functional requirements are met with real Transaction structs
4. Test error conditions and recovery using real blockchain scenarios
**Dependencies**: T014, T021, T022, T023, T024
**Status**: Pending

## T028: Error Handling Tests
**Task**: Create tests for error handling scenarios [P]
**Files**: tests/integration/error_handling_test.go
**Instructions**:
1. Test Redis connection failures and recovery.
2. Test TRON blockchain scanner errors using localhost TRON node.
3. Test queue processing errors and retries with real Transaction structs.
4. Simulate and test daemon shutdown during various operations (e.g., normal sync, backlog processing).
5. Verify that the system performs a graceful shutdown, saves its state correctly, and resumes without data loss upon restart, particularly testing interruptions during the most critical phases like large backlog processing.
**Dependencies**: T003, T004, T005, T012
**Status**: Pending

## T029: CLI Interface Implementation
**Task**: Add CLI commands for daemon management [P]
**Files**: cmd/cli/main.go, cmd/cli/commands.go
**Instructions**:
1. Implement start command to launch daemon
2. Add status command to check daemon status
3. Add stop command for graceful shutdown
4. Add config command for configuration management
**Dependencies**: T014, T016
**Status**: Pending

## T030: Documentation and README
**Task**: Create documentation and README for the daemon
**Files**: README.md, docs/configuration.md, docs/troubleshooting.md
**Instructions**:
1. Create README with setup and usage instructions
2. Document configuration options and environment variables
3. Add troubleshooting guide for common issues
4. Include architecture diagram and workflow explanation
**Dependencies**: All previous tasks
**Status**: Pending

---
*Generated by /tasks command based on specification and implementation plan*