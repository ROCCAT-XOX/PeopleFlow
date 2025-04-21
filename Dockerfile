# ========== Build Stage ==========
FROM alpine:latest AS builder

# Go 1.23.4 installieren
RUN apk add --no-cache ca-certificates gcc git musl-dev

# Download und Installation von Go 1.23.4
ENV GOLANG_VERSION=1.23.4 \
    GOPATH=/go \
    PATH=/go/bin:/usr/local/go/bin:$PATH

# Explizit für AMD64 bauen, unabhängig von der Architektur des Build-Hosts
RUN wget -O go.tgz "https://dl.google.com/go/go${GOLANG_VERSION}.linux-amd64.tar.gz" && \
    tar -C /usr/local -xzf go.tgz && \
    rm go.tgz && \
    mkdir -p "$GOPATH/src" "$GOPATH/bin" && \
    chmod -R 777 "$GOPATH"

WORKDIR /app

# Go-Module-Abhängigkeiten cachen
COPY go.mod go.sum ./
RUN go mod download

# Quellcode kopieren
COPY . .

# Binary explizit für AMD64-Architektur bauen
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o PeopleFlow

# ========== Final Stage ==========
FROM alpine:latest

WORKDIR /app

# Laufzeit-Abhängigkeiten
RUN apk --no-cache add ca-certificates tzdata

# Binary und Assets übernehmen
COPY --from=builder /app/PeoplePilot /app/PeoplePilot
COPY --from=builder /app/frontend /app/frontend

# Upload-Verzeichnis anlegen und als Volume deklarieren
RUN mkdir -p /app/uploads && chmod 777 /app/uploads
VOLUME ["/app/uploads"]

# Umgebungsvariablen
ENV GIN_MODE=release \
    TZ=Europe/Berlin \
    PORT=8080 \
    MONGODB_URI=mongodb://mongodb:27017/peoplepilot

# Port freigeben
EXPOSE 8080

# Startkommando
CMD ["/app/PeoplePilot"]