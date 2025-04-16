package repository

import (
	"context"
	"time"

	"PeoplePilot/backend/db"
	"PeoplePilot/backend/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserRepository enthält alle Datenbankoperationen für das User-Modell
type UserRepository struct {
	collection *mongo.Collection
}

// NewUserRepository erstellt ein neues UserRepository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		collection: db.GetCollection("users"),
	}
}

// Create erstellt einen neuen Benutzer
func (r *UserRepository) Create(user *model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Passwort hashen
	if err := user.HashPassword(); err != nil {
		return err
	}

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByID findet einen Benutzer anhand seiner ID
func (r *UserRepository) FindByID(id string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user model.User
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// FindByEmail findet einen Benutzer anhand seiner E-Mail
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// FindAll findet alle Benutzer
func (r *UserRepository) FindAll() ([]*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var users []*model.User
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Update aktualisiert einen Benutzer
func (r *UserRepository) Update(user *model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user.UpdatedAt = time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": user},
	)
	return err
}

// Delete löscht einen Benutzer
func (r *UserRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// CreateAdminUserIfNotExists erstellt einen Admin-Benutzer, falls keiner existiert
func (r *UserRepository) CreateAdminUserIfNotExists() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prüfen, ob bereits ein Admin-Benutzer existiert
	count, err := r.collection.CountDocuments(ctx, bson.M{"role": model.RoleAdmin})
	if err != nil {
		return err
	}

	// Wenn bereits ein Admin existiert, nichts tun
	if count > 0 {
		return nil
	}

	// Admin-Benutzer erstellen
	admin := &model.User{
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@peoplepilot.com",
		Password:  "admin",
		Role:      model.RoleAdmin,
		Status:    model.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Passwort hashen
	if err := admin.HashPassword(); err != nil {
		return err
	}

	// Admin in der Datenbank speichern
	_, err = r.collection.InsertOne(ctx, admin)
	return err
}
