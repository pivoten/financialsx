import './style.css';
import './app.css';
import './auth.css';

import logo from './assets/images/pivoten.logo.png';
import { auth } from './auth';
import { LoginForm, RegisterForm, Dashboard } from './components';
// Logout is now imported in auth.js

class App {
    constructor() {
        this.appElement = document.querySelector('#app');
        this.init().catch(console.error);
    }

    async init() {
        try {
            console.log('Initializing app...');
            
            // Show loading message
            this.appElement.innerHTML = '<div style="padding: 20px; text-align: center;">Loading...</div>';
            
            // Check if user has valid session
            console.log('Checking session...');
            const hasSession = await auth.checkSession();
            console.log('Has session:', hasSession);
            
            if (hasSession) {
                console.log('Showing dashboard');
                this.showDashboard();
            } else {
                console.log('Showing login');
                this.showLogin();
            }
        } catch (error) {
            console.error('Error during app initialization:', error);
            this.appElement.innerHTML = '<div style="padding: 20px; color: red;">Error loading app: ' + error.message + '</div>';
        }
    }

    async showLogin() {
        try {
            console.log('Loading companies...');
            const companies = await auth.loadCompanies();
            console.log('Companies loaded:', companies);
            
            this.appElement.innerHTML = LoginForm(null, companies);
        
        // Add logo
        const logoContainer = document.createElement('div');
        logoContainer.className = 'logo-container';
        logoContainer.innerHTML = `<img src="${logo}" alt="Pivoten Logo" class="auth-logo">`;
        this.appElement.insertBefore(logoContainer, this.appElement.firstChild);
        
        // Setup form handlers
        const form = document.getElementById('login-form');
        const errorDiv = document.getElementById('login-error');
        
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            errorDiv.textContent = '';
            
            const formData = new FormData(form);
            const username = formData.get('username');
            const password = formData.get('password');
            const company = formData.get('company');
            
            if (!company) {
                errorDiv.textContent = 'Please select a company';
                return;
            }
            
            try {
                await auth.login(username, password, company);
                this.showDashboard();
            } catch (err) {
                errorDiv.textContent = err.message || 'Login failed';
            }
        });
        
        // Switch to register
        document.getElementById('switch-to-register').addEventListener('click', (e) => {
            e.preventDefault();
            this.showRegister();
        });
        } catch (error) {
            console.error('Error in showLogin:', error);
            this.appElement.innerHTML = '<div style="padding: 20px; color: red;">Error loading login page: ' + error.message + '</div>';
        }
    }

    async showRegister() {
        const companies = await auth.loadCompanies();
        
        this.appElement.innerHTML = RegisterForm(null, companies);
        
        // Add logo
        const logoContainer = document.createElement('div');
        logoContainer.className = 'logo-container';
        logoContainer.innerHTML = `<img src="${logo}" alt="Pivoten Logo" class="auth-logo">`;
        this.appElement.insertBefore(logoContainer, this.appElement.firstChild);
        
        // Setup form handlers
        const form = document.getElementById('register-form');
        const errorDiv = document.getElementById('register-error');
        
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            errorDiv.textContent = '';
            
            const formData = new FormData(form);
            const username = formData.get('username');
            const password = formData.get('password');
            const confirmPassword = formData.get('confirm-password');
            const email = formData.get('email');
            const company = formData.get('company');
            
            if (!company) {
                errorDiv.textContent = 'Please select a company';
                return;
            }
            
            if (password !== confirmPassword) {
                errorDiv.textContent = 'Passwords do not match';
                return;
            }
            
            try {
                await auth.register(username, password, email, company);
                this.showDashboard();
            } catch (err) {
                errorDiv.textContent = err.message || 'Registration failed';
            }
        });
        
        // Switch to login
        document.getElementById('switch-to-login').addEventListener('click', (e) => {
            e.preventDefault();
            this.showLogin();
        });
    }

    showDashboard() {
        if (!auth.currentUser) {
            this.showLogin();
            return;
        }
        
        this.appElement.innerHTML = Dashboard(auth.currentUser);
        
        // Setup logout handler
        document.getElementById('logout-btn').addEventListener('click', async () => {
            await auth.logout();
            this.showLogin();
        });
    }
}

// Add additional styles for auth pages
const style = document.createElement('style');
style.textContent = `
    .logo-container {
        text-align: center;
        margin: 40px 0 30px 0;
        padding-top: 20px;
    }
    
    .auth-logo {
        max-width: 200px;
        height: auto;
    }
    
    #app {
        min-height: 100vh;
        background-color: #f5f5f5;
    }
    
    body {
        margin: 0;
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
    }
`;
document.head.appendChild(style);

// Initialize app
new App();