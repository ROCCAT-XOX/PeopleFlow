// auth.js - Authentifizierungslogik für Astro.js
// Die Cookies-API wird nur in .astro-Dateien verwendet, nicht hier
import bcrypt from 'bcryptjs';
import jwt from 'jsonwebtoken';

// JWT-Secret sollte in der Produktion aus einer Umgebungsvariable kommen
const JWT_SECRET = import.meta.env.JWT_SECRET || 'your-secret-key';

// Fiktive Funktion, um einen Benutzer aus der Datenbank zu finden
// In der tatsächlichen Implementierung würdest du MongoDB verwenden
async function findUserByEmail(email) {
    // Hier würde die MongoDB-Integrationslogik kommen,
    // ähnlich zu deinem Repository-Pattern im Go-Backend

    // Für das Beispiel geben wir einen Beispielbenutzer zurück, wenn die E-Mail stimmt
    if (email === 'admin@PeopleFlow.com') {
        return {
            id: '1',
            email: 'admin@PeopleFlow.com',
            firstName: 'Admin',
            lastName: 'User',
            // Das Passwort wäre in einer echten Implementierung gehashed
            password: '$2a$10$1r1Mh1SJukZVNqjQRCH9UOk06UyCDI5F1G.ixW2H3Yg6OiXq4q9iC', // "admin"
            role: 'admin',
            status: 'active'
        };
    }

    return null;
}

// Token generieren
function generateToken(userId, role) {
    return jwt.sign(
        {
            userId,
            role,
            iat: Math.floor(Date.now() / 1000),
            exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60) // 24 Stunden
        },
        JWT_SECRET
    );
}

// Passwort überprüfen
async function checkPassword(plainPassword, hashedPassword) {
    return await bcrypt.compare(plainPassword, hashedPassword);
}

// Authentifizierungsfunktion
export async function authenticate(email, password) {
    try {
        // Benutzer suchen
        const user = await findUserByEmail(email);

        // Wenn kein Benutzer gefunden wurde
        if (!user) {
            return { success: false, error: 'Ungültige E-Mail oder Passwort' };
        }

        // Überprüfen, ob der Benutzer aktiv ist
        if (user.status !== 'active') {
            return { success: false, error: 'Ihr Konto ist inaktiv' };
        }

        // Passwort überprüfen
        const isPasswordValid = await checkPassword(password, user.password);
        if (!isPasswordValid) {
            return { success: false, error: 'Ungültige E-Mail oder Passwort' };
        }

        // Token generieren
        const token = generateToken(user.id, user.role);

        return {
            success: true,
            user: {
                id: user.id,
                firstName: user.firstName,
                lastName: user.lastName,
                email: user.email,
                role: user.role
            },
            token
        };
    } catch (error) {
        console.error('Authentication error:', error);
        return { success: false, error: 'Ein interner Fehler ist aufgetreten' };
    }
}

// Token validieren
export function validateToken(token) {
    try {
        const decoded = jwt.verify(token, JWT_SECRET);
        return { valid: true, userId: decoded.userId, role: decoded.role };
    } catch (error) {
        return { valid: false };
    }
}

// Middleware zum Überprüfen der Authentifizierung
export function requireAuth(Astro) {
    const token = Astro.cookies.get('token')?.value;

    if (!token) {
        return Astro.redirect('/login');
    }

    const { valid } = validateToken(token);

    if (!valid) {
        // Token-Cookie löschen
        Astro.cookies.delete('token', { path: '/' });
        return Astro.redirect('/login');
    }

    // Authentifizierung erfolgreich, weitermachen
    return null;
}