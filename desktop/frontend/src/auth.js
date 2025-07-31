import { Login, Register, GetCompanies, ValidateSession, Logout } from '../wailsjs/go/main/App';

export class Auth {
    constructor() {
        this.currentUser = null;
        this.session = null;
        this.companies = [];
    }

    async loadCompanies() {
        try {
            console.log('Calling GetCompanies...');
            this.companies = await GetCompanies();
            console.log('GetCompanies result:', this.companies);
            // Ensure we always return an array
            return Array.isArray(this.companies) ? this.companies : [];
        } catch (err) {
            console.error('Failed to load companies:', err);
            return [];
        }
    }

    async login(username, password, companyName) {
        try {
            const result = await Login(username, password, companyName);
            this.currentUser = result.user;
            this.session = result.session;
            
            // Store session in localStorage
            localStorage.setItem('session_token', this.session.token);
            localStorage.setItem('company_name', companyName);
            
            return result;
        } catch (err) {
            throw err;
        }
    }

    async register(username, password, email, companyName) {
        try {
            console.log('Registering:', { username, email, companyName });
            const result = await Register(username, password, email, companyName);
            console.log('Register result:', result);
            this.currentUser = result.user;
            this.session = result.session;
            
            // Store session in localStorage
            localStorage.setItem('session_token', this.session.token);
            localStorage.setItem('company_name', companyName);
            
            return result;
        } catch (err) {
            console.error('Register error:', err);
            throw err;
        }
    }

    async checkSession() {
        const token = localStorage.getItem('session_token');
        console.log('Session token:', token ? 'present' : 'not found');
        
        if (!token) {
            return false;
        }

        try {
            console.log('Validating session...');
            this.currentUser = await ValidateSession(token);
            console.log('Session valid, user:', this.currentUser);
            return true;
        } catch (err) {
            console.log('Session invalid:', err.message);
            // Session invalid, clear storage
            localStorage.removeItem('session_token');
            localStorage.removeItem('company_name');
            return false;
        }
    }

    async logout() {
        const token = localStorage.getItem('session_token');
        if (token) {
            try {
                await Logout(token);
            } catch (err) {
                console.error('Logout error:', err);
            }
        }
        
        this.currentUser = null;
        this.session = null;
        localStorage.removeItem('session_token');
        localStorage.removeItem('company_name');
    }
}

export const auth = new Auth();