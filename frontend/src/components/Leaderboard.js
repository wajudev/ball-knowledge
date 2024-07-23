import React, { useEffect, useState } from 'react';
import { getLeaderboard } from '../services/api'; // Import your API function

const Leaderboard = () => {
    const [leaderboard, setLeaderboard] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        const fetchLeaderboard = async () => {
            try {
                const response = await getLeaderboard();
                setLeaderboard(response.data);
                setLoading(false);
            } catch (error) {
                setError('Error fetching leaderboard');
                setLoading(false);
            }
        };

        fetchLeaderboard();
    }, []);

    if (loading) return <p>Loading...</p>;
    if (error) return <p>{error}</p>;

    return (
        <div>
            <h1>Leaderboard</h1>
            <table>
                <thead>
                <tr>
                    <th>Rank</th>
                    <th>Username</th>
                    <th>Points</th>
                </tr>
                </thead>
                <tbody>
                {leaderboard.map((player, index) => (
                    <tr key={player.user_id}>
                        <td>{index + 1}</td>
                        <td>{player.username}</td>
                        <td>{player.points}</td>
                    </tr>
                ))}
                </tbody>
            </table>
        </div>
    );
};

export default Leaderboard;
