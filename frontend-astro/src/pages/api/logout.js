import { BACKEND_URL } from '../../config';

export async function get({ cookies, redirect }) {
    try {
        // Logout beim Backend aufrufen (optional)
        await fetch(`${BACKEND_URL}/logout`, {
            method: 'GET',
            credentials: 'include',
        });

        // Lokales Cookie löschen
        cookies.delete('token', { path: '/' });

        // Zur Login-Seite umleiten
        return redirect('/login');
    } catch (error) {
        console.error('Logout error:', error);

        // Im Fehlerfall trotzdem lokales Cookie löschen und umleiten
        cookies.delete('token', { path: '/' });
        return redirect('/login');
    }
}