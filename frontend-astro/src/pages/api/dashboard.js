// API endpoint for dashboard data
import { BACKEND_URL } from '../../config';

export async function get({ request, cookies }) {
    const token = cookies.get('token')?.value;
    
    if (!token) {
        return new Response(JSON.stringify({ error: 'Unauthorized' }), {
            status: 401,
            headers: {
                'Content-Type': 'application/json',
            },
        });
    }
    
    try {
        // Fetch dashboard HTML from backend
        const response = await fetch(`${BACKEND_URL}/dashboard`, {
            method: 'GET',
            headers: {
                'Cookie': `token=${token}`,
                'Accept': 'text/html',
            },
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const html = await response.text();
        
        // Parse HTML to extract dashboard data
        const parser = new DOMParser();
        const doc = parser.parseFromString(html, 'text/html');
        
        // Detect user role
        const isUserDashboard = html.includes('USER DASHBOARD') || html.includes('Mein Dashboard');
        
        let dashboardData = {};
        
        if (isUserDashboard) {
            // Extract user-specific data
            const overtimeElement = Array.from(doc.querySelectorAll('.stat-card')).find(card => 
                card.textContent.includes('Überstunden-Saldo')
            );
            const overtimeText = overtimeElement?.querySelector('.text-2xl')?.textContent || '+0.0 Std';
            const overtimeMatch = overtimeText.match(/([+-]?\d+\.?\d*)/);
            
            const vacationElement = Array.from(doc.querySelectorAll('.stat-card')).find(card => 
                card.textContent.includes('Resturlaub')
            );
            const vacationText = vacationElement?.querySelector('.text-2xl')?.textContent || '0 Tage';
            const vacationMatch = vacationText.match(/(\d+)/);
            
            const usedVacationElement = Array.from(doc.querySelectorAll('.stat-card')).find(card => 
                card.textContent.includes('Genommener Urlaub')
            );
            const usedVacationText = usedVacationElement?.querySelector('.text-2xl')?.textContent || '0 Tage';
            const usedVacationMatch = usedVacationText.match(/(\d+)/);
            
            dashboardData = {
                userRole: 'user',
                overtimeBalance: overtimeMatch ? parseFloat(overtimeMatch[1]) : 0,
                remainingVacation: vacationMatch ? parseInt(vacationMatch[1]) : 0,
                totalVacation: 30,
                usedVacation: usedVacationMatch ? parseInt(usedVacationMatch[1]) : 0,
                pendingAbsences: 0,
                recentActivities: []
            };
        } else {
            // Extract admin/manager data
            const employeeElement = Array.from(doc.querySelectorAll('.stat-card, .flex.items-center.p-4.bg-white')).find(card => 
                card.textContent.includes('Gesamtmitarbeiter')
            );
            const employeeText = employeeElement?.querySelector('.text-2xl')?.textContent || '0';
            
            const costsElement = Array.from(doc.querySelectorAll('.stat-card, .flex.items-center.p-4.bg-white')).find(card => 
                card.textContent.includes('Personalkosten')
            );
            const costsText = costsElement?.querySelector('.text-2xl')?.textContent || '0 €';
            
            const reviewsElement = Array.from(doc.querySelectorAll('.stat-card, .flex.items-center.p-4.bg-white')).find(card => 
                card.textContent.includes('Gespräche')
            );
            const reviewsText = reviewsElement?.querySelector('.text-2xl')?.textContent || '0';
            
            const documentsElement = Array.from(doc.querySelectorAll('.stat-card, .flex.items-center.p-4.bg-white')).find(card => 
                card.textContent.includes('Dokumente')
            );
            const documentsText = documentsElement?.querySelector('.text-2xl')?.textContent || '0';
            
            dashboardData = {
                userRole: 'admin',
                totalEmployees: parseInt(employeeText) || 45,
                monthlyLaborCosts: costsText.replace('€', '').trim() || "48,325.00",
                upcomingReviews: parseInt(reviewsText) || 3,
                expiredDocuments: parseInt(documentsText) || 2,
                recentActivities: [],
                chartData: {
                    monthlyTrend: {
                        labels: ['Jan', 'Feb', 'Mär', 'Apr', 'Mai', 'Jun'],
                        datasets: [{
                            label: 'Personalkosten',
                            data: [42000, 44500, 46000, 48325, 47800, 49000],
                            borderColor: '#22c55e',
                            backgroundColor: 'rgba(34, 197, 94, 0.1)',
                            tension: 0.4
                        }]
                    },
                    departmentDistribution: {
                        labels: ['Entwicklung', 'Marketing', 'Vertrieb', 'HR', 'Finanzen'],
                        datasets: [{
                            data: [12, 8, 15, 5, 5],
                            backgroundColor: [
                                '#22c55e',
                                '#16a34a',
                                '#15803d',
                                '#166534',
                                '#14532d'
                            ]
                        }]
                    }
                }
            };
        }
        
        return new Response(JSON.stringify(dashboardData), {
            status: 200,
            headers: {
                'Content-Type': 'application/json',
            },
        });
        
    } catch (error) {
        console.error('Dashboard API error:', error);
        return new Response(JSON.stringify({ 
            error: 'Failed to fetch dashboard data',
            message: error.message 
        }), {
            status: 500,
            headers: {
                'Content-Type': 'application/json',
            },
        });
    }
}