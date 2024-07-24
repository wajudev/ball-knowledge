import React from 'react';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import App from '../App';
import { AuthProvider } from '../contexts/AuthContext';

// Mock the components used in App.js that might cause issues
jest.mock('../components/Register', () => () => <div>Register</div>);
jest.mock('../components/Login', () => () => (
    <div>
        <h1>Login Page</h1>
        <form>
            <input placeholder="Email" />
            <input placeholder="Password" />
            <button type="submit">Login</button>
        </form>
    </div>
));
jest.mock('../components/Prediction', () => () => <div>Prediction</div>);
jest.mock('../components/LandingPage', () => () => <div>Welcome to the Football Prediction App</div>);
jest.mock('../components/Leaderboard', () => () => <div>Leaderboard</div>);
jest.mock('../components/ProtectedRoute', () => ({ children }) => <div>{children}</div>);

describe('App', () => {
    test('renders landing page', () => {
        render(
            <MemoryRouter initialEntries={['/']}>
                <AuthProvider>
                    <App />
                </AuthProvider>
            </MemoryRouter>
        );

        expect(screen.getByText(/Welcome to the Football Prediction App/i)).toBeInTheDocument();
    });

    test('renders login page on /login route', () => {
        render(
            <MemoryRouter initialEntries={['/login']}>
                <AuthProvider>
                    <App />
                </AuthProvider>
            </MemoryRouter>
        );

        expect(screen.getByText(/Login Page/i)).toBeInTheDocument();
        expect(screen.getByPlaceholderText(/Email/i)).toBeInTheDocument();
        expect(screen.getByPlaceholderText(/Password/i)).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Login/i })).toBeInTheDocument();
    });

    test('renders register page on /register route', () => {
        render(
            <MemoryRouter initialEntries={['/register']}>
                <AuthProvider>
                    <App />
                </AuthProvider>
            </MemoryRouter>
        );

        expect(screen.getByText(/Register/i)).toBeInTheDocument();
    });

    // Add more tests as needed
});
