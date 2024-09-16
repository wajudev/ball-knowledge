import React, { useState, useEffect } from 'react';
import { createPrediction, getMatches, getPrediction } from '../services/api';

const Prediction = () => {
    const [matches, setMatches] = useState([]);
    const [selectedMatchId, setSelectedMatchId] = useState('');
    const [predictedScoreHome, setPredictedScoreHome] = useState('');
    const [predictedScoreAway, setPredictedScoreAway] = useState('');
    const [message, setMessage] = useState('');
    const [existingPrediction, setExistingPrediction] = useState(null);

    useEffect(() => {
        const fetchMatches = async () => {
            try {
                const response = await getMatches();
                setMatches(response.data || []);  // Ensure that matches is an empty array if no data is returned
            } catch (error) {
                console.error('Error fetching matches:', error);
            }
        };

        fetchMatches().then(r => console.log(r));
    }, []);

    useEffect(() => {
        const fetchPrediction = async () => {
            if (selectedMatchId) {
                try {
                    const response = await getPrediction(selectedMatchId);
                    setExistingPrediction(response.data || null);  // Ensure that existingPrediction is null if no data is returned
                } catch (error) {
                    console.error('Error fetching prediction:', error);
                }
            }
        };
        fetchPrediction().then(r => console.log(r));
    }, [selectedMatchId]);

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
                {matches.length > 0 ? (
                    matches.map((match) => (
                        <option key={match.id} value={match.id}>
                            {match.home_team} vs {match.away_team} - {match.date}
                        </option>
                    ))
                ) : (
                    <option disabled>No matches available</option>
                )}
            </select>
            {existingPrediction ? (
                <p>You have already made a prediction for this match: {existingPrediction.predicted_score_home} - {existingPrediction.predicted_score_away}</p>
            ) : (
                <>
                    <input
                        type="number"
                        value={predictedScoreHome}
                        onChange={(e) => setPredictedScoreHome(e.target.value)}
                        placeholder="Predicted Score Home"
                        required
                    />
                    <input
                        type="number"
                        value={predictedScoreAway}
                        onChange={(e) => setPredictedScoreAway(e.target.value)}
                        placeholder="Predicted Score Away"
                        required
                    />
                    <button type="submit">Submit Prediction</button>
                </>
            )}
            {message && <p>{message}</p>}
        </form>
    );
};

export default Prediction;
