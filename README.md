# TRON Events

TRON Events is a service that scans the TRON blockchain and publishes transaction data to Redis streams for real-time processing.

## Docker Image Usage

The service is available as a Docker image published to the GitHub Container Registry.

### Pulling the Docker Image

```bash
docker pull ghcr.io/sunbankio/tronevents:latest
```

### Running the Service

```bash
docker run -d \
  --name tronevents \
  -e REDIS_ADDR=redis:6379 \
  -e REDIS_PASSWORD= \
  -e TRON_API_KEY=your_tron_api_key \
  -e START_BLOCK=0 \
  ghcr.io/sunbankio/tronevents:latest
```

### Docker Compose Example

```yaml
version: '3.8'
services:
  tronevents:
    image: ghcr.io/sunbankio/tronevents:latest
    environment:
      - REDIS_ADDR=redis:6379
      - REDIS_PASSWORD=
      - TRON_API_KEY=your_tron_api_key
      - START_BLOCK=0
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

## Example Usage

The repository includes an example Redis subscriber in `cmd/redis_subscriber/main.go` that demonstrates how to consume events from the Redis stream.

### Example Subscriber Code

The example shows how to:
- Connect to Redis
- Create a consumer group for the `tron:events` stream
- Process events with resume capability
- Store checkpoint information to resume processing after restarts

To run the example:
```bash
go run cmd/redis_subscriber/main.go
```

## Transaction Structure

When transactions are published to the Redis stream, they contain the following structure:

### Transaction
- `id` (string): Unique transaction ID
- `contract` (Contract): Contract details
- `ret` (RetInfo): Return information
- `timestamp` (time.Time): Transaction timestamp
- `block_number` (int64): Block number containing the transaction
- `block_timestamp` (time.Time): Block timestamp
- `expiration` (time.Time): Transaction expiration time
- `receipt` (Receipt): Transaction receipt information
- `logs` ([]LogInfo): Array of log events
- `signers` ([]string): All signers for the transaction

### Contract
- `type` (string): Contract type
- `parameter` (interface{}): Contract parameters
- `permission_id` (int): Permission ID

### RetInfo
- `contract_ret` (string): Contract return value

### Receipt
- `energy_usage` (int64): Energy usage
- `energy_fee` (int64): Energy fee
- `origin_energy_usage` (int64): Origin energy usage
- `energy_usage_total` (int64): Total energy usage
- `net_usage` (int64): Network usage
- `net_fee` (int64): Network fee

### LogInfo
- `event_name` (string): Name of the event
- `signature` (string): Event signature
- `inputs` ([]EventInput): Array of event inputs
- `address` (string): Contract address

### EventInput
- `name` (string): Input parameter name
- `type` (string): Input parameter type
- `value` (interface{}): Input parameter value

## Redis Streams

The service publishes TRON transaction events to Redis streams:
- Stream name: `tron:events`
- Consumer group: Your choice (example uses `exampleGroupNew`)
- Consumer name: Your choice (example uses `exampleConsumer1`)

Each message in the stream contains a JSON-encoded transaction with the structure described above.