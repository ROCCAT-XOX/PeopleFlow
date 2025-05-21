// src/middleware.js
import { defineMiddleware } from 'astro:middleware';
import { BACKEND_URL } from './config';

export const onRequest = defineMiddleware(async ({ request, cookies, locals, redirect }, next) => {
    const url = new URL(request.url);

    // Öffentliche Routen ohne Authentifizierung
    const publicRoutes = ['/login', '/register', '/reset-password'];
    if (publicRoutes.includes(url.pathname)) {
        return next();
    }

    // Token aus Cookies abrufen
    const token = cookies.get('token')?.value;

    // Wenn kein Token vorhanden ist, zur Login-Seite umleiten
    if (!token) {
        return redirect('/login');
    }

    try {
        // Token validieren durch Anfrage an das Backend
        const response = await fetch(`${BACKEND_URL}/api/validate-token`, {
            method: 'GET',
            headers: {
                'Cookie': `token=${token}`,
            },
            credentials: 'include',
        });

        if (!response.ok) {
            // Token ist ungültig, zur Login-Seite umleiten
            cookies.delete('token');
            return redirect('/login');
        }

        // Benutzerdaten aus der Antwort extrahieren und zu locals hinzufügen
        const userData = await response.json();
        locals.user = userData;

        // Anfrage fortsetzen
        return next();
    } catch (error) {
        console.error('Fehler bei der Token-Validierung:', error);
        return redirect('/login');
    }
});