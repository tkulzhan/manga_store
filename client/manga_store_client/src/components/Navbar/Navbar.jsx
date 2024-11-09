import React from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import './Navbar.css';
import axios from "../../axios"

const Navbar = () => {
    const location = useLocation();
    const navigate = useNavigate();

    const handleLogout = () => {
        axios.post("/auth/logout").then((_res) => {
            navigate("/login");
        })  
    };

    return (
        <nav className="navbar">
            <div className="navbar-logo">
                <Link to="/home">MangaStore</Link>
            </div>
            <ul className="navbar-links">
                <li className={location.pathname === "/home" ? "active" : ""}>
                    <Link to="/home">Home</Link>
                </li>
                <li className={location.pathname === "/search" ? "active" : ""}>
                    <Link to="/search">Search</Link>
                </li>
                <li className={location.pathname === "/profile" ? "active" : ""}>
                    <Link to="/profile">Profile</Link>
                </li>
                <li className={location.pathname === "/login" ? "active" : ""}>
                    <Link to="/login">Login</Link>
                </li>
                <li className={location.pathname === "/register" ? "active" : ""}>
                    <Link to="/register">Register</Link>
                </li>
                <li>
                    <button onClick={handleLogout} className="logout-button">Logout</button>
                </li>
            </ul>
        </nav>
    );
};

export default Navbar;
