# Anleitung zur Einrichtung des PeoplePilot-Projekts mit Astro.js

## Projektstruktur

Hier ist die empfohlene Verzeichnisstruktur für dein PeoplePilot-Projekt mit Astro.js:

```
people-pilot/
├── astro.config.mjs
├── package.json
├── public/
│   ├── favicon.svg
│   ├── images/
│   │   └── PeopleFlow-Logoschrift.svg
├── src/
│   ├── components/
│   │   └── Login.astro
│   ├── layouts/
│   │   └── Layout.astro
│   ├── lib/
│   │   └── auth.js
│   └── pages/
│       ├── auth.astro
│       ├── dashboard.astro (noch zu erstellen)
│       ├── index.astro
│       ├── login.astro
│       └── logout.astro
└── tailwind.config.cjs
```

## Einrichtungsschritte

1. **Projekt initialisieren**

   ```bash
   # Neues Astro-Projekt erstellen
   npm create astro@latest people-pilot
   
   # In das Projektverzeichnis wechseln
   cd people-pilot
   
   # Tailwind CSS Integration hinzufügen
   npx astro add tailwind
   
   # Weitere Abhängigkeiten installieren
   npm install bcryptjs jsonwebtoken particles.js
   ```

2. **Dateien kopieren**

   Kopiere alle erstellten Dateien in die entsprechenden Verzeichnisse der Projektstruktur.

3. **Ressourcen vorbereiten**

   Stelle sicher, dass du das PeopleFlow-Logo-SVG in das Verzeichnis `public/images/` kopierst.

4. **Server starten**

   ```bash
   npm run dev
   ```

   Die Anwendung sollte jetzt unter http://localhost:3000 erreichbar sein.

## Migration von Go zu Astro.js

Während der Migration von deinem Go-Backend zu Astro.js solltest du folgendes beachten:

1. **Datenbank-Integration**:
    - Astro.js ist primär ein Frontend-Framework, du wirst für die Datenbankanbindung zusätzliche Bibliotheken benötigen
    - Für MongoDB kannst du das `mongodb`-Paket installieren: `npm install mongodb`
    - Für eine komplette Backend-Integration könntest du auch Astro mit einem Node.js API-Framework wie Express kombinieren

2. **Authentifizierung**:
    - Die bereitgestellte Authentifizierungslogik nutzt JWT-Tokens ähnlich deiner Go-Implementierung
    - Für eine Produktionsumgebung sollte das JWT-Secret aus Umgebungsvariablen geladen werden
    - Passwort-Hashing wird mit bcrypt durchgeführt, ähnlich wie in deinem Go-Code

3. **API-Endpunkte**:
    - Mit Astro.js kannst du API-Endpunkte über .astro-Dateien im pages-Verzeichnis erstellen
    - Für komplexere APIs könntest du den Astro SSR-Adapter für Node.js verwenden

4. **UI-Komponenten**:
    - Deine bestehenden HTML-Templates können schrittweise in Astro-Komponenten umgewandelt werden
    - Tailwind CSS funktioniert nahtlos mit Astro

## Nächste Schritte

Nach der Implementierung des Login-Systems solltest du folgende Komponenten migrieren:

1. Dashboard-Seite
2. Mitarbeiter-Übersicht
3. Benutzerprofilseite
4. Einstellungsseite

Für jede dieser Komponenten solltest du:
- Eine entsprechende Astro-Komponente erstellen
- Die Datenbankanbindung implementieren
- Die Benutzerauthentifizierung und -autorisierung sicherstellen

Viel Erfolg bei der Migration deines PeoplePilot HR-Tools zu Astro.js!