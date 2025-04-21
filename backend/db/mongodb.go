package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB Connection-URI
var mongoURI = "mongodb://localhost:27017" // Default-Wert

// DBClient ist der shared MongoDB-Client
var DBClient *mongo.Client

// ConnectDB stellt eine Verbindung zur MongoDB her
func ConnectDB() error {
	// Umgebungsvariable prüfen und verwenden, falls vorhanden
	if uri := os.Getenv("MONGODB_URI"); uri != "" {
		mongoURI = uri
	}

	log.Printf("Verbinde zu MongoDB: %s", mongoURI)

	// Verbindungskontext mit Timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verbindung zur MongoDB herstellen
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Printf("Fehler beim Verbinden zur MongoDB: %v", err)
		return err
	}

	// Ping zur Überprüfung der Verbindung
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Printf("Fehler beim Pingen der MongoDB: %v", err)
		return err
	}

	DBClient = client
	log.Println("Erfolgreich mit MongoDB verbunden")
	return nil
}

// GetCollection gibt eine Kollektion aus der Datenbank zurück
func GetCollection(collectionName string) *mongo.Collection {
	return DBClient.Database("peoplepilot").Collection(collectionName)
}

// DisconnectDB trennt die Verbindung zur MongoDB
func DisconnectDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := DBClient.Disconnect(ctx); err != nil {
		log.Printf("Fehler beim Trennen der MongoDB-Verbindung: %v", err)
		return err
	}

	log.Println("Verbindung zur MongoDB getrennt")
	return nil
}
