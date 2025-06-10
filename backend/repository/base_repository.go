// backend/repository/base_repository.go
package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Common errors
var (
	ErrInvalidID      = errors.New("invalid object ID")
	ErrNotFound       = errors.New("document not found")
	ErrInvalidInput   = errors.New("invalid input data")
	ErrDuplicateEntry = errors.New("duplicate entry")
	ErrNilDocument    = errors.New("document is nil")
	ErrNilFilter      = errors.New("filter is nil")
)

// BaseRepository provides common functionality for all repositories
type BaseRepository struct {
	collection *mongo.Collection
	timeout    time.Duration
	logger     *slog.Logger
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(collection *mongo.Collection) *BaseRepository {
	return &BaseRepository{
		collection: collection,
		timeout:    10 * time.Second,
		logger:     slog.Default().With("component", "repository", "collection", collection.Name()),
	}
}

// WithLogger sets a custom logger
func (r *BaseRepository) WithLogger(logger *slog.Logger) *BaseRepository {
	r.logger = logger.With("component", "repository", "collection", r.collection.Name())
	return r
}

// GetContext returns a context with timeout
func (r *BaseRepository) GetContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), r.timeout)
}

// GetContextWithTransaction returns a context with timeout for transactions
func (r *BaseRepository) GetContextWithTransaction() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), r.timeout*2) // Longer timeout for transactions
}

// ValidateObjectID validates and converts a string to ObjectID
func (r *BaseRepository) ValidateObjectID(id string) (primitive.ObjectID, error) {
	if id == "" {
		r.logger.Error("ValidateObjectID: empty ID provided")
		return primitive.NilObjectID, fmt.Errorf("%w: ID cannot be empty", ErrInvalidID)
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		r.logger.Error("ValidateObjectID: invalid ID format", "id", id, "error", err)
		return primitive.NilObjectID, fmt.Errorf("%w: %v", ErrInvalidID, err)
	}

	return objID, nil
}

// FindByID finds a document by ID
func (r *BaseRepository) FindByID(id string, result interface{}) error {
	if result == nil {
		return ErrNilDocument
	}

	ctx, cancel := r.GetContext()
	defer cancel()

	objID, err := r.ValidateObjectID(id)
	if err != nil {
		return err
	}

	r.logger.Debug("FindByID", "id", id)

	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			r.logger.Debug("FindByID: document not found", "id", id)
			return fmt.Errorf("%w: ID %s", ErrNotFound, id)
		}
		r.logger.Error("FindByID: database error", "id", id, "error", err)
		return fmt.Errorf("failed to find document: %w", err)
	}

	r.logger.Debug("FindByID: document found", "id", id)
	return nil
}

// FindOne finds a single document matching the filter
func (r *BaseRepository) FindOne(filter interface{}, result interface{}, opts ...*options.FindOneOptions) error {
	if filter == nil {
		return ErrNilFilter
	}
	if result == nil {
		return ErrNilDocument
	}

	ctx, cancel := r.GetContext()
	defer cancel()

	r.logger.Debug("FindOne", "filter", filter)

	err := r.collection.FindOne(ctx, filter, opts...).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			r.logger.Debug("FindOne: document not found", "filter", filter)
			return ErrNotFound
		}
		r.logger.Error("FindOne: database error", "filter", filter, "error", err)
		return fmt.Errorf("failed to find document: %w", err)
	}

	r.logger.Debug("FindOne: document found", "filter", filter)
	return nil
}

// FindAll finds all documents matching the filter
func (r *BaseRepository) FindAll(filter interface{}, results interface{}, opts ...*options.FindOptions) error {
	if filter == nil {
		return ErrNilFilter
	}
	if results == nil {
		return ErrNilDocument
	}

	ctx, cancel := r.GetContext()
	defer cancel()

	r.logger.Debug("FindAll", "filter", filter)

	cursor, err := r.collection.Find(ctx, filter, opts...)
	if err != nil {
		r.logger.Error("FindAll: database error", "filter", filter, "error", err)
		return fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, results); err != nil {
		r.logger.Error("FindAll: decode error", "filter", filter, "error", err)
		return fmt.Errorf("failed to decode documents: %w", err)
	}

	r.logger.Debug("FindAll: documents found", "filter", filter)
	return nil
}

// InsertOne inserts a single document
func (r *BaseRepository) InsertOne(document interface{}) (*primitive.ObjectID, error) {
	if document == nil {
		return nil, ErrNilDocument
	}

	ctx, cancel := r.GetContext()
	defer cancel()

	r.logger.Debug("InsertOne", "document_type", fmt.Sprintf("%T", document))

	result, err := r.collection.InsertOne(ctx, document)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			r.logger.Info("InsertOne: duplicate key error", "error", err)
			return nil, ErrDuplicateEntry
		}
		r.logger.Error("InsertOne: database error", "error", err)
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}

	id := result.InsertedID.(primitive.ObjectID)
	r.logger.Info("InsertOne: document created", "id", id.Hex())
	return &id, nil
}

// UpdateOne updates a single document
func (r *BaseRepository) UpdateOne(filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	if filter == nil {
		return nil, ErrNilFilter
	}
	if update == nil {
		return nil, ErrNilDocument
	}

	ctx, cancel := r.GetContext()
	defer cancel()

	r.logger.Debug("UpdateOne", "filter", filter, "update", update)

	result, err := r.collection.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			r.logger.Info("UpdateOne: duplicate key error", "filter", filter, "error", err)
			return nil, ErrDuplicateEntry
		}
		r.logger.Error("UpdateOne: database error", "filter", filter, "error", err)
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	if result.MatchedCount == 0 {
		r.logger.Debug("UpdateOne: no document matched", "filter", filter)
		return nil, ErrNotFound
	}

	r.logger.Info("UpdateOne: document updated",
		"matched", result.MatchedCount,
		"modified", result.ModifiedCount)
	return result, nil
}

// UpdateByID updates a document by ID
func (r *BaseRepository) UpdateByID(id string, update interface{}) error {
	objID, err := r.ValidateObjectID(id)
	if err != nil {
		return err
	}

	result, err := r.UpdateOne(bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("%w: ID %s", ErrNotFound, id)
	}

	return nil
}

// DeleteOne deletes a single document
func (r *BaseRepository) DeleteOne(filter interface{}) error {
	if filter == nil {
		return ErrNilFilter
	}

	ctx, cancel := r.GetContext()
	defer cancel()

	r.logger.Debug("DeleteOne", "filter", filter)

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		r.logger.Error("DeleteOne: database error", "filter", filter, "error", err)
		return fmt.Errorf("failed to delete document: %w", err)
	}

	if result.DeletedCount == 0 {
		r.logger.Debug("DeleteOne: no document matched", "filter", filter)
		return ErrNotFound
	}

	r.logger.Info("DeleteOne: document deleted", "filter", filter)
	return nil
}

// DeleteByID deletes a document by ID
func (r *BaseRepository) DeleteByID(id string) error {
	objID, err := r.ValidateObjectID(id)
	if err != nil {
		return err
	}

	return r.DeleteOne(bson.M{"_id": objID})
}

// Count counts documents matching the filter
func (r *BaseRepository) Count(filter interface{}) (int64, error) {
	if filter == nil {
		filter = bson.M{}
	}

	ctx, cancel := r.GetContext()
	defer cancel()

	r.logger.Debug("Count", "filter", filter)

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error("Count: database error", "filter", filter, "error", err)
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	r.logger.Debug("Count: result", "filter", filter, "count", count)
	return count, nil
}

// Exists checks if a document exists
func (r *BaseRepository) Exists(filter interface{}) (bool, error) {
	count, err := r.Count(filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Transaction executes a function within a transaction
func (r *BaseRepository) Transaction(fn func(sessCtx mongo.SessionContext) error) error {
	client := r.collection.Database().Client()

	ctx, cancel := r.GetContextWithTransaction()
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		r.logger.Error("Transaction: failed to start session", "error", err)
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	r.logger.Debug("Transaction: starting")

	err = mongo.WithSession(ctx, session, func(sessCtx mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			r.logger.Error("Transaction: failed to start transaction", "error", err)
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		if err := fn(sessCtx); err != nil {
			r.logger.Error("Transaction: operation failed, aborting", "error", err)
			if abortErr := session.AbortTransaction(sessCtx); abortErr != nil {
				r.logger.Error("Transaction: failed to abort", "abortError", abortErr, "originalError", err)
				return fmt.Errorf("failed to abort transaction: %v (original error: %w)", abortErr, err)
			}
			return err
		}

		if err := session.CommitTransaction(sessCtx); err != nil {
			r.logger.Error("Transaction: failed to commit", "error", err)
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		r.logger.Debug("Transaction: committed successfully")
		return nil
	})

	return err
}

// CreateIndex creates an index on the collection
func (r *BaseRepository) CreateIndex(keys interface{}, unique bool) error {
	ctx, cancel := r.GetContext()
	defer cancel()

	r.logger.Debug("CreateIndex", "keys", keys, "unique", unique)

	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetUnique(unique),
	}

	indexName, err := r.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		r.logger.Error("CreateIndex: failed", "keys", keys, "unique", unique, "error", err)
		return fmt.Errorf("failed to create index: %w", err)
	}

	r.logger.Info("CreateIndex: created", "name", indexName, "keys", keys, "unique", unique)
	return nil
}

// BulkWrite performs bulk write operations
func (r *BaseRepository) BulkWrite(models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	if len(models) == 0 {
		return nil, errors.New("no write models provided")
	}

	ctx, cancel := r.GetContext()
	defer cancel()

	r.logger.Debug("BulkWrite", "operations", len(models))

	result, err := r.collection.BulkWrite(ctx, models, opts...)
	if err != nil {
		r.logger.Error("BulkWrite: failed", "operations", len(models), "error", err)
		return nil, fmt.Errorf("failed to perform bulk write: %w", err)
	}

	r.logger.Info("BulkWrite: completed",
		"inserted", result.InsertedCount,
		"modified", result.ModifiedCount,
		"deleted", result.DeletedCount)

	return result, nil
}

// GetCollection returns the underlying MongoDB collection
func (r *BaseRepository) GetCollection() *mongo.Collection {
	return r.collection
}
