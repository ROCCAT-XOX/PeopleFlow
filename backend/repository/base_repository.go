package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"PeopleFlow/backend/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Common errors
var (
	ErrNotFound     = errors.New("document not found")
	ErrInvalidID    = errors.New("invalid object ID")
	ErrDuplicateKey = errors.New("duplicate key error")
	ErrTimeout      = errors.New("operation timeout")
	ErrValidation   = errors.New("validation error")
)

// BaseRepository provides common database operations with logging and error handling
type BaseRepository struct {
	collection *mongo.Collection
	logger     *slog.Logger
	timeout    time.Duration
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(collection *mongo.Collection) *BaseRepository {
	return &BaseRepository{
		collection: collection,
		logger:     utils.GetLogger(),
		timeout:    10 * time.Second,
	}
}

// WithContext returns a repository with context-aware logger
func (r *BaseRepository) WithContext(ctx context.Context) *BaseRepository {
	newRepo := *r
	newRepo.logger = utils.ContextLogger(ctx)
	return &newRepo
}

// SetTimeout sets the timeout for database operations
func (r *BaseRepository) SetTimeout(timeout time.Duration) {
	r.timeout = timeout
}

// GetContext returns a context with timeout
func (r *BaseRepository) GetContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), r.timeout)
}

// GetContextWithParent returns a context with timeout based on parent context
func (r *BaseRepository) GetContextWithParent(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, r.timeout)
}

// ValidateObjectID validates and converts string ID to ObjectID
func (r *BaseRepository) ValidateObjectID(id string) (*primitive.ObjectID, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: ID cannot be empty", ErrInvalidID)
	}
	
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidID, err.Error())
	}
	
	return &objID, nil
}

// HandleError converts MongoDB errors to application errors and logs them
func (r *BaseRepository) HandleError(ctx context.Context, err error, operation string) error {
	if err == nil {
		return nil
	}

	var loggedErr error
	
	switch {
	case errors.Is(err, mongo.ErrNoDocuments):
		loggedErr = ErrNotFound
		r.logger.Debug("Document not found", 
			"operation", operation,
			"collection", r.collection.Name(),
		)
	case mongo.IsDuplicateKeyError(err):
		loggedErr = ErrDuplicateKey
		r.logger.Warn("Duplicate key error",
			"operation", operation,
			"collection", r.collection.Name(),
			"error", err.Error(),
		)
	case errors.Is(err, context.DeadlineExceeded):
		loggedErr = ErrTimeout
		r.logger.Error("Database operation timeout",
			"operation", operation,
			"collection", r.collection.Name(),
			"timeout", r.timeout,
			"error", err.Error(),
		)
	default:
		loggedErr = err
		r.logger.Error("Database operation failed",
			"operation", operation,
			"collection", r.collection.Name(),
			"error", err.Error(),
		)
	}

	return loggedErr
}

// FindByID finds a document by ID
func (r *BaseRepository) FindByID(id string, result interface{}) error {
	ctx, cancel := r.GetContext()
	defer cancel()

	start := time.Now()
	defer func() {
		utils.LogRepositoryOperation(ctx, "FindByID", r.collection.Name(), time.Since(start), nil)
	}()

	objID, err := r.ValidateObjectID(id)
	if err != nil {
		return r.HandleError(ctx, err, "FindByID")
	}

	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(result)
	return r.HandleError(ctx, err, "FindByID")
}

// FindOne finds a single document by filter
func (r *BaseRepository) FindOne(filter bson.M, result interface{}, opts ...*options.FindOneOptions) error {
	ctx, cancel := r.GetContext()
	defer cancel()

	start := time.Now()
	defer func() {
		utils.LogRepositoryOperation(ctx, "FindOne", r.collection.Name(), time.Since(start), nil)
	}()

	err := r.collection.FindOne(ctx, filter, opts...).Decode(result)
	return r.HandleError(ctx, err, "FindOne")
}

// FindAll finds multiple documents by filter
func (r *BaseRepository) FindAll(filter bson.M, results interface{}, opts ...*options.FindOptions) error {
	ctx, cancel := r.GetContext()
	defer cancel()

	start := time.Now()
	defer func() {
		utils.LogRepositoryOperation(ctx, "FindAll", r.collection.Name(), time.Since(start), nil)
	}()

	cursor, err := r.collection.Find(ctx, filter, opts...)
	if err != nil {
		return r.HandleError(ctx, err, "FindAll")
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, results)
	return r.HandleError(ctx, err, "FindAll")
}

// InsertOne inserts a single document
func (r *BaseRepository) InsertOne(document interface{}) (*primitive.ObjectID, error) {
	ctx, cancel := r.GetContext()
	defer cancel()

	start := time.Now()
	defer func() {
		utils.LogRepositoryOperation(ctx, "InsertOne", r.collection.Name(), time.Since(start), nil)
	}()

	result, err := r.collection.InsertOne(ctx, document)
	if err != nil {
		return nil, r.HandleError(ctx, err, "InsertOne")
	}

	id := result.InsertedID.(primitive.ObjectID)
	r.logger.Debug("Document inserted successfully",
		"collection", r.collection.Name(),
		"id", id.Hex(),
	)

	return &id, nil
}

// InsertMany inserts multiple documents
func (r *BaseRepository) InsertMany(documents []interface{}) ([]primitive.ObjectID, error) {
	ctx, cancel := r.GetContext()
	defer cancel()

	start := time.Now()
	defer func() {
		utils.LogRepositoryOperation(ctx, "InsertMany", r.collection.Name(), time.Since(start), nil)
	}()

	result, err := r.collection.InsertMany(ctx, documents)
	if err != nil {
		return nil, r.HandleError(ctx, err, "InsertMany")
	}

	ids := make([]primitive.ObjectID, len(result.InsertedIDs))
	for i, id := range result.InsertedIDs {
		ids[i] = id.(primitive.ObjectID)
	}

	r.logger.Debug("Documents inserted successfully",
		"collection", r.collection.Name(),
		"count", len(ids),
	)

	return ids, nil
}

// UpdateByID updates a document by ID
func (r *BaseRepository) UpdateByID(id string, update bson.M) error {
	ctx, cancel := r.GetContext()
	defer cancel()

	start := time.Now()
	defer func() {
		utils.LogRepositoryOperation(ctx, "UpdateByID", r.collection.Name(), time.Since(start), nil)
	}()

	objID, err := r.ValidateObjectID(id)
	if err != nil {
		return r.HandleError(ctx, err, "UpdateByID")
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return r.HandleError(ctx, err, "UpdateByID")
	}

	if result.MatchedCount == 0 {
		return r.HandleError(ctx, ErrNotFound, "UpdateByID")
	}

	r.logger.Debug("Document updated successfully",
		"collection", r.collection.Name(),
		"id", id,
		"modified_count", result.ModifiedCount,
	)

	return nil
}

// UpdateOne updates a single document by filter
func (r *BaseRepository) UpdateOne(filter bson.M, update bson.M) (*mongo.UpdateResult, error) {
	ctx, cancel := r.GetContext()
	defer cancel()

	start := time.Now()
	defer func() {
		utils.LogRepositoryOperation(ctx, "UpdateOne", r.collection.Name(), time.Since(start), nil)
	}()

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, r.HandleError(ctx, err, "UpdateOne")
	}

	r.logger.Debug("Document updated successfully",
		"collection", r.collection.Name(),
		"matched_count", result.MatchedCount,
		"modified_count", result.ModifiedCount,
	)

	return result, nil
}

// UpdateMany updates multiple documents by filter
func (r *BaseRepository) UpdateMany(filter bson.M, update bson.M) (*mongo.UpdateResult, error) {
	ctx, cancel := r.GetContext()
	defer cancel()

	start := time.Now()
	defer func() {
		utils.LogRepositoryOperation(ctx, "UpdateMany", r.collection.Name(), time.Since(start), nil)
	}()

	result, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return nil, r.HandleError(ctx, err, "UpdateMany")
	}

	r.logger.Debug("Documents updated successfully",
		"collection", r.collection.Name(),
		"matched_count", result.MatchedCount,
		"modified_count", result.ModifiedCount,
	)

	return result, nil
}

// DeleteByID deletes a document by ID
func (r *BaseRepository) DeleteByID(id string) error {
	ctx, cancel := r.GetContext()
	defer cancel()

	start := time.Now()
	defer func() {
		utils.LogRepositoryOperation(ctx, "DeleteByID", r.collection.Name(), time.Since(start), nil)
	}()

	objID, err := r.ValidateObjectID(id)
	if err != nil {
		return r.HandleError(ctx, err, "DeleteByID")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return r.HandleError(ctx, err, "DeleteByID")
	}

	if result.DeletedCount == 0 {
		return r.HandleError(ctx, ErrNotFound, "DeleteByID")
	}

	r.logger.Debug("Document deleted successfully",
		"collection", r.collection.Name(),
		"id", id,
		"deleted_count", result.DeletedCount,
	)

	return nil
}

// DeleteOne deletes a single document by filter
func (r *BaseRepository) DeleteOne(filter bson.M) (*mongo.DeleteResult, error) {
	ctx, cancel := r.GetContext()
	defer cancel()

	start := time.Now()
	defer func() {
		utils.LogRepositoryOperation(ctx, "DeleteOne", r.collection.Name(), time.Since(start), nil)
	}()

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, r.HandleError(ctx, err, "DeleteOne")
	}

	r.logger.Debug("Document deleted successfully",
		"collection", r.collection.Name(),
		"deleted_count", result.DeletedCount,
	)

	return result, nil
}

// DeleteMany deletes multiple documents by filter
func (r *BaseRepository) DeleteMany(filter bson.M) (*mongo.DeleteResult, error) {
	ctx, cancel := r.GetContext()
	defer cancel()

	start := time.Now()
	defer func() {
		utils.LogRepositoryOperation(ctx, "DeleteMany", r.collection.Name(), time.Since(start), nil)
	}()

	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return nil, r.HandleError(ctx, err, "DeleteMany")
	}

	r.logger.Debug("Documents deleted successfully",
		"collection", r.collection.Name(),
		"deleted_count", result.DeletedCount,
	)

	return result, nil
}

// Count counts documents by filter
func (r *BaseRepository) Count(filter bson.M) (int64, error) {
	ctx, cancel := r.GetContext()
	defer cancel()

	start := time.Now()
	defer func() {
		utils.LogRepositoryOperation(ctx, "Count", r.collection.Name(), time.Since(start), nil)
	}()

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, r.HandleError(ctx, err, "Count")
	}

	r.logger.Debug("Document count completed",
		"collection", r.collection.Name(),
		"count", count,
	)

	return count, nil
}

// Exists checks if a document exists by filter
func (r *BaseRepository) Exists(filter bson.M) (bool, error) {
	count, err := r.Count(filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Aggregate performs aggregation pipeline
func (r *BaseRepository) Aggregate(pipeline []bson.M, results interface{}) error {
	ctx, cancel := r.GetContext()
	defer cancel()

	start := time.Now()
	defer func() {
		utils.LogRepositoryOperation(ctx, "Aggregate", r.collection.Name(), time.Since(start), nil)
	}()

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return r.HandleError(ctx, err, "Aggregate")
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, results)
	return r.HandleError(ctx, err, "Aggregate")
}

// CreateIndex creates an index on the collection
func (r *BaseRepository) CreateIndex(keys bson.M, unique bool) error {
	ctx, cancel := r.GetContext()
	defer cancel()

	start := time.Now()
	defer func() {
		utils.LogRepositoryOperation(ctx, "CreateIndex", r.collection.Name(), time.Since(start), nil)
	}()

	indexModel := mongo.IndexModel{
		Keys: keys,
		Options: options.Index().SetUnique(unique),
	}

	_, err := r.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return r.HandleError(ctx, err, "CreateIndex")
	}

	r.logger.Debug("Index created successfully",
		"collection", r.collection.Name(),
		"keys", keys,
		"unique", unique,
	)

	return nil
}

// Transaction executes a function within a transaction
func (r *BaseRepository) Transaction(fn func(sessCtx mongo.SessionContext) error) error {
	ctx, cancel := r.GetContext()
	defer cancel()

	start := time.Now()
	defer func() {
		utils.LogRepositoryOperation(ctx, "Transaction", r.collection.Name(), time.Since(start), nil)
	}()

	session, err := r.collection.Database().Client().StartSession()
	if err != nil {
		return r.HandleError(ctx, err, "Transaction")
	}
	defer session.EndSession(ctx)

	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		_, err := session.WithTransaction(sc, func(sc mongo.SessionContext) (interface{}, error) {
			return nil, fn(sc)
		})
		return err
	})

	if err != nil {
		r.logger.Error("Transaction failed",
			"collection", r.collection.Name(),
			"error", err.Error(),
		)
		return r.HandleError(ctx, err, "Transaction")
	}

	r.logger.Debug("Transaction completed successfully",
		"collection", r.collection.Name(),
	)

	return nil
}

// BulkWrite performs bulk operations
func (r *BaseRepository) BulkWrite(operations []mongo.WriteModel) (*mongo.BulkWriteResult, error) {
	ctx, cancel := r.GetContext()
	defer cancel()

	start := time.Now()
	defer func() {
		utils.LogRepositoryOperation(ctx, "BulkWrite", r.collection.Name(), time.Since(start), nil)
	}()

	result, err := r.collection.BulkWrite(ctx, operations)
	if err != nil {
		return nil, r.HandleError(ctx, err, "BulkWrite")
	}

	r.logger.Debug("Bulk write completed successfully",
		"collection", r.collection.Name(),
		"inserted_count", result.InsertedCount,
		"modified_count", result.ModifiedCount,
		"deleted_count", result.DeletedCount,
		"upserted_count", result.UpsertedCount,
	)

	return result, nil
}

// GetCollection returns the underlying MongoDB collection
func (r *BaseRepository) GetCollection() *mongo.Collection {
	return r.collection
}

// Ping tests the connection to the database
func (r *BaseRepository) Ping() error {
	ctx, cancel := r.GetContext()
	defer cancel()

	return r.collection.Database().Client().Ping(ctx, nil)
}