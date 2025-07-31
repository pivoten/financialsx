export function LoginForm(onSubmit, companies = [], onSwitchToRegister) {
    return `
        <div class="auth-container">
            <h2>Login to FinancialsX</h2>
            <form id="login-form" class="auth-form">
                <div class="form-group">
                    <label for="company">Company</label>
                    <select id="company" name="company" required>
                        <option value="">Select a company...</option>
                        ${(companies || []).map(c => `<option value="${c.name}">${c.name}</option>`).join('')}
                    </select>
                </div>
                <div class="form-group">
                    <label for="username">Username</label>
                    <input type="text" id="username" name="username" required autocomplete="username">
                </div>
                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required autocomplete="current-password">
                </div>
                <div class="form-error" id="login-error"></div>
                <button type="submit" class="btn btn-primary">Login</button>
                <p class="auth-switch">
                    Don't have an account? 
                    <a href="#" id="switch-to-register">Register</a>
                </p>
            </form>
        </div>
    `;
}

export function RegisterForm(onSubmit, companies = [], onSwitchToLogin) {
    return `
        <div class="auth-container">
            <h2>Register for FinancialsX</h2>
            <form id="register-form" class="auth-form">
                <div class="form-group">
                    <label for="company">Company</label>
                    <select id="company" name="company" required>
                        <option value="">Select a company...</option>
                        ${(companies || []).map(c => `<option value="${c.name}">${c.name}</option>`).join('')}
                    </select>
                </div>
                <div class="form-group">
                    <label for="username">Username</label>
                    <input type="text" id="username" name="username" required autocomplete="username">
                </div>
                <div class="form-group">
                    <label for="email">Email (optional)</label>
                    <input type="email" id="email" name="email" autocomplete="email">
                </div>
                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required autocomplete="new-password">
                </div>
                <div class="form-group">
                    <label for="confirm-password">Confirm Password</label>
                    <input type="password" id="confirm-password" name="confirm-password" required>
                </div>
                <div class="form-error" id="register-error"></div>
                <button type="submit" class="btn btn-primary">Register</button>
                <p class="auth-switch">
                    Already have an account? 
                    <a href="#" id="switch-to-login">Login</a>
                </p>
            </form>
        </div>
    `;
}

export function Dashboard(user) {
    return `
        <div class="dashboard">
            <header class="dashboard-header">
                <h1>FinancialsX Dashboard</h1>
                <div class="user-info">
                    <span>Welcome, ${user.username}</span>
                    <span class="company-name">${user.company_name}</span>
                    <button id="logout-btn" class="btn btn-secondary">Logout</button>
                </div>
            </header>
            <main class="dashboard-content">
                <div class="welcome-message">
                    <h2>Welcome to FinancialsX</h2>
                    <p>Your modern companion to the legacy Visual FoxPro Accounting Manager.</p>
                    <p>Company: <strong>${user.company_name}</strong></p>
                </div>
            </main>
        </div>
    `;
}