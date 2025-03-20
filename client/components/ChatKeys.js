/* eslint-disable no-undef */
import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { useParams, useNavigate } from 'react-router-dom';
import { Typography, CircularProgress } from '@mui/material';
import { generatePrivateKey, generatePublicKey, computeSharedSecret } from './diffieHellman';

function ChatKeys() {
    const { chatId } = useParams();
    const navigate = useNavigate();
    const [loading, setLoading] = useState(true);
    const [status, setStatus] = useState('Проверка ключей...');

    const userId = localStorage.getItem('user_id');
    const privateKeyKey = `${userId}_${chatId}_private`;
    const publicKeyKey = `${userId}_${chatId}_public`;
    const sharedSecretKeyKey = `${userId}_${chatId}_shared_secret`;

    const generateAndSaveKeys = async () => {
        try {
            // Параметры для Diffie-Hellman
            const p = 23; // Простое число 
            const g = 5;  // Генератор

            // Генерация приватного и публичного ключей
            const privateKey = generatePrivateKey(p);
            const publicKey = generatePublicKey(g, privateKey, p);

            // Сохраняем ключи в localStorage
            localStorage.setItem(privateKeyKey, privateKey.toString());
            localStorage.setItem(publicKeyKey, publicKey.toString());

            // Логирование перед отправкой
            console.log('Публичный ключ сгенерирован:', publicKey.toString());
            console.log('Отправка публичного ключа на сервер...');

            // Отправляем публичный ключ на сервер
            const response = await axios.post(`http://localhost:8080/chats/${chatId}/keys`, {
                public_key: publicKey.toString(),
            }, {
                headers: {
                    Authorization: `Bearer ${localStorage.getItem('token')}`,
                },
            });

            // Логирование ответа от сервера
            console.log('Ответ от сервера:', response.data);

            setStatus('Ключи сгенерированы и отправлены на сервер. Ожидание собеседника...');
        } catch (err) {
            // Логирование ошибки
            console.error('Ошибка генерации или отправки ключей:', err);
            setStatus('Ошибка генерации или отправки ключей');
        }
    };

    const fetchPublicKey = async () => {
        try {
            // Логирование перед запросом публичного ключа
            console.log('Запрос публичного ключа собеседника...');

            // Получаем публичный ключ собеседника с сервера
            const response = await axios.get(`http://localhost:8080/chats/${chatId}/keys`, {
                headers: {
                    Authorization: `Bearer ${localStorage.getItem('token')}`,
                },
            });

            // Логирование ответа от сервера
            console.log('Публичный ключ собеседника:', response.data.public_key);

            const otherPublicKey = BigInt(response.data.public_key);
            const privateKey = BigInt(localStorage.getItem(privateKeyKey));
            const p = 23; // Используем то же простое число

            // Вычисляем общий секретный ключ
            const sharedSecret = computeSharedSecret(otherPublicKey, privateKey, p);
            localStorage.setItem(sharedSecretKeyKey, sharedSecret.toString());

            // Логирование успешного вычисления общего ключа
            console.log('Общий секретный ключ вычислен:', sharedSecret.toString());

            // Перенаправляем на страницу сообщений
            navigate(`/chats/${chatId}/`);
        } catch (err) {
            // Логирование ошибки
            console.error('Ошибка при получении публичного ключа:', err);
            setStatus('Ожидание публичного ключа от собеседника...');
        }
    };

    useEffect(() => {
        const sharedSecret = localStorage.getItem(sharedSecretKeyKey);
        if (sharedSecret) {
            navigate(`/chats/${chatId}/`);
            return;
        }

        const publicKey = localStorage.getItem(publicKeyKey);
        if (!publicKey) {
            generateAndSaveKeys();
        }

        // Периодически проверяем публичный ключ собеседника
        const intervalId = setInterval(fetchPublicKey, 10000);
        return () => clearInterval(intervalId);
    }, [chatId, navigate]);

    return (
        <div style={{ textAlign: 'center', marginTop: '20%' }}>
            <CircularProgress />
            <Typography variant="h5" mt={2}>{status}</Typography>
        </div>
    );
}

export default ChatKeys;