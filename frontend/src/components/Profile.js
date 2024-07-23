import React, { useEffect, useState } from 'react';
import { getProfile } from '../services/api';

const Profile = () => {
    const [profile, setProfile] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        const fetchProfile = async () => {
            try {
                const data = await getProfile();
                setProfile(data.data); // Adjust based on how your API returns data
                setLoading(false);
            } catch (error) {
                setError('Error fetching profile');
                setLoading(false);
            }
        };

        fetchProfile();
    }, []);

    if (loading) return <p>Loading...</p>;
    if (error) return <p>{error}</p>;

    return (
        <div>
            <h1>Profile</h1>
            {profile && (
                <div>
                    <p><strong>Username:</strong> {profile.username}</p>
                    <p><strong>Email:</strong> {profile.email}</p>
                </div>
            )}
        </div>
    );
};

export default Profile;
