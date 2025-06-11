package service

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EmailService verwaltet das Senden von E-Mails
type EmailService struct {
	settingsRepo *repository.SystemSettingsRepository
}

// NewEmailService erstellt einen neuen EmailService
func NewEmailService() *EmailService {
	return &EmailService{
		settingsRepo: repository.NewSystemSettingsRepository(),
	}
}

// EmailTemplate definiert die Struktur für E-Mail-Templates
type EmailTemplate struct {
	Subject string
	Body    string
	IsHTML  bool
}

// SendEmail sendet eine E-Mail mit den angegebenen Parametern
func (es *EmailService) SendEmail(to, subject, body string, isHTML bool) error {
	settings, err := es.settingsRepo.GetSettings()
	if err != nil {
		return fmt.Errorf("fehler beim Abrufen der System-Einstellungen: %v", err)
	}

	if !settings.IsEmailConfigured() {
		return fmt.Errorf("E-Mail-Konfiguration ist nicht vollständig")
	}

	emailSettings := settings.EmailNotifications

	// SMTP-Authentifizierung
	auth := smtp.PlainAuth("", emailSettings.SMTPUser, emailSettings.SMTPPass, emailSettings.SMTPHost)

	// E-Mail-Header erstellen
	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", emailSettings.FromName, emailSettings.FromEmail))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	
	if isHTML {
		msg.WriteString("MIME-Version: 1.0\r\n")
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}
	
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// SMTP-Server-Adresse
	addr := fmt.Sprintf("%s:%d", emailSettings.SMTPHost, emailSettings.SMTPPort)

	if emailSettings.UseTLS {
		return es.sendEmailWithTLS(addr, auth, emailSettings.FromEmail, []string{to}, msg.Bytes())
	}

	return smtp.SendMail(addr, auth, emailSettings.FromEmail, []string{to}, msg.Bytes())
}

// sendEmailWithTLS sendet E-Mail über TLS-Verbindung
func (es *EmailService) sendEmailWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// TLS-Konfiguration
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         addr[:len(addr)-4], // Remove port from address
	}

	// Verbindung zum SMTP-Server aufbauen
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("fehler beim Aufbau der TLS-Verbindung: %v", err)
	}
	defer conn.Close()

	// SMTP-Client erstellen
	client, err := smtp.NewClient(conn, tlsConfig.ServerName)
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des SMTP-Clients: %v", err)
	}
	defer client.Quit()

	// Authentifizierung
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP-Authentifizierung fehlgeschlagen: %v", err)
		}
	}

	// Absender setzen
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("fehler beim Setzen des Absenders: %v", err)
	}

	// Empfänger hinzufügen
	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("fehler beim Hinzufügen des Empfängers %s: %v", recipient, err)
		}
	}

	// E-Mail-Inhalt senden
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("fehler beim Öffnen des Daten-Streams: %v", err)
	}

	_, err = writer.Write(msg)
	if err != nil {
		return fmt.Errorf("fehler beim Schreiben der E-Mail-Daten: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("fehler beim Schließen des Daten-Streams: %v", err)
	}

	return nil
}

// SendPasswordResetEmail sendet eine Passwort-Reset-E-Mail
func (es *EmailService) SendPasswordResetEmail(email, token string) error {
	resetURL := fmt.Sprintf("http://localhost:8080/reset-password?token=%s", token)
	
	subject := "Passwort zurücksetzen - PeopleFlow"
	
	// HTML-Template für Passwort-Reset
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #10b981; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 24px; background-color: #10b981; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Passwort zurücksetzen</h1>
        </div>
        <div class="content">
            <p>Hallo,</p>
            <p>Sie haben eine Anfrage zum Zurücksetzen Ihres Passworts gestellt. Klicken Sie auf den folgenden Link, um ein neues Passwort zu erstellen:</p>
            <p style="text-align: center;">
                <a href="{{.ResetURL}}" class="button">Passwort zurücksetzen</a>
            </p>
            <p>Dieser Link ist 1 Stunde gültig.</p>
            <p>Falls Sie diese Anfrage nicht gestellt haben, können Sie diese E-Mail ignorieren.</p>
            <p>Mit freundlichen Grüßen,<br>Ihr PeopleFlow Team</p>
        </div>
        <div class="footer">
            <p>Diese E-Mail wurde automatisch generiert. Bitte antworten Sie nicht auf diese E-Mail.</p>
        </div>
    </div>
</body>
</html>`

	tmpl, err := template.New("passwordReset").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("fehler beim Parsen des E-Mail-Templates: %v", err)
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, map[string]string{
		"ResetURL": resetURL,
	})
	if err != nil {
		return fmt.Errorf("fehler beim Ausführen des E-Mail-Templates: %v", err)
	}

	return es.SendEmail(email, subject, body.String(), true)
}

// SendWeeklyReport sendet einen wöchentlichen Bericht an einen Mitarbeiter
func (es *EmailService) SendWeeklyReport(employee *model.Employee, weekStart, weekEnd time.Time, totalHours float64, activities []model.Activity) error {
	subject := fmt.Sprintf("Wochenbericht %s - %s", weekStart.Format("02.01.2006"), weekEnd.Format("02.01.2006"))
	
	// HTML-Template für Wochenbericht
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #10b981; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .summary { background-color: white; padding: 15px; margin: 20px 0; border-radius: 5px; border-left: 4px solid #10b981; }
        .activity-list { margin: 20px 0; }
        .activity-item { background-color: white; padding: 10px; margin: 5px 0; border-radius: 3px; border-left: 2px solid #10b981; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
        .hours { font-size: 24px; font-weight: bold; color: #10b981; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Wochenbericht</h1>
            <p>{{.WeekStart}} - {{.WeekEnd}}</p>
        </div>
        <div class="content">
            <p>Hallo {{.EmployeeName}},</p>
            <p>hier ist Ihr Wochenbericht für die vergangene Woche:</p>
            
            <div class="summary">
                <h3>Zusammenfassung</h3>
                <p>Gesamtarbeitszeit: <span class="hours">{{.TotalHours}} Stunden</span></p>
            </div>
            
            {{if .Activities}}
            <div class="activity-list">
                <h3>Aktivitäten dieser Woche</h3>
                {{range .Activities}}
                <div class="activity-item">
                    <strong>{{.Date.Format "02.01.2006"}}</strong> - {{.Type}}<br>
                    {{if .Description}}{{.Description}}{{end}}
                </div>
                {{end}}
            </div>
            {{end}}
            
            <p>Haben Sie eine schöne Woche!</p>
            <p>Mit freundlichen Grüßen,<br>Ihr PeopleFlow Team</p>
        </div>
        <div class="footer">
            <p>Diese E-Mail wurde automatisch generiert. Bei Fragen wenden Sie sich an Ihre Personalabteilung.</p>
        </div>
    </div>
</body>
</html>`

	tmpl, err := template.New("weeklyReport").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("fehler beim Parsen des E-Mail-Templates: %v", err)
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, map[string]interface{}{
		"EmployeeName": employee.FirstName + " " + employee.LastName,
		"WeekStart":    weekStart.Format("02.01.2006"),
		"WeekEnd":      weekEnd.Format("02.01.2006"),
		"TotalHours":   fmt.Sprintf("%.1f", totalHours),
		"Activities":   activities,
	})
	if err != nil {
		return fmt.Errorf("fehler beim Ausführen des E-Mail-Templates: %v", err)
	}

	return es.SendEmail(employee.Email, subject, body.String(), true)
}

// SendTestEmail sendet eine Test-E-Mail zur Überprüfung der Konfiguration
func (es *EmailService) SendTestEmail(to string) error {
	subject := "Test-E-Mail - PeopleFlow SMTP-Konfiguration"
	
	body := `
Hallo,

dies ist eine Test-E-Mail zur Überprüfung Ihrer SMTP-Konfiguration in PeopleFlow.

Wenn Sie diese E-Mail erhalten, ist Ihre E-Mail-Konfiguration korrekt eingerichtet.

Mit freundlichen Grüßen,
Ihr PeopleFlow System

---
Diese E-Mail wurde automatisch generiert am ` + time.Now().Format("02.01.2006 15:04:05") + `
`

	return es.SendEmail(to, subject, body, false)
}

// IsEmailConfigured prüft, ob E-Mail-Funktionen verfügbar sind
func (es *EmailService) IsEmailConfigured() bool {
	settings, err := es.settingsRepo.GetSettings()
	if err != nil {
		log.Printf("Fehler beim Abrufen der E-Mail-Einstellungen: %v", err)
		return false
	}

	return settings.IsEmailConfigured()
}

// LogEmailActivity protokolliert E-Mail-Aktivitäten
func (es *EmailService) LogEmailActivity(activityType model.ActivityType, userID interface{}, description string) {
	activityRepo := repository.NewActivityRepository()
	
	// Bestimme die Benutzer-ID basierend auf dem Typ und konvertiere zu ObjectID
	var targetUserID primitive.ObjectID
	var userName string
	
	switch id := userID.(type) {
	case string:
		if objID, err := primitive.ObjectIDFromHex(id); err == nil {
			targetUserID = objID
		} else {
			// Falls die Konvertierung fehlschlägt, verwende eine neue ObjectID
			targetUserID = primitive.NewObjectID()
		}
		userName = "System"
	case primitive.ObjectID:
		targetUserID = id
		userName = "System"
	default:
		// Falls der Typ nicht unterstützt wird, verwende eine neue ObjectID
		targetUserID = primitive.NewObjectID()
		userName = "System"
	}
	
	_, err := activityRepo.LogActivity(
		activityType,
		targetUserID,
		userName,
		targetUserID,
		"email",
		"E-Mail-Service",
		description,
	)
	
	if err != nil {
		log.Printf("Fehler beim Protokollieren der E-Mail-Aktivität: %v", err)
	}
}