#!/bin/bash
# PeopleFlow-deploy.sh - Deployment-Skript für PeopleFlow mit Update-Logik
# Farbcodes für Ausgabe
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Variablen
REPO_URL=${REPO_URL:-"https://github.com/yourusername/PeopleFlow.git"}
REPO_BRANCH=${REPO_BRANCH:-"main"}
REPO_DIR=${REPO_DIR:-"./PeopleFlow"}
MONGODB_PORT=${MONGODB_PORT:-27017}
APP_PORT=${APP_PORT:-5000}
IMAGE_TAG=${IMAGE_TAG:-"latest"}
PLATFORM=${PLATFORM:-"linux/amd64"}

handle_error() {
  echo -e "${RED}FEHLER: $1${NC}"
  exit 1
}

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     PeopleFlow Smart Deployment       ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"

# 1. Git-Repository aktualisieren oder klonen
if [ -d "$REPO_DIR/.git" ]; then
  echo -e "${YELLOW}Repository bereits vorhanden, aktualisiere...${NC}"
  cd $REPO_DIR || handle_error "Konnte nicht ins Repository-Verzeichnis wechseln"
  git fetch --all || handle_error "Git fetch fehlgeschlagen"
  git reset --hard origin/$REPO_BRANCH || handle_error "Git reset fehlgeschlagen"
  git pull origin $REPO_BRANCH || handle_error "Git pull fehlgeschlagen"
else
  echo -e "${YELLOW}Klone Repository...${NC}"
  git clone --branch $REPO_BRANCH $REPO_URL $REPO_DIR || handle_error "Git clone fehlgeschlagen"
  cd $REPO_DIR || handle_error "Konnte nicht ins Repository-Verzeichnis wechseln"
fi

echo -e "${GREEN}Repository erfolgreich aktualisiert.${NC}"

# 2. Docker-Image bauen
echo -e "${YELLOW}Baue Docker-Image für $PLATFORM...${NC}"
docker build --platform $PLATFORM -t peopleflow:$IMAGE_TAG . || handle_error "Docker Build fehlgeschlagen"
echo -e "${GREEN}Docker-Image erfolgreich gebaut.${NC}"

# 3. Docker-Netzwerk sicherstellen
echo -e "${YELLOW}Stelle sicher, dass das Netzwerk existiert...${NC}"
docker network create peopleflow-network 2>/dev/null || true

# 4. MongoDB prüfen
if docker ps -q -f name=mongodb >/dev/null; then
  echo -e "${GREEN}MongoDB läuft bereits. Kein Neustart erforderlich.${NC}"
else
  if docker ps -a -q -f name=mongodb >/dev/null; then
    echo -e "${YELLOW}Starte vorhandenen MongoDB-Container...${NC}"
    docker start mongodb || handle_error "MongoDB konnte nicht gestartet werden"
  else
    echo -e "${YELLOW}Starte neuen MongoDB 4.4.18-Container...${NC}"
    docker run -d --name mongodb \
      --network peopleflow-network \
      -p $MONGODB_PORT:27017 \
      -v mongodb_data:/data/db \
      --restart unless-stopped \
      mongo:4.4.18 || handle_error "MongoDB-Start fehlgeschlagen"
  fi
fi

# Warte auf MongoDB, wenn sie gerade neu gestartet wurde
sleep 3
MONGO_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' mongodb)
echo -e "${GREEN}MongoDB läuft auf IP: ${MONGO_IP}${NC}"

# 5. PeopleFlow neu starten
if docker ps -q -f name=peopleflow >/dev/null; then
  echo -e "${YELLOW}Stoppe und entferne alten PeopleFlow-Container...${NC}"
  docker stop peopleflow && docker rm peopleflow
elif docker ps -a -q -f name=peopleflow >/dev/null; then
  echo -e "${YELLOW}Entferne alten, gestoppten PeopleFlow-Container...${NC}"
  docker rm peopleflow
fi

echo -e "${YELLOW}Starte neuen PeopleFlow-Container...${NC}"
docker run -d --name peopleflow \
  --network peopleflow-network \
  -p $APP_PORT:8080 \
  -e MONGODB_URI=mongodb://${MONGO_IP}:27017/PeopleFlow \
  -v peopleflow_uploads:/app/uploads \
  --restart unless-stopped \
  peopleflow:$IMAGE_TAG || handle_error "PeopleFlow-Start fehlgeschlagen"

# 6. Finaler Status
echo -e "${GREEN}PeopleFlow erfolgreich (re)deployt!${NC}"
SERVER_IP=$(hostname -I | awk '{print $1}')
echo -e "${GREEN}Die Anwendung ist erreichbar unter: http://${SERVER_IP}:${APP_PORT}${NC}"
echo -e "${YELLOW}MongoDB läuft auf Port: ${MONGODB_PORT}${NC}"

echo -e "\n${YELLOW}Log-Ausgabe (letzte 10 Zeilen):${NC}"
docker logs peopleflow --tail 10

echo -e "\n${BLUE}════════════════════════════════════════${NC}"
echo -e "${GREEN}Deployment abgeschlossen!${NC}"
echo -e "${BLUE}════════════════════════════════════════${NC}"