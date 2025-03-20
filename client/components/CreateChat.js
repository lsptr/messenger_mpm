import React, { useState } from 'react';
import { Stack, Typography, TextField, Button, MenuItem, Select, FormControl, InputLabel } from '@mui/material';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';

function CreateChat() {
    const [chatName, setChatName] = useState('');
    const [algorithm, setAlgorithm] = useState('RC5'); // По умолчанию выбран RC5
    const [username, setUsername] = useState('');
    const [error, setError] = useState('');
    const navigate = useNavigate();

    const handleCreateChat = async () => {
        if (!chatName || !algorithm || !username) {
            setError('Заполните все поля');
            return;
        }

        try {
            const token = localStorage.getItem('token');
            await axios.post(
                'http://localhost:8080/chats',
                {
                    name: chatName,
                    algorithm: algorithm,
                    username: username, // Имя пользователя для создания чата
                },
                {
                    headers: {
                        Authorization: `Bearer ${token}`,
                    },
                }
            );

            // Перенаправляем на страницу чатов после создания
            navigate('/chats');
        } catch (err) {
            setError('Не удалось создать чат');
            console.error('Ошибка при создании чата:', err);
        }
    };

    return (
        <Stack mt={4} alignItems="center" spacing={2} width="100%">
            <Typography variant="h4" textAlign="center">
                Создать новый чат
            </Typography>

            {error && (
                <Typography color="error" textAlign="center">
                    {error}
                </Typography>
            )}

            <Stack width="300px" spacing={2}>
                {/* Поле "Название чата" */}
                <TextField
                    fullWidth
                    label="Название чата"
                    value={chatName}
                    onChange={(e) => setChatName(e.target.value)}
                />

                {/* Поле "Алгоритм шифрования" (выпадающий список) */}
                <FormControl fullWidth>
                    <InputLabel>Алгоритм шифрования</InputLabel>
                    <Select
                        value={algorithm}
                        onChange={(e) => setAlgorithm(e.target.value)}
                        label="Алгоритм шифрования"
                    >
                        <MenuItem value="RC5">RC5</MenuItem>
                        <MenuItem value="Serpent">Serpent</MenuItem>
                    </Select>
                </FormControl>

                {/* Поле "Имя пользователя" */}
                <TextField
                    fullWidth
                    label="Имя пользователя"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                />

                {/* Кнопка "Создать" */}
                <Button variant="contained" color='success' onClick={handleCreateChat}>
                    Создать
                </Button>
                <Button variant="contained" onClick={() => navigate('/chats')}>
                    Назад
                </Button>
            </Stack>
        </Stack>
    );
}

export default CreateChat;