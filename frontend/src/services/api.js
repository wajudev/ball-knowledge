import axios from 'axios';
import {jwtDecode} from "jwt-decode";
import {useAuth} from "../contexts/AuthContext";

const API_URL = 'http://localhost:8081/api';

// create an axios instance
const api = axios.create({
    baseURL: API_URL,
})

// Interceptor to check for token expiration
api.interceptors.request.use(
    config => {
        const token = localStorage.getItem('token');
        if (token) {
            const decoded = jwtDecode(token);
            if (decoded.exp * 1000 < Date.now()) {
                // Token expired
                const { logout } = useAuth();
                logout();
                window.location.href = '/login';
                return Promise.reject(new Error('Token expired'));
            }
            config.headers.Authorization = `Bearer ${token}`;
        }
    return config;
},
error => {
        return Promise.reject(error);
});

export const loginUser = async (userData) => {
    const response = await api.post(`/login`, userData);
    return response.data;
}

export const registerUser = async (userData) => {
    const response = await api.post(`/register`, userData);
    return response.data;
};

export const createPrediction = async (predictionData) => {
    const response = await api.post(`/predictions`, predictionData);
    return response.data;
};

export const getMatches = async () => {
    const response = await api.get(`/matches`);
    return response.data;
};

export const getLeaderboard = async () => {
    const response = await api.get(`/leaderboard`);
    return response.data;
}

export const getProfile = async () => {
    const response = await api.get(`/profile`);
    return response.data;
}
