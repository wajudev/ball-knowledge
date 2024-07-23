import React from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const LandingPage = () => {
    const { user } = useAuth();

    return (
        <div style={{ textAlign: 'center', marginTop: '50px' }}>
            <h1>Welcome to the Football Prediction App</h1>
            <p>Predict the scores of football matches and compete with others!</p>
            <div>
                {!user ? (
                    <>
                        <Link to="/register">
                            <button style={{ margin: '10px', padding: '10px 20px', fontSize: '16px' }}>Register</button>
                        </Link>
                        <Link to="/login">
                            <button style={{ margin: '10px', padding: '10px 20px', fontSize: '16px' }}>Login</button>
                        </Link>
                    </>
                ) : (
                    <Link to="/predict">
                        <button style={{ margin: '10px', padding: '10px 20px', fontSize: '16px' }}>Make a Prediction</button>
                    </Link>
                )}
            </div>
        </div>
    );
};

export default LandingPage;
