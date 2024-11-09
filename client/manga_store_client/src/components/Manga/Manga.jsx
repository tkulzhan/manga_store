// Manga.js
import React, { useEffect, useState } from 'react';
import axios from '../../axios';
import { useParams } from 'react-router-dom';

const Manga = () => {
    const { mangaId } = useParams();
    const [manga, setManga] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [rating, setRating] = useState(0); // State to store user rating

    useEffect(() => {
        const fetchMangaDetails = async () => {
            try {
                const response = await axios.get(`/manga/${mangaId}`);
                setManga(response.data);
            } catch (err) {
                setError(err.message);
            } finally {
                setLoading(false);
            }
        };

        fetchMangaDetails();
    }, [mangaId]);

    const handlePurchase = async () => {
        try {
            const response = await axios.post('/manga/purchase', { mangaId });
            alert(response.data.message);
        } catch (error) {
            alert(`Error purchasing manga: ${error.response.data.error}`);
        }
    };

    const handleRate = async () => {
        try {
            const response = await axios.post(`/manga/${mangaId}/rate`, { score: rating });
            alert(response.data.message);
        } catch (error) {
            alert(`Error rating manga: ${error.response.data.error}`);
        }
    };

    if (loading) {
        return <div>Loading...</div>;
    }

    if (error) {
        return <div>Error fetching manga details: {error}</div>;
    }

    const styles = {
        container: {
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            padding: '2rem',
            backgroundColor: '#f9fafb',
        },
        image: {
            width: '300px',
            height: 'auto',
            borderRadius: '8px',
            boxShadow: '0 4px 8px rgba(0, 0, 0, 0.1)',
            objectFit: 'contain',
        },
        title: {
            fontSize: '2rem',
            margin: '1rem 0',
            color: '#1f2937',
        },
        price: {
            fontSize: '1.5rem',
            color: '#3b82f6',
        },
        description: {
            margin: '1rem 0',
            textAlign: 'center',
            color: '#374151',
        },
        author: {
            margin: '0.5rem 0',
            color: '#6b7280',
        },
        genres: {
            margin: '0.5rem 0',
            color: '#6b7280',
        },
        rating: {
            margin: '0.5rem 0',
            color: '#6b7280',
        },
        button: {
            padding: '0.8rem 1.5rem',
            backgroundColor: '#3b82f6',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            transition: 'background-color 0.3s',
            fontSize: '1rem',
            margin: '0.5rem',
        },
        buttonHover: {
            backgroundColor: '#2563eb',
        },
        ratingContainer: {
            marginTop: '1.5rem',
            textAlign: 'center',
        },
        ratingInput: {
            width: '50px',
            marginRight: '1rem',
            border: '1px solid #d1d5db',
            borderRadius: '4px',
            padding: '0.5rem',
            textAlign: 'center',
        },
    };

    return (
        <div style={styles.container}>
            <img src={manga.imageUrl} alt={manga.title} style={styles.image} />
            <h1 style={styles.title}>{manga.title}</h1>
            <p style={styles.price}>Price: ${manga.price}</p>
            <p style={styles.description}>Description: {manga.description}</p>
            <p style={styles.author}>Author: {manga.author}</p>
            <p style={styles.genres}>Genres: {manga.genres.join(', ')}</p>
            <p style={styles.rating}>Rating: {manga.rating}</p>

            <button style={styles.button} onClick={handlePurchase}>Purchase</button>

            <div style={styles.ratingContainer}>
                <h3>Rate this Manga</h3>
                <input
                    type="number"
                    value={rating}
                    onChange={(e) => setRating(Math.min(Math.max(e.target.value, 0), 5))}
                    min="0"
                    max="5"
                    style={styles.ratingInput}
                />
                <button style={styles.button} onClick={handleRate}>Submit Rating</button>
            </div>
        </div>
    );
};

export default Manga;
