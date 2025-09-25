package models

import "github.com/sunbankio/tronevents/pkg/scanner"

// Event is an adapter for the scanner.Transaction struct,
// optimized for Redis Stream publishing.
type Event scanner.Transaction
