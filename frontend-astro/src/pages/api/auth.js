// src/pages/api/auth.js
import { BACKEND_URL } from '../../config';

export async function post({ request, cookies, redirect }) {
    try {
        // Formular-Daten abrufen
        const formData = await request.formData();
        const email = formData.get('email');
        const password = formData.get('password');

        // Backend-Anfrage vorbereiten
        const response = await fetch(`${BACKEND_URL}/auth`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: new URLSearchParams({
                email,
                password,
            }),
            // Nicht automatisch Cookies folgen
            redirect: 'manual',
        });

        // Erfolgreiche Anmeldung - Cookie extrahieren
        if (response.status === 302) {
            // Cookie-Header aus der Antwort lesen
            const cookieHeader = response.headers.get('set-cookie');
            if (cookieHeader) {
                // Token extrahieren
                const tokenMatch = cookieHeader.match(/token=([^;]+)/);
                if (tokenMatch && tokenMatch[1]) {
                    // Token in den Astro-Cookies speichern
                    cookies.set('token', tokenMatch[1], {
                        path: '/',
                        httpOnly: true,
                        sameSite: 'strict',
                        secure: process.env.NODE_ENV === 'production',
                    });

                    // Zum Astro-Dashboard umleiten
                    return redirect('/dashboard');
                }
            }
        }

        // Bei Fehler zurück zur Login-Seite mit Fehlermeldung
        return redirect('/login?error=Anmeldung fehlgeschlagen. Bitte überprüfen Sie Ihre Zugangsdaten.');
    } catch (error) {
        console.error('Authentication error:', error);
        return redirect('/login?error=Ein Fehler ist aufgetreten. Bitte versuchen Sie es später erneut.');
    }
}