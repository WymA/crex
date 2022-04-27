package dataloader

import (
	"time"

	. "github.com/WymA/crex"
)

type DataLoader interface {
	Setup(start time.Time, end time.Time) error
	ReadOrderBooks() []*OrderBook
	ReadRecords(limit int) []*Record
	HasMoreData() bool
}
