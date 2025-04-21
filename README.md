# PeopleFlow - HR Management System

```
PeopleFlow/
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

PeopleFlow ist ein modulares HR-Management-System, entwickelt mit Go, Gin für das Backend und Tailwind CSS für das Frontend. Das System ermöglicht die effiziente Verwaltung von Mitarbeiterdaten, Dokumenten, Urlaub und mehr.

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
- E-Mail: admin@PeopleFlow.com
- Passwort: admin

# Implementierungsleitfaden für rollenbasierte Benutzerverwaltung in PeopleFlow

## Überblick

Die Implementierung einer rollenbasierten Benutzerverwaltung für PeopleFlow ermöglicht eine differenzierte Zugriffssteuerung mit folgenden Benutzerrollen:

1. **Benutzer (user)**: Kann nur die eigenen Daten sehen und verwalten
2. **Personalverwaltung (hr)**: Kann Mitarbeiter und Dokumente verwalten
3. **Manager (manager)**: Kann Mitarbeiter, Dokumente und Berichte verwalten
4. **Administrator (admin)**: Hat vollen Zugriff auf alle Funktionen, einschließlich Benutzerverwaltung

## Implementierte Komponenten

### 1. Middleware für rollenbasierte Zugriffskontrolle

Die neue Middleware `RoleMiddleware` und `SelfOrAdminMiddleware` in `backend/middleware/roleMiddleware.go` überprüft die Benutzerrollen und beschränkt den Zugriff auf bestimmte Funktionen.

### 2. Benutzer-Handler

Der neue `UserHandler` in `backend/handler/userHandler.go` enthält Funktionen für die Benutzerverwaltung:
- Anzeigen aller Benutzer (nur für Admins)
- Hinzufügen, Bearbeiten und Löschen von Benutzern
- Benutzerprofilansicht
- Passwortänderung

### 3. Aktualisierte Aktivitätstypen

Das Modell `model/activity.go` wurde um neue Aktivitätstypen für Benutzeraktionen erweitert.

### 4. Neue Templates

Folgende Templates wurden erstellt:
- `users.html`: Übersicht aller Benutzer (nur für Admins)
- `user_add.html`: Formular zum Hinzufügen eines Benutzers
- `user_edit.html`: Formular zum Bearbeiten eines Benutzers
- `profile.html`: Anzeige und Bearbeitung des eigenen Profils

### 5. Rollenbasierte Navigation

Die Navigationsleiste wurde so angepasst, dass Menüpunkte basierend auf der Benutzerrolle ein- oder ausgeblendet werden.

### 6. Rollenbasiertes Dashboard

Das Dashboard zeigt je nach Benutzerrolle unterschiedliche Inhalte an.

## Installationsschritte

1. **Dateien erstellen/aktualisieren:**
    - Neue Middleware-Datei: `backend/middleware/roleMiddleware.go`
    - Neuer Handler: `backend/handler/userHandler.go`
    - Neue Templates im Verzeichnis `frontend/templates/`
    - Aktualisierung der Navigation: `frontend/templates/components/navigation.html`

2. **Router-Konfiguration aktualisieren:**
    - `backend/router.go`: Neue Routen für die Benutzerverwaltung hinzufügen

3. **Model erweitern:**
    - `model/activity.go`: Neue Aktivitätstypen für Benutzeraktionen hinzufügen

## Benutzerverwaltung für Administratoren

Administratoren haben Zugriff auf eine spezielle Benutzerverwaltungsseite unter `/users`, auf der sie:
- Eine Liste aller Benutzer einsehen können
- Neue Benutzer hinzufügen können
- Bestehende Benutzer bearbeiten können
- Benutzer löschen können
- Die Rolle eines Benutzers ändern können

## Profilansicht für alle Benutzer

Jeder Benutzer hat Zugriff auf sein eigenes Profil unter `/profile`, wo er:
- Seine persönlichen Informationen einsehen kann
- Sein Passwort ändern kann

## Sicherheitsaspekte

1. **Passwort-Hashing**: Alle Passwörter werden mit bcrypt gehasht, bevor sie in der Datenbank gespeichert werden
2. **Rollenbasierte Zugriffssteuerung**: Benutzer können nur auf Funktionen zugreifen, für die sie berechtigt sind
3. **Validierung von Eingaben**: Alle Benutzereingaben werden validiert, um Sicherheitsrisiken zu minimieren
4. **Selbst- oder Admin-Zugriff**: Benutzer können nur ihre eigenen Daten bearbeiten, es sei denn, sie sind Administratoren

## Zusätzliche Features

- **Aktivitätsverfolgung**: Alle wichtigen Benutzeraktionen werden protokolliert und im Dashboard angezeigt
- **Responsive Design**: Die Benutzeroberfläche ist für Desktop- und mobile Geräte optimiert
- **Benutzerfreundliche Fehlermeldungen**: Verständliche Fehlermeldungen bei Problemen


## Deployment

### Automatisiertes Deployment

PeoplePilot kann mit dem mitgelieferten Deployment-Skript einfach auf jedem Linux-Server mit Docker installiert werden:

1. Lade das Deployment-Skript herunter:
   ```bash
   curl -O https://raw.githubusercontent.com/yourusername/PeoplePilot/main/scripts/peoplepilot-deploy.sh
   chmod +x peoplepilot-deploy.sh

Führe das Skript aus:
bash./peoplepilot-deploy.sh

Nach erfolgreichem Deployment ist die Anwendung im Browser unter http://[Server-IP]:5000 erreichbar.

Anpassung des Deployments
Das Deployment-Skript kann durch Umgebungsvariablen angepasst werden:
bash# Beispiel für angepasstes Deployment
REPO_BRANCH="develop" APP_PORT=8080 MONGODB_PORT=27017 ./peoplepilot-deploy.sh
Verfügbare Optionen:

REPO_URL: URL des Git-Repositories (Standard: "https://github.com/yourusername/PeoplePilot.git")
REPO_BRANCH: Branch für das Deployment (Standard: "main")
MONGODB_PORT: Port für MongoDB (Standard: 27018)
APP_PORT: Port für die Anwendung (Standard: 5000)
IMAGE_TAG: Tag für das Docker-Image (Standard: "latest")
PLATFORM: Docker-Build-Plattform (Standard: "linux/amd64")

Manuelles Deployment
Für ein manuelles Deployment folgen Sie diesen Schritten:

Klonen Sie das Repository:
bashgit clone https://github.com/yourusername/PeoplePilot.git
cd PeoplePilot

Bauen Sie das Docker-Image:
bashdocker build -t peopleflow:latest .

Erstellen Sie ein Docker-Netzwerk:
bashdocker network create peopleflow-network

Starten Sie MongoDB:
```
bashdocker run -d --name mongodb \
--network peopleflow-network \
-p 27018:27017 \
-v mongodb_data:/data/db \
--restart unless-stopped \
mongo:latest
```

Starten Sie PeoplePilot:
```
bashMONGO_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' mongodb)
docker run -d --name peopleflow \
--network peopleflow-network \
-p 5000:8080 \
-e MONGODB_URI=mongodb://${MONGO_IP}:27017/peoplepilot \
-v peopleflow_uploads:/app/uploads \
--restart unless-stopped \
peopleflow:latest
```