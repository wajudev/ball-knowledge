import { useState, useEffect } from 'react';
import { predictionService } from '../services/prediction.service';
import { useAuth } from './useAuth';

export const usePredictions = () => {
    const [predictions, setPredictions] = useState({});
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const { user, token } = useAuth();

    const fetchUserPredictions = async () => {
        if (!user || !token) return;

        try {
            setLoading(true);
            const data = await predictionService.getUserPredictions();
            const predMap = {};
            (data.predictions || []).forEach(pred => {
                predMap[pred.match_id] = pred;
            });
            setPredictions(predMap);
            setError(null);
        } catch (err) {
            setError(err.message);
            console.error('Predictions fetch error:', err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (user && token) {
            fetchUserPredictions();
        }
    }, [user, token]);

    const addPrediction = (matchId, prediction) => {
        setPredictions(prev => ({
            ...prev,
            [matchId]: prediction
        }));
    };

    const refetch = () => {
        fetchUserPredictions();
    };

    return { predictions, loading, error, addPrediction, refetch };
};