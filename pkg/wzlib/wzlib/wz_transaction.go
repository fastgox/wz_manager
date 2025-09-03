package wzlib

import (
	"errors"
	"io"
	"sync"
)

// Transaction represents a single transactional operation.
type Transaction struct {
	reader    *WzBinaryReader
	isActive  bool
	lock      sync.Mutex
	backupPos int64 // Backup of the stream position for rollback
}

// BeginTransaction starts a new transaction.
func BeginTransaction(reader *WzBinaryReader) *Transaction {
	return &Transaction{
		reader:   reader,
		isActive: true,
	}
}

// Commit ends the transaction and confirms all operations.
func (t *Transaction) Commit() error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if !t.isActive {
		return errors.New("transaction is not active")
	}

	t.isActive = false
	return nil
}

// Rollback reverts all changes made during the transaction.
func (t *Transaction) Rollback() error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if !t.isActive {
		return errors.New("transaction is not active")
	}

	// Revert the position of the stream
	_, err := t.reader.BaseStream.Seek(t.backupPos, io.SeekStart)
	t.isActive = false
	return err
}

// Do performs an operation within the transaction.
func (t *Transaction) Do(action func(*WzBinaryReader) error) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if !t.isActive {
		return errors.New("transaction is not active")
	}

	// Save the initial position for rollback
	if t.backupPos == 0 {
		pos, err := t.reader.BaseStream.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		t.backupPos = pos
	}

	// Execute the transactional action
	return action(t.reader)
}
