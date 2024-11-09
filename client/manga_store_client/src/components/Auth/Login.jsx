// Login.js
import React, { useState } from 'react';
import axios from '../../axios';
import { useNavigate } from 'react-router-dom';
import './Form.css';

const Login = () => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const navigate = useNavigate();

    const handleLogin = async (e) => {
        e.preventDefault();
        axios.post('/auth/login', { email, password })
            .then((_res) => {
                navigate('/home');
            });
    };

    return (
        <div className="form-container">
            <div className="form-box">
                <h2>Login</h2>
                <form onSubmit={handleLogin}>
                    <input 
                        type="email" 
                        placeholder="Email" 
                        value={email} 
                        onChange={(e) => setEmail(e.target.value)} 
                        required 
                        className="form-input" 
                    />
                    <input 
                        type="password" 
                        placeholder="Password" 
                        value={password} 
                        onChange={(e) => setPassword(e.target.value)} 
                        required 
                        className="form-input" 
                    />
                    <button type="submit" className="form-button">Login</button>
                </form>
            </div>
        </div>
    );
};

export default Login;
