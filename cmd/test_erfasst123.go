package main

import (
	"PeopleFlow/backend/db"
	"PeopleFlow/backend/service"
	"fmt"
	"log"
	"time"
)

func main() {
	fmt.Println("==============================================")
	fmt.Println("123erfasst Synchronisation Test")
	fmt.Println("==============================================")

	// Datenbankverbindung herstellen
	if err := db.ConnectDB(); err != nil {
		log.Fatal("❌ Fehler beim Verbinden zur Datenbank:", err)
	}
	defer db.DisconnectDB()

	// Service erstellen
	erfasst123Service := service.NewErfasst123Service()

	// Prüfe ob 123erfasst verbunden ist
	if !erfasst123Service.IsConnected() {
		fmt.Println("❌ 123erfasst ist nicht verbunden!")
		return
	}

	// Teste verschiedene Synchronisationen
	tests := []struct {
		name  string
		start string
		end   string
	}{
		{"Heute", time.Now().Format("2006-01-02"), time.Now().Format("2006-01-02")},
		{"Diese Woche", time.Now().AddDate(0, 0, -7).Format("2006-01-02"), time.Now().Format("2006-01-02")},
		{"Juni 2025", "2025-06-01", "2025-06-08"},
		{"Mai 2025", "2025-05-01", "2025-05-31"},
	}

	for _, test := range tests {
		fmt.Printf("\n=== Synchronisiere %s (%s bis %s) ===\n", test.name, test.start, test.end)

		// Erst prüfen, ob es Daten gibt
		timeEntries, err := erfasst123Service.GetTimeEntries(test.start, test.end)
		if err != nil {
			fmt.Printf("❌ Fehler beim Abrufen: %v\n", err)
			continue
		}

		fmt.Printf("📊 Gefunden: %d Zeiteinträge\n", len(timeEntries))

		if len(timeEntries) == 0 {
			fmt.Println("⏭️  Überspringe Synchronisation - keine Daten")
			continue
		}

		// Synchronisation durchführen
		fmt.Println("🔄 Starte Synchronisation...")
		updatedCount, err := erfasst123Service.SyncErfasst123TimeEntries(test.start, test.end)

		if err != nil {
			fmt.Printf("❌ Fehler bei Synchronisation: %v\n", err)
		} else {
			fmt.Printf("✅ Erfolgreich! %d Mitarbeiter aktualisiert\n", updatedCount)
		}
	}

	fmt.Println("\n==============================================")
	fmt.Println("Test abgeschlossen")
	fmt.Println("==============================================")
}
