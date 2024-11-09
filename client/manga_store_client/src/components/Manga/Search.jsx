// Search.js
import React, { useState } from 'react';
import axios from '../../axios';

const Search = () => {
    const [query, setQuery] = useState('');
    const [genres, setGenres] = useState([]);
    const [author, setAuthor] = useState('');
    const [limit, setLimit] = useState(10);
    const [results, setResults] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const handleSearch = async () => {
        setLoading(true);
        setError(null);

        try {
            const response = await axios.post('/manga/search', {
                query,
                genres,
                author,
                limit,
            });
            setResults(response.data);
        } catch (err) {
            setError(err.response ? err.response.data.error : 'Error performing search');
        } finally {
            setLoading(false);
        }
    };

    const handleGenreChange = (e) => {
        const value = e.target.value;
        setGenres(value ? value.split(',') : []);
    };

    return (
        <div className="search-container">
            <h2>Search Manga</h2>
            <div className="search-form">
                <label>Query:</label>
                <input
                    type="text"
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                />
                <label>Genres (comma-separated):</label>
                <input
                    type="text"
                    value={genres.join(',')}
                    onChange={handleGenreChange}
                />
                <label>Author:</label>
                <input
                    type="text"
                    value={author}
                    onChange={(e) => setAuthor(e.target.value)}
                />
                <label>Limit:</label>
                <input
                    type="number"
                    value={limit}
                    onChange={(e) => setLimit(parseInt(e.target.value) || 10)}
                    min="1"
                />
                <button onClick={handleSearch} className="search-button">Search</button>
            </div>

            {loading && <p className="results-message">Loading...</p>}
            {error && <p className="results-message">{error}</p>}

            <div className="results-container">
                {results.map((manga) => (
                    <div
                        key={manga.id}
                        className="manga-item"
                    >
                        <img src={manga.imageUrl} alt={manga.title} className="manga-image" />
                        <h3 className="manga-title">{manga.title}</h3>
                        <p>Price: ${manga.price}</p>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default Search;
