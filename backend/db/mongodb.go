package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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

	// E-Mail-Normalisierung durchführen
	if err := normalizeExistingEmails(); err != nil {
		log.Printf("Warnung: E-Mail-Normalisierung fehlgeschlagen: %v", err)
		// Fehler hier nicht zurückgeben, da die Anwendung trotzdem funktionieren soll
	}

	return nil
}

// GetCollection gibt eine Kollektion aus der Datenbank zurück
func GetCollection(collectionName string) *mongo.Collection {
	return DBClient.Database("PeopleFlow").Collection(collectionName)
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

// normalizeExistingEmails normalisiert alle E-Mail-Adressen in der Datenbank
func normalizeExistingEmails() error {
	log.Println("Prüfe E-Mail-Normalisierung...")

	// Prüfen, ob die Normalisierung bereits durchgeführt wurde
	metaCollection := GetCollection("_meta")
	ctx := context.Background()

	var meta bson.M
	err := metaCollection.FindOne(ctx, bson.M{"_id": "email_normalized"}).Decode(&meta)
	if err == nil {
		// Normalisierung wurde bereits durchgeführt
		log.Println("E-Mail-Normalisierung bereits abgeschlossen")
		return nil
	}

	log.Println("Starte E-Mail-Normalisierung...")

	// Users Collection normalisieren
	if err := normalizeEmailsInCollection(GetCollection("users")); err != nil {
		return fmt.Errorf("fehler bei Users-Normalisierung: %v", err)
	}

	// Employees Collection normalisieren
	if err := normalizeEmailsInCollection(GetCollection("employees")); err != nil {
		return fmt.Errorf("fehler bei Employees-Normalisierung: %v", err)
	}

	// Markierung setzen, dass die Normalisierung durchgeführt wurde
	_, err = metaCollection.InsertOne(ctx, bson.M{
		"_id":       "email_normalized",
		"timestamp": time.Now(),
		"version":   1,
	})
	if err != nil {
		log.Printf("Warnung: Konnte Normalisierungs-Flag nicht setzen: %v", err)
	}

	log.Println("E-Mail-Normalisierung erfolgreich abgeschlossen")
	return nil
}

// normalizeEmailsInCollection normalisiert alle E-Mail-Adressen in einer Collection
func normalizeEmailsInCollection(collection *mongo.Collection) error {
	ctx := context.Background()

	// Alle Dokumente mit E-Mail-Feld finden
	cursor, err := collection.Find(ctx, bson.M{"email": bson.M{"$exists": true}})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	updateCount := 0
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		// E-Mail extrahieren und prüfen
		email, ok := doc["email"].(string)
		if !ok || email == "" {
			continue
		}

		// E-Mail normalisieren
		normalizedEmail := strings.ToLower(strings.TrimSpace(email))

		// Nur aktualisieren, wenn sich die E-Mail geändert hat
		if normalizedEmail != email {
			_, err := collection.UpdateOne(
				ctx,
				bson.M{"_id": doc["_id"]},
				bson.M{"$set": bson.M{
					"email":     normalizedEmail,
					"updatedAt": time.Now(),
				}},
			)
			if err != nil {
				log.Printf("Fehler beim Aktualisieren von %s: %v", email, err)
			} else {
				updateCount++
			}
		}
	}

	if updateCount > 0 {
		log.Printf("%d E-Mail-Adressen in %s normalisiert", updateCount, collection.Name())
	}

	return cursor.Err()
}
