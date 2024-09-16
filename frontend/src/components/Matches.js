import React, { useState, useEffect } from 'react';
import { getMatches } from '../services/api';

const Matches = () => {
    const [matches, setMatches] = useState([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchMatches = async () => {
            try {
                const response = await getMatches();
                setMatches(response.data);
                setLoading(false);
            } catch (error) {
                console.error('Error fetching matches:', error);
                setLoading(false);
            }
        };

        fetchMatches();
    }, []);

    if (loading) return <p>Loading...</p>;

    return (
        <div>
            <h2>Matches</h2>
            <ul>
                {matches.map(match => (
                    <li key={match.id}>
                        {match.home_team} vs {match.away_team} on {new Date(match.date).toLocaleDateString()}
                    </li>
                ))}
            </ul>
        </div>
    );
};

export default Matches;
