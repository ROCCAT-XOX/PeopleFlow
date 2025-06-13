// Debug version of auth endpoint
import { BACKEND_URL } from '../../config';

export async function post({ request, cookies, redirect }) {
    try {
        const formData = await request.formData();
        const email = formData.get('email');
        const password = formData.get('password');

        console.log('=== AUTH DEBUG ===');
        console.log('Email:', email);
        console.log('Backend URL:', BACKEND_URL);

        // Create form data for backend
        const formDataForBackend = new URLSearchParams();
        formDataForBackend.append('email', email);
        formDataForBackend.append('password', password);
        
        const response = await fetch(`${BACKEND_URL}/auth`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: formDataForBackend.toString(),
            credentials: 'include',
            redirect: 'manual',
        });

        console.log('Response Status:', response.status);
        console.log('Response Headers:', Object.fromEntries(response.headers.entries()));
        
        // Get all headers
        const headers = {};
        response.headers.forEach((value, key) => {
            headers[key] = value;
        });
        console.log('All Headers:', headers);

        if (response.status === 302) {
            const setCookieHeader = response.headers.get('set-cookie');
            console.log('Set-Cookie Header:', setCookieHeader);
            
            if (setCookieHeader) {
                const tokenMatch = setCookieHeader.match(/token=([^;]+)/);
                if (tokenMatch && tokenMatch[1]) {
                    const token = tokenMatch[1];
                    console.log('Extracted Token:', token);
                    
                    // Set token in Astro cookies
                    cookies.set('token', token, {
                        path: '/',
                        httpOnly: true,
                        sameSite: 'lax',
                        secure: false,
                        maxAge: 60 * 60 * 24,
                    });
                    
                    return new Response(JSON.stringify({ success: true, redirect: '/dashboard' }), {
                        status: 200,
                        headers: { 'Content-Type': 'application/json' }
                    });
                }
            }
        }

        const responseText = await response.text();
        console.log('Response Body:', responseText.substring(0, 200));
        
        return new Response(JSON.stringify({ 
            success: false, 
            error: 'Login failed',
            status: response.status,
            headers: headers
        }), {
            status: 400,
            headers: { 'Content-Type': 'application/json' }
        });
        
    } catch (error) {
        console.error('Auth Error:', error);
        return new Response(JSON.stringify({ 
            success: false, 
            error: error.message 
        }), {
            status: 500,
            headers: { 'Content-Type': 'application/json' }
        });
    }
}