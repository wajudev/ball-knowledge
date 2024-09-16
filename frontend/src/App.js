import React from 'react';
import {Link, Route, Routes} from 'react-router-dom';
import Register from './components/Register';
import Login from './components/Login';
import Prediction from './components/Prediction';
import LandingPage from './components/LandingPage';
import {AuthProvider, useAuth} from './contexts/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';
import Leaderboard from "./components/Leaderboard";
import Profile from "./components/Profile";
import Matches from "./components/Matches";


function App() {
    return (
        <AuthProvider>
            <Header />
            <Routes>
                <Route path="/" element={<LandingPage />} />
                <Route path="/register" element={<Register />} />
                <Route path="/login" element={<Login />} />
                <Route element={<ProtectedRoute />}>
                    <Route path="/predict" element={<Prediction />} />
                    <Route path="/leaderboard" element={<Leaderboard />} />
                    <Route path="/profile" element={<Profile />} />
                    <Route path="/matches/:gameweek" element={<Matches />} />
                </Route>
            </Routes>
        </AuthProvider>
    );
}

const Header = () => {
    const { user, logout } = useAuth();

    return (
        <nav style={{ display: 'flex', justifyContent: 'space-between', padding: '10px' }}>
            <div>
                <Link to="/">Home</Link>
                {user && (
                    <>
                        <Link to="/predict" style={{ marginLeft: '10px' }}>Predict</Link>
                        <Link to="/leaderboard" style={{ marginLeft: '10px' }}>Leaderboard</Link>
                        <Link to="/profile" style={{ marginLeft: '10px' }}>Profile</Link>
                    </>
                )}
            </div>
            <div>
                {user ? (
                    <button onClick={logout}>Logout</button>
                ) : (
                    <Link to="/login">Login</Link>
                )}
            </div>
        </nav>
    );
};

export default App;
