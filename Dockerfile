# Build Stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Kopieren der Go Module Dateien
COPY go.mod go.sum ./
RUN go mod download

# Kopieren des Quellcodes
COPY . .

# Build der Anwendung
RUN CGO_ENABLED=0 GOOS=linux go build -o peoplepilot

# Final Stage
FROM alpine:latest

WORKDIR /app

# Benötigte Pakete installieren
RUN apk --no-cache add ca-certificates tzdata

# Zeitzone setzen
ENV TZ=Europe/Berlin

# Kopieren der kompilierten Binary aus dem Build-Stage
COPY --from=builder /app/peoplepilot .

# Kopieren der Templates und statischen Dateien
COPY --from=builder /app/frontend /app/frontend

# Erstellen des Upload-Verzeichnisses
RUN mkdir -p /app/uploads && chmod 777 /app/uploads

# Setzen der Umgebungsvariablen
ENV GIN_MODE=release
ENV PORT=8080
ENV MONGODB_URI=mongodb://mongo:27017

# Freigabe des Ports
EXPOSE 8080

# Starten der Anwendung
CMD ["./peoplepilot"]