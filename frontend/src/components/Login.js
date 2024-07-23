import React, { useState } from "react";
import { loginUser } from "../services/api";
import {useAuth} from "../contexts/AuthContext";


const Login = () => {
    const [usernameOrEmail, setUsernameOrEmail] = useState("");
    const [password, setPassword] = useState("");
    const [message, setMessage] = useState('');
    const { login } = useAuth();

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            const response = await loginUser({ usernameOrEmail, password });
            console.log('Received token:', response.token); // Log the token
            login(response.token);
            setMessage('Login successful!');
            console.log(response);
        } catch (error) {
            setMessage('Login failed!');
            console.error(error.response ? error.response.data : error.message);
        }
    };

    return (
        <div>
            <h2>Login</h2>
            <form onSubmit={handleSubmit}>
                <input
                    type="text"
                    value={usernameOrEmail}
                    onChange={(e) => setUsernameOrEmail(e.target.value)}
                    placeholder="Username or Email"
                    required
                />
                <input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="Password"
                    required
                />
                <button type="submit">Login</button>
            </form>
            {message && <p>{message}</p>}
        </div>
    );
};

export default Login;
