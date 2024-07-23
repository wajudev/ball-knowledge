import React, { useState, useEffect } from 'react';
import { createPrediction, getMatches } from '../services/api';

const Prediction = () => {
    const [matches, setMatches] = useState([]);
    const [selectedMatchId, setSelectedMatchId] = useState('');
    const [predictedScoreHome, setPredictedScoreHome] = useState('');
    const [predictedScoreAway, setPredictedScoreAway] = useState('');
    const [message, setMessage] = useState('');

    useEffect(() => {
        const fetchMatches = async () => {
            try {
                const response = await getMatches();
                setMatches(response.data);
            } catch (error) {
                console.error('Error fetching matches:', error);
            }
        };

        fetchMatches();
    }, []);

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            const response = await createPrediction({
                match_id: selectedMatchId,
                predicted_score_home: parseInt(predictedScoreHome),
                predicted_score_away: parseInt(predictedScoreAway),
            });
            setMessage('Prediction added successfully!');
            console.log(response);
        } catch (error) {
            setMessage('Prediction not added!');
            console.error(error.response ? error.response.data : error.message);
        }
    };

    return (
        <form onSubmit={handleSubmit}>
            <select
                value={selectedMatchId}
                onChange={(e) => setSelectedMatchId(e.target.value)}
                required
            >
                <option value="" disabled>Select a match</option>
                {matches.map((match) => (
                    <option key={match.id} value={match.id}>
                        {match.home_team} vs {match.away_team} - {match.date}
                    </option>
                ))}
            </select>
            <input
                type="number"
                value={predictedScoreHome}
                onChange={(e) => setPredictedScoreHome(e.target.value)}
                placeholder="Predicted Score Team 1"
                required
            />
            <input
                type="number"
                value={predictedScoreAway}
                onChange={(e) => setPredictedScoreAway(e.target.value)}
                placeholder="Predicted Score Team 2"
                required
            />
            <button type="submit">Submit Prediction</button>
            {message && <p>{message}</p>}
        </form>
    );
};

export default Prediction;
