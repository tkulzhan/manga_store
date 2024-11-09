import React, { useEffect, useState } from 'react';
import axios from '../../axios';
import "./Profile.css"
import { useNavigate } from 'react-router-dom';

const Profile = () => {
    const [user, setUser] = useState(null);
    const [purchaseDetails, setPurchaseDetails] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    const navigate = useNavigate();

    useEffect(() => {
        const fetchUserData = async () => {
            try {
                const userResponse = await axios.get('/user');
                setUser(userResponse.data);

                const purchasePromises = userResponse.data.purchaseHistory.map(async (purchase) => {
                    const mangaResponse = await axios.get(`/manga/${purchase.mangaId}`);
                    return {
                        ...purchase,
                        ...mangaResponse.data,
                    };
                });

                const detailedPurchases = await Promise.all(purchasePromises);
                setPurchaseDetails(detailedPurchases);
            } catch (err) {
                setError(err.message);
            } finally {
                setLoading(false);
            }
        };

        fetchUserData();
    }, []);

    const handleMangaClick = (mangaId) => {
        navigate(`/manga/${mangaId}`);
    };

    if (loading) return <div>Loading...</div>;
    if (error) return <div>Error: {error}</div>;

    return (
        <div className="profile-container">
            <h1>User Profile</h1>
            {user && (
                <div className="user-info">
                    <p><strong>Email:</strong> {user.email}</p>
                </div>
            )}
            <h2>Purchase History</h2>
            <div className="purchase-history">
                {purchaseDetails.length > 0 ? (
                    purchaseDetails.map((purchase) => (
                        <div key={purchase.mangaId} className="purchase-card" onClick={handleMangaClick(purchase.mangaId)}>
                            <img src={purchase.imageUrl} alt={purchase.title} className="manga-image" />
                            <div className="card-content">
                                <h3 className="manga-title">{purchase.title}</h3>
                                <p>Price: ${purchase.price}</p>
                                <p>Purchased on: {purchase.purchaseDate}</p>
                            </div>
                        </div>
                    ))
                ) : (
                    <p>No purchases found.</p>
                )}
            </div>
        </div>
    );
};

export default Profile;
