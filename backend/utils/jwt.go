package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var jwtSecret = []byte("your-secret-key") // In einer Produktionsumgebung sollte dieser Wert aus einer Umgebungsvariable kommen

// Claims repräsentiert die JWT-Claims
type Claims struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT generiert ein JWT-Token für den angegebenen Benutzer
func GenerateJWT(userID, role string) (string, error) {
	// Claims erstellen
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token läuft nach 24 Stunden ab
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "peoplepilot",
		},
	}

	// Token generieren
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Token signieren
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validiert ein JWT-Token und gibt die Claims zurück
func ValidateJWT(tokenString string) (*Claims, error) {
	// Token parsen
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// Claims extrahieren
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
