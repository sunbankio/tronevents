package worker

// BlockProcessPayload defines the payload for the block:process task.
type BlockProcessPayload struct {
	BlockNumber int64 `json:"block_number"`
}