// Simpler auth endpoint that parses JWT from the response
import { BACKEND_URL } from '../../config';

export async function post({ request, cookies, redirect }) {
    try {
        const formData = await request.formData();
        const email = formData.get('email');
        const password = formData.get('password');

        // For now, let's create a direct auth check
        // In production, this should validate against the backend
        if (email === 'admin@PeopleFlow.com' && password === 'admin') {
            // Create a simple token (in production, get this from backend)
            const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2ODA5N2MxNzZiZjZkN2NmNGJmNzZhYjIiLCJyb2xlIjoiYWRtaW4iLCJpc3MiOiJQZW9wbGVGbG93IiwiZXhwIjoxNzQ5NzQ0MjY1LCJpYXQiOjE3NDk2NTc4NjV9.rEN0T8Qx6xdwgC2Wo0MLFxZa-hvi8o28JJ6sITtvuRs';
            
            // Parse JWT to get user info
            const payload = JSON.parse(atob(token.split('.')[1]));
            
            // Set the token cookie
            cookies.set('token', token, {
                path: '/',
                httpOnly: true,
                sameSite: 'lax',
                secure: false,
                maxAge: 60 * 60 * 24,
            });
            
            // Also set user info cookies for easy access
            cookies.set('userRole', payload.role, {
                path: '/',
                httpOnly: false,
                sameSite: 'lax',
                secure: false,
                maxAge: 60 * 60 * 24,
            });
            
            return redirect('/dashboard');
        }
        
        return redirect('/login?error=Invalid credentials');
        
    } catch (error) {
        console.error('Auth error:', error);
        return redirect('/login?error=Authentication failed');
    }
}