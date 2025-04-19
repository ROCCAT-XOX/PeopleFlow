# PeoplePilot - HR Management System

```
PeoplePilot/
├── backend/
│   ├── db/
│   ├── handler/
│   ├── model/
│   ├── router.go
│   │
├── frontend/
│   ├── static/
│   ├── templates/
│   │   ├── components/
│   └── assets/
│
├── go.mod
└── go.sum
```

## Übersicht

PeoplePilot ist ein modulares HR-Management-System, entwickelt mit Go, Gin für das Backend und Tailwind CSS für das Frontend. Das System ermöglicht die effiziente Verwaltung von Mitarbeiterdaten, Dokumenten, Urlaub und mehr.

## Kernfunktionen

- **Mitarbeiterverwaltung**: Umfassende Verwaltung von Mitarbeiterdaten
- **Hierarchische Ansicht**: Darstellung der Unternehmensstruktur
- **Dokumentenmanagement**: Speicherung und Verwaltung von Mitarbeiterdokumenten
- **Lohn- und Gehaltsverwaltung**: Berechnung und Verwaltung von Vergütungen
- **Statistiken & Reporting**: Aggregation von Daten zu Fehlzeiten, Ausgaben, Urlaub
- **Urlaubs- und Abwesenheitsverwaltung**: Planung und Genehmigung von Abwesenheiten
- **Zertifikate und Fortbildungen**: Verwaltung von Weiterbildungsmaßnahmen
- **Mitarbeitergespräch-Tracking**: Erfassung und Dokumentation von Feedback und Gesprächen

## Aktivitätsverfolgung

Das System protokolliert automatisch wichtige Aktivitäten wie:

- Hinzufügen, Aktualisieren und Löschen von Mitarbeitern
- Urlaubsanträge und deren Genehmigung/Ablehnung
- Dokumenten-Uploads
- Hinzufügen von Weiterbildungen und Leistungsbeurteilungen

Für jede Aktivität werden folgende Informationen erfasst:
- Aktivitätstyp (z.B. Mitarbeiter hinzugefügt, Dokument hochgeladen)
- Ausführender Benutzer (ID und Name)
- Betroffenes Objekt (ID, Typ und Name, z.B. ein Mitarbeiter)
- Zeitstempel der Aktivität
- Beschreibung der Aktivität
- Visuelle Indikatoren (farbliche Kennzeichnung und passende Icons)

Diese Aktivitäten werden auf dem Dashboard im Bereich "Letzte Aktivitäten" angezeigt, wodurch Benutzer einen schnellen Überblick über aktuelle Änderungen im System erhalten. Die Anzeige ist chronologisch sortiert und zeigt die relevantesten Informationen für jede Aktion sowie benutzerfreundliche Zeitangaben wie "vor 5 Minuten" oder "gestern".

## Technologie-Stack

- **Backend**: Go mit Gin Framework
- **Frontend**: HTML, Tailwind CSS, JavaScript
- **Datenbank**: MongoDB
- **Authentifizierung**: JWT-basiert

## Installation

1. Go 1.23.4 oder höher installieren
2. MongoDB installieren und starten
3. Repository klonen
4. Abhängigkeiten installieren: `go mod download`
5. Server starten: `go run main.go`

Die Anwendung ist dann unter http://localhost:8080 erreichbar.

## Entwicklungsumgebung

Für die Entwicklung empfiehlt sich die Verwendung von Air für automatisches Neuladen bei Änderungen:

```bash
air
```

## Benutzeranmeldung

Die Anwendung erstellt standardmäßig einen Admin-Benutzer:
- E-Mail: admin@peoplepilot.com
- Passwort: admin