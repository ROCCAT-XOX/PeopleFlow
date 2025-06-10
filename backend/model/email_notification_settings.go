package model

type EmailNotificationSettings struct {
	Enabled   bool   `bson:"enabled" json:"enabled"`
	SMTPHost  string `bson:"smtpHost" json:"smtpHost"`
	SMTPPort  int    `bson:"smtpPort" json:"smtpPort"`
	SMTPUser  string `bson:"smtpUser" json:"smtpUser"`
	SMTPPass  string `bson:"smtpPass" json:"smtpPass"`
	FromEmail string `bson:"fromEmail" json:"fromEmail"`
	FromName  string `bson:"fromName" json:"fromName"`
	UseTLS    bool   `bson:"useTLS" json:"useTLS"`
}
