// src/pages/api/auth.js
import { BACKEND_URL } from '../../config';

export async function post({ request, cookies, redirect }) {
    try {
        // Formular-Daten abrufen
        const formData = await request.formData();
        const email = formData.get('email');
        const password = formData.get('password');

        console.log(`Sende Anmeldeanfrage an: ${BACKEND_URL}/auth`);

        // Backend-Anfrage vorbereiten
        const response = await fetch(`${BACKEND_URL}/auth`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                email,
                password,
            }),
        });

        console.log('Antwort vom Backend erhalten:', response.status);

        // JSON-Antwort abrufen
        const data = await response.json();

        // Erfolgreiche Anmeldung
        if (response.ok && data.token) {
            // Token in den Astro-Cookies speichern
            cookies.set('token', data.token, {
                path: '/',
                httpOnly: true,
                sameSite: 'lax',
                secure: process.env.NODE_ENV === 'production',
                maxAge: 60 * 60 * 24, // 1 Tag
            });

            // Zum Astro-Dashboard umleiten
            return redirect('/dashboard');
        }

        // Bei Fehler zurück zur Login-Seite mit Fehlermeldung
        return redirect('/login?error=' + encodeURIComponent(data.message || 'Anmeldung fehlgeschlagen. Bitte überprüfen Sie Ihre Zugangsdaten.'));
    } catch (error) {
        console.error('Authentication error:', error);
        return redirect('/login?error=' + encodeURIComponent('Ein Fehler ist aufgetreten. Bitte versuchen Sie es später erneut.'));
    }
}