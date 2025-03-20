import React, { useState } from 'react';
import axios from 'axios';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';
import { TextField, Typography } from '@mui/material';
import Cookies from 'js-cookie';

function Login() {
    const [login, setLogin] = useState('');
    const [password, setPassword] = useState('');
    const [loginError, setLoginError] = useState(false);
    const [passwordError, setPasswordError] = useState(false);
    const [errorMessage, setErrorMessage] = useState('');

    const handleLoginChange = () => {
        if (login === '') {
            setLoginError(true);
        } else {
            setLoginError(false);
        }
    };

    const handlePasswordChange = () => {
        if (password === '') {
            setPasswordError(true);
        } else {
            setPasswordError(false);
        }
    };

    const handleClick = async () => {
        if (login && password) {
            try {
                const response = await axios.post('http://localhost:8080/auth/login', {
                    username: login,
                    password: password,
                });
    
                // Сохраняем токен в localStorage
                localStorage.setItem('token', response.data.token);
                localStorage.setItem('userID', response.data.userID);
                localStorage.setItem('username', response.data.username);
                console.log('data', response.data);
                // Перенаправляем на страницу чатов
                window.location.href = '/chats';
            } catch (err) {
                setErrorMessage('Ошибка входа. Проверьте логин и пароль.');
                console.error('Login error:', err);
            }
        }
    };

    return (
        <Stack mt={'300px'} alignItems={'center'} spacing={2} width="100%">
            <Typography fontSize={24} textAlign="center">Вход в мессенджер</Typography>

            <Stack width="300px" spacing={2}>
                <TextField
                    fullWidth
                    onChange={(e) => setLogin(e.target.value)}
                    error={loginError}
                    onBlur={handleLoginChange}
                    label="Логин"
                    helperText={loginError ? 'Неправильный логин' : ''}
                />

                <TextField
                    fullWidth
                    onChange={(e) => setPassword(e.target.value)}
                    error={passwordError}
                    onBlur={handlePasswordChange}
                    label="Пароль"
                    helperText={passwordError ? 'Неправильный пароль' : ''}
                    type="password"
                />

                {errorMessage && (
                    <Typography color="error" textAlign="center">
                        {errorMessage}
                    </Typography>
                )}

                <Button fullWidth onClick={handleClick} variant="contained">
                    Войти
                </Button>
            </Stack>
        </Stack>
    );
}

export default Login;