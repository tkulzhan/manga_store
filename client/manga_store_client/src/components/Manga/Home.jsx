// Home.js
import React, { useEffect, useState } from 'react';
import axios from '../../axios';
import { useNavigate } from 'react-router-dom';

const Home = () => {
    const [popularManga, setPopularManga] = useState([]);
    const [recommendedManga, setRecommendedManga] = useState([]);
    const [similarTasteManga, setSimilarTasteManga] = useState([]);
    const [newestManga, setNewestManga] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    const navigate = useNavigate();

    useEffect(() => {
        const fetchMangaData = async () => {
            try {
                const [popularResponse, recommendedResponse, similarResponse, newestResponse] = await Promise.all([
                    axios.get('/manga/popular'),
                    axios.get('/user/recs/preferences'),
                    axios.get('/user/recs/similar_users'),
                    axios.get('/manga'),
                ]);

                setPopularManga(popularResponse.data);
                setRecommendedManga(recommendedResponse.data);
                setSimilarTasteManga(similarResponse.data);
                setNewestManga(newestResponse.data);
            } catch (err) {
                setError(err.message);
            } finally {
                setLoading(false);
            }
        };

        fetchMangaData();
    }, [navigate]);

    const handleMangaClick = (mangaId) => {
        navigate(`/manga/${mangaId}`);
    };

    if (loading) {
        return <div className="loading-message">Loading...</div>;
    }

    if (error) {
        return <div className="error-message">Error fetching data: {error}</div>;
    }

    return (
        <div className="home-container">
            <h1>Welcome to the Manga Store</h1>

            <div className="manga-section">
                <h2>Popular Manga</h2>
                <div className="manga-row">
                    {popularManga.map((manga) => (
                        <div
                            key={manga.id}
                            className="manga-item"
                            onClick={() => handleMangaClick(manga.id)}
                        >
                            <img src={manga.imageUrl} alt={manga.title} className="manga-image" />
                            <h3 className="manga-title">{manga.title}</h3>
                            <p>Price: ${manga.price}</p>
                        </div>
                    ))}
                </div>
            </div>

            <div className="manga-section">
                <h2>Personalized Recommendations</h2>
                <div className="manga-row">
                    {recommendedManga.map((manga) => (
                        <div
                            key={manga.id}
                            className="manga-item"
                            onClick={() => handleMangaClick(manga.id)}
                        >
                            <img src={manga.imageUrl} alt={manga.title} className="manga-image" />
                            <h3 className="manga-title">{manga.title}</h3>
                            <p>Price: ${manga.price}</p>
                        </div>
                    ))}
                </div>
            </div>

            <div className="manga-section">
                <h2>Users Similar to You Like</h2>
                <div className="manga-row">
                    {similarTasteManga.map((manga) => (
                        <div
                            key={manga.id}
                            className="manga-item"
                            onClick={() => handleMangaClick(manga.id)}
                        >
                            <img src={manga.imageUrl} alt={manga.title} className="manga-image" />
                            <h3 className="manga-title">{manga.title}</h3>
                            <p>Price: ${manga.price}</p>
                        </div>
                    ))}
                </div>
            </div>

            <div className="manga-section">
                <h2>Newest Manga</h2>
                <div className="manga-row">
                    {newestManga.map((manga) => (
                        <div
                            key={manga.id}
                            className="manga-item"
                            onClick={() => handleMangaClick(manga.id)}
                        >
                            <img src={manga.imageUrl} alt={manga.title} className="manga-image" />
                            <h3 className="manga-title">{manga.title}</h3>
                            <p>Price: ${manga.price}</p>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default Home;
