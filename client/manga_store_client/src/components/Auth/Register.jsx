// Register.js
import React, { useState } from 'react';
import axios from '../../axios';
import { useNavigate } from 'react-router-dom';
import './Form.css';

const Register = () => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const navigate = useNavigate();

    const handleRegister = async (e) => {
        e.preventDefault();
        axios.post('/auth/register', { email, password }).then((_res) => {
            navigate("/login");
        });
    };

    return (
        <div className="form-container">
            <div className="form-box">
                <h2>Register</h2>
                <form onSubmit={handleRegister}>
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
                    <button type="submit" className="form-button">Register</button>
                </form>
            </div>
        </div>
    );
};

export default Register;
