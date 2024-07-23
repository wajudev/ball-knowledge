import React from 'react';
import { useAuth } from '../contexts/AuthContext';

const Logout = () => {
    const { logout } = useAuth();

    const handleLogout = () => {
        logout();
    };

    return (
        <button onClick={handleLogout} style={{ margin: '10px', padding: '10px 20px', fontSize: '16px' }}>
            Logout
        </button>
    );
};

export default Logout;
