#!/bin/bash
# PeopleFlow-deploy.sh - Vollständiges Deployment-Skript für PeopleFlow
# Farbcodes für schönere Ausgabe
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Konfigurierbare Variablen
MONGODB_PORT=${MONGODB_PORT:-27017}  # Jetzt Standard-Port
APP_PORT=${APP_PORT:-5000}
IMAGE_TAG=${IMAGE_TAG:-"latest"}
PLATFORM=${PLATFORM:-"linux/amd64"}

# Hilfsfunktion für Fehlerbehandlung
handle_error() {
  echo -e "${RED}FEHLER: $1${NC}"
  exit 1
}

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║      PeopleFlow Deployment Skript     ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"

# 1. Docker-Image bauen
echo -e "${YELLOW}Baue Docker-Image für $PLATFORM...${NC}"
docker build --platform $PLATFORM -t peopleflow:$IMAGE_TAG . || handle_error "Docker Build fehlgeschlagen"
echo -e "${GREEN}Docker-Image erfolgreich gebaut.${NC}"

# 2. Alte Container stoppen und entfernen
echo -e "${YELLOW}Stoppe alte Container...${NC}"
docker stop peopleflow mongodb 2>/dev/null
docker rm peopleflow mongodb 2>/dev/null
echo -e "${GREEN}Alte Container bereinigt.${NC}"

# 3. Netzwerk erstellen (falls nicht vorhanden)
echo -e "${YELLOW}Erstelle Docker-Netzwerk...${NC}"
docker network create peopleflow-network 2>/dev/null || true
echo -e "${GREEN}Netzwerk bereit.${NC}"

# 4. MongoDB 4.4.18 starten
echo -e "${YELLOW}Starte MongoDB 4.4.18...${NC}"
docker run -d --name mongodb \
  --network peopleflow-network \
  -p $MONGODB_PORT:27017 \
  -v mongodb_data:/data/db \
  --restart unless-stopped \
  mongo:4.4.18 || handle_error "MongoDB-Start fehlgeschlagen"

# Warte kurz, bis MongoDB gestartet ist
echo -e "${YELLOW}Warte auf MongoDB-Start...${NC}"
sleep 5

# 5. MongoDB-IP ermitteln
MONGO_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' mongodb)
echo -e "${GREEN}MongoDB läuft auf IP: ${MONGO_IP}${NC}"

# 6. PeopleFlow starten
echo -e "${YELLOW}Starte PeopleFlow...${NC}"
docker run -d --name peopleflow \
  --network peopleflow-network \
  -p $APP_PORT:8080 \
  -e MONGODB_URI=mongodb://${MONGO_IP}:27017/PeopleFlow \
  -v peopleflow_uploads:/app/uploads \
  --restart unless-stopped \
  peopleflow:$IMAGE_TAG || handle_error "PeopleFlow-Start fehlgeschlagen"

# 7. Status prüfen
echo -e "${YELLOW}Prüfe Container-Status...${NC}"
if [ "$(docker ps -q -f name=peopleflow)" ] && [ "$(docker ps -q -f name=mongodb)" ]; then
  echo -e "${GREEN}PeopleFlow wurde erfolgreich gestartet!${NC}"
  SERVER_IP=$(hostname -I | awk '{print $1}')
  echo -e "${GREEN}Die Anwendung ist verfügbar unter: http://${SERVER_IP}:${APP_PORT}${NC}"
  echo -e "${YELLOW}MongoDB läuft auf Port: ${MONGODB_PORT}${NC}"

  # Container-Logs anzeigen
  echo -e "\n${YELLOW}Log-Ausgabe des PeopleFlow-Containers:${NC}"
  docker logs peopleflow --tail 10
else
  handle_error "Es gab ein Problem beim Starten der Container. Überprüfe die Logs mit: docker logs peopleflow"
fi

echo -e "\n${BLUE}════════════════════════════════════════${NC}"
echo -e "${GREEN}Deployment abgeschlossen!${NC}"
echo -e "${BLUE}════════════════════════════════════════${NC}"
