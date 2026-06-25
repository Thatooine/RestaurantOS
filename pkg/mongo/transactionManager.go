package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// TransactionManager runs a callback within a single MongoDB transaction. The
// callback receives a session-scoped context that must be passed to every mongo
// operation that should participate in the transaction.
type TransactionManager interface {
	RunInTransaction(ctx context.Context, fn func(sessionCtx context.Context) error) error
}

// transactionManager is the *mongo.Client-backed implementation of TransactionManager.
type transactionManager struct {
	client *mongo.Client
}

// NewTransactionManager returns a TransactionManager backed by the given client.
func NewTransactionManager(client *mongo.Client) TransactionManager {
	return &transactionManager{client: client}
}

// RunInTransaction starts a session, runs fn inside a transaction, and commits
// it if fn returns nil. If fn returns an error the transaction is aborted and
// the error is returned. Transient errors are retried by the driver.
func (t *transactionManager) RunInTransaction(ctx context.Context, fn func(sessionCtx context.Context) error) error {
	session, err := t.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start mongo session: %w", err)
	}
	defer session.EndSession(ctx)

	if _, err := session.WithTransaction(ctx, func(sessionCtx context.Context) (any, error) {
		return nil, fn(sessionCtx)
	}); err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}
