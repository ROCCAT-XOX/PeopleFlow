package repository

import (
	"context"
	"testing"
	"time"

	"PeopleFlow/backend/db"
	"PeopleFlow/backend/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestDocument represents a test document for repository testing
type TestDocument struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Email     string             `bson:"email"`
	Age       int                `bson:"age"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

var (
	testRepo       *BaseRepository
	testCollection *mongo.Collection
)

func setupTestRepository(t *testing.T) {
	// Initialize logger for testing
	err := utils.InitLogger(utils.LoggerConfig{
		Level:  utils.LogLevelDebug,
		Format: "text",
	})
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Connect to test database
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration tests")
	}

	// Use test database
	testDB := client.Database("peopleflow_test")
	testCollection = testDB.Collection("test_documents")
	testRepo = NewBaseRepository(testCollection)

	// Clean up any existing test data
	_, err = testCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		t.Fatalf("Failed to clean test collection: %v", err)
	}
}

func teardownTestRepository(t *testing.T) {
	if testCollection != nil {
		// Clean up test data
		_, _ = testCollection.DeleteMany(context.Background(), bson.M{})
		_ = testCollection.Drop(context.Background())
	}
}

func TestNewBaseRepository(t *testing.T) {
	setupTestRepository(t)
	defer teardownTestRepository(t)

	if testRepo == nil {
		t.Fatal("BaseRepository should not be nil")
	}

	if testRepo.collection != testCollection {
		t.Error("BaseRepository collection not set correctly")
	}

	if testRepo.timeout != 10*time.Second {
		t.Errorf("Expected timeout to be 10s, got %v", testRepo.timeout)
	}
}

func TestBaseRepository_SetTimeout(t *testing.T) {
	setupTestRepository(t)
	defer teardownTestRepository(t)

	newTimeout := 5 * time.Second
	testRepo.SetTimeout(newTimeout)

	if testRepo.timeout != newTimeout {
		t.Errorf("Expected timeout to be %v, got %v", newTimeout, testRepo.timeout)
	}
}

func TestBaseRepository_ValidateObjectID(t *testing.T) {
	setupTestRepository(t)
	defer teardownTestRepository(t)

	tests := []struct {
		name        string
		id          string
		shouldError bool
	}{
		{
			name:        "valid ObjectID",
			id:          primitive.NewObjectID().Hex(),
			shouldError: false,
		},
		{
			name:        "empty ID",
			id:          "",
			shouldError: true,
		},
		{
			name:        "invalid ObjectID",
			id:          "invalid",
			shouldError: true,
		},
		{
			name:        "invalid hex string",
			id:          "zzzzzzzzzzzzzzzzzzzzzzzz",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objID, err := testRepo.ValidateObjectID(tt.id)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if objID != nil {
					t.Error("Expected nil ObjectID when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if objID == nil {
					t.Error("Expected valid ObjectID but got nil")
				}
				if objID.Hex() != tt.id {
					t.Errorf("Expected ObjectID %s, got %s", tt.id, objID.Hex())
				}
			}
		})
	}
}

func TestBaseRepository_InsertOne(t *testing.T) {
	setupTestRepository(t)
	defer teardownTestRepository(t)

	doc := TestDocument{
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := testRepo.InsertOne(doc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	if id == nil {
		t.Fatal("Expected ID to be returned")
	}

	if id.IsZero() {
		t.Error("Expected valid ObjectID")
	}
}

func TestBaseRepository_FindByID(t *testing.T) {
	setupTestRepository(t)
	defer teardownTestRepository(t)

	// Insert test document
	doc := TestDocument{
		Name:      "Jane Doe",
		Email:     "jane@example.com",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := testRepo.InsertOne(doc)
	if err != nil {
		t.Fatalf("Failed to insert test document: %v", err)
	}

	// Test finding by ID
	var result TestDocument
	err = testRepo.FindByID(id.Hex(), &result)
	if err != nil {
		t.Fatalf("Failed to find document by ID: %v", err)
	}

	if result.Name != doc.Name {
		t.Errorf("Expected name %s, got %s", doc.Name, result.Name)
	}
	if result.Email != doc.Email {
		t.Errorf("Expected email %s, got %s", doc.Email, result.Email)
	}
	if result.Age != doc.Age {
		t.Errorf("Expected age %d, got %d", doc.Age, result.Age)
	}

	// Test finding non-existent document
	nonExistentID := primitive.NewObjectID()
	var nonExistentResult TestDocument
	err = testRepo.FindByID(nonExistentID.Hex(), &nonExistentResult)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestBaseRepository_FindOne(t *testing.T) {
	setupTestRepository(t)
	defer teardownTestRepository(t)

	// Insert test documents
	docs := []TestDocument{
		{Name: "Alice", Email: "alice@example.com", Age: 28},
		{Name: "Bob", Email: "bob@example.com", Age: 32},
	}

	for _, doc := range docs {
		_, err := testRepo.InsertOne(doc)
		if err != nil {
			t.Fatalf("Failed to insert test document: %v", err)
		}
	}

	// Test finding by filter
	var result TestDocument
	filter := bson.M{"name": "Alice"}
	err := testRepo.FindOne(filter, &result)
	if err != nil {
		t.Fatalf("Failed to find document: %v", err)
	}

	if result.Name != "Alice" {
		t.Errorf("Expected name Alice, got %s", result.Name)
	}

	// Test finding non-existent document
	var nonExistentResult TestDocument
	filter = bson.M{"name": "Charlie"}
	err = testRepo.FindOne(filter, &nonExistentResult)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestBaseRepository_FindAll(t *testing.T) {
	setupTestRepository(t)
	defer teardownTestRepository(t)

	// Insert test documents
	docs := []TestDocument{
		{Name: "Alice", Email: "alice@example.com", Age: 28},
		{Name: "Bob", Email: "bob@example.com", Age: 32},
		{Name: "Charlie", Email: "charlie@example.com", Age: 25},
	}

	for _, doc := range docs {
		_, err := testRepo.InsertOne(doc)
		if err != nil {
			t.Fatalf("Failed to insert test document: %v", err)
		}
	}

	// Test finding all documents
	var results []TestDocument
	err := testRepo.FindAll(bson.M{}, &results)
	if err != nil {
		t.Fatalf("Failed to find documents: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(results))
	}

	// Test finding with filter
	var filteredResults []TestDocument
	filter := bson.M{"age": bson.M{"$gte": 30}}
	err = testRepo.FindAll(filter, &filteredResults)
	if err != nil {
		t.Fatalf("Failed to find filtered documents: %v", err)
	}

	if len(filteredResults) != 1 {
		t.Errorf("Expected 1 document, got %d", len(filteredResults))
	}

	if filteredResults[0].Name != "Bob" {
		t.Errorf("Expected Bob, got %s", filteredResults[0].Name)
	}
}

func TestBaseRepository_UpdateByID(t *testing.T) {
	setupTestRepository(t)
	defer teardownTestRepository(t)

	// Insert test document
	doc := TestDocument{
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := testRepo.InsertOne(doc)
	if err != nil {
		t.Fatalf("Failed to insert test document: %v", err)
	}

	// Update document
	update := bson.M{
		"$set": bson.M{
			"name": "John Smith",
			"age":  31,
		},
	}

	err = testRepo.UpdateByID(id.Hex(), update)
	if err != nil {
		t.Fatalf("Failed to update document: %v", err)
	}

	// Verify update
	var result TestDocument
	err = testRepo.FindByID(id.Hex(), &result)
	if err != nil {
		t.Fatalf("Failed to find updated document: %v", err)
	}

	if result.Name != "John Smith" {
		t.Errorf("Expected name John Smith, got %s", result.Name)
	}
	if result.Age != 31 {
		t.Errorf("Expected age 31, got %d", result.Age)
	}

	// Test updating non-existent document
	nonExistentID := primitive.NewObjectID()
	err = testRepo.UpdateByID(nonExistentID.Hex(), update)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestBaseRepository_DeleteByID(t *testing.T) {
	setupTestRepository(t)
	defer teardownTestRepository(t)

	// Insert test document
	doc := TestDocument{
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := testRepo.InsertOne(doc)
	if err != nil {
		t.Fatalf("Failed to insert test document: %v", err)
	}

	// Delete document
	err = testRepo.DeleteByID(id.Hex())
	if err != nil {
		t.Fatalf("Failed to delete document: %v", err)
	}

	// Verify deletion
	var result TestDocument
	err = testRepo.FindByID(id.Hex(), &result)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound after deletion, got %v", err)
	}

	// Test deleting non-existent document
	nonExistentID := primitive.NewObjectID()
	err = testRepo.DeleteByID(nonExistentID.Hex())
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestBaseRepository_Count(t *testing.T) {
	setupTestRepository(t)
	defer teardownTestRepository(t)

	// Insert test documents
	docs := []TestDocument{
		{Name: "Alice", Email: "alice@example.com", Age: 28},
		{Name: "Bob", Email: "bob@example.com", Age: 32},
		{Name: "Charlie", Email: "charlie@example.com", Age: 25},
	}

	for _, doc := range docs {
		_, err := testRepo.InsertOne(doc)
		if err != nil {
			t.Fatalf("Failed to insert test document: %v", err)
		}
	}

	// Test counting all documents
	count, err := testRepo.Count(bson.M{})
	if err != nil {
		t.Fatalf("Failed to count documents: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	// Test counting with filter
	filter := bson.M{"age": bson.M{"$gte": 30}}
	count, err = testRepo.Count(filter)
	if err != nil {
		t.Fatalf("Failed to count filtered documents: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestBaseRepository_Exists(t *testing.T) {
	setupTestRepository(t)
	defer teardownTestRepository(t)

	// Insert test document
	doc := TestDocument{
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := testRepo.InsertOne(doc)
	if err != nil {
		t.Fatalf("Failed to insert test document: %v", err)
	}

	// Test existing document
	exists, err := testRepo.Exists(bson.M{"name": "John Doe"})
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}

	if !exists {
		t.Error("Expected document to exist")
	}

	// Test non-existing document
	exists, err = testRepo.Exists(bson.M{"name": "Jane Doe"})
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}

	if exists {
		t.Error("Expected document not to exist")
	}
}

func TestBaseRepository_CreateIndex(t *testing.T) {
	setupTestRepository(t)
	defer teardownTestRepository(t)

	// Create unique index on email
	indexKeys := bson.M{"email": 1}
	err := testRepo.CreateIndex(indexKeys, true)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// Insert document
	doc := TestDocument{
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = testRepo.InsertOne(doc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	// Try to insert duplicate email (should fail due to unique index)
	duplicateDoc := TestDocument{
		Name:      "Jane Doe",
		Email:     "john@example.com", // Same email
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = testRepo.InsertOne(duplicateDoc)
	if err != ErrDuplicateKey {
		t.Errorf("Expected ErrDuplicateKey due to unique index, got %v", err)
	}
}

func TestBaseRepository_WithContext(t *testing.T) {
	setupTestRepository(t)
	defer teardownTestRepository(t)

	ctx := context.WithValue(context.Background(), "requestID", "test-123")
	repoWithContext := testRepo.WithContext(ctx)

	if repoWithContext == testRepo {
		t.Error("WithContext should return a new repository instance")
	}

	if repoWithContext.collection != testRepo.collection {
		t.Error("WithContext should preserve the collection")
	}
}

// Benchmark tests
func BenchmarkBaseRepository_InsertOne(b *testing.B) {
	setupTestRepository(&testing.T{})
	defer teardownTestRepository(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc := TestDocument{
			Name:      "Benchmark User",
			Email:     "benchmark@example.com",
			Age:       30,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_, _ = testRepo.InsertOne(doc)
	}
}

func BenchmarkBaseRepository_FindByID(b *testing.B) {
	setupTestRepository(&testing.T{})
	defer teardownTestRepository(&testing.T{})

	// Insert test document
	doc := TestDocument{
		Name:      "Benchmark User",
		Email:     "benchmark@example.com",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, _ := testRepo.InsertOne(doc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result TestDocument
		_ = testRepo.FindByID(id.Hex(), &result)
	}
}