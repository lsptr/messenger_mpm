import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';
import { Typography, TextField, Button, CircularProgress, List, ListItem, Stack, Box, Divider } from '@mui/material';
import SendIcon from '@mui/icons-material/Send';
import ArrowBackIosNewIcon from '@mui/icons-material/ArrowBackIosNew';
import AttachFileIcon from '@mui/icons-material/AttachFile';

function Messages({chatName, chatId, algorithm}) {
    const navigate = useNavigate();

    const username = localStorage.getItem('username');
    const [messages, setMessages] = useState([]);
    const [newMessage, setNewMessage] = useState('');
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [ws, setWs] = useState(null);
    const [file, setFile] = useState(null);

    const userId = localStorage.getItem('user_id');
    const sharedSecretKeyKey = `${userId}_${chatId}_shared_secret`;
    const localStorageKey = `chat_${chatId}_messages`;

    // Проверка наличия общего секретного ключа
    useEffect(() => {
        const sharedSecret = localStorage.getItem(sharedSecretKeyKey);
        if (!sharedSecret) {
            navigate(`/chats/${chatId}/keys`);
            return;
        }
        loadMessages();
    }, [chatId]);

    useEffect(() => {
        const socket = new WebSocket(`ws://localhost:8080/ws?token=${localStorage.getItem('token')}&&chat_id=${chatId}`);
        socket.onmessage = (event) => {
            const message = JSON.parse(event.data).msg;
            if (message.senderName === username || JSON.parse(event.data).chatId !== chatId) return;
            message.source = "ws";
            setMessages((prev) => [...prev, message]);
        };
        setWs(socket);
        return () => socket.close();
    }, [chatId])

    const handleKeyDown = (e) => {
        if (e.key === "Enter" && newMessage.trim() !== "") {
            sendMessage();
        }
    };

    const handleAddFile = (e) => {
        const selectedFile = e.target.files[0];
        if (selectedFile) {
            if (selectedFile.size > 10485760) {
                alert("Файл слишком большой. Максимальный размер — 10 MB.");
                return;
            }
            const reader = new FileReader();
            reader.onload = (event) => {
                const base64Image = event.target.result;
                setFile({ data: base64Image, name: selectedFile.name });
            };
            reader.readAsDataURL(selectedFile);
        }
    };

    const formatDate = (dateString) => {
        const date = new Date(dateString);
        return date.toLocaleString('ru-RU');
    };

    const setLocalMessages = async (newLocalMessage) => {
        // Добавляем новое сообщение в локальное хранилище
        const localMessages = messages;
        const updatedLocalMessages = [...localMessages, newLocalMessage];
        console.log('updatedLocalMessages',updatedLocalMessages);
        // Сортируем все сообщения после добавления нового
        updatedLocalMessages.sort((a, b) => {
            const dateA = a.createdAt ? new Date(a.createdAt) : new Date(a.timestamp);
            const dateB = b.createdAt ? new Date(b.createdAt) : new Date(b.timestamp);
            return dateA - dateB;
        });

        // Сохраняем обновлённые и отсортированные сообщения локально
        localStorage.setItem(localStorageKey, JSON.stringify(updatedLocalMessages));

        // Обновляем состояние
        setMessages(updatedLocalMessages);
    }

    // Загрузка сообщений
    const loadMessages = async () => {
        try {
            // Загружаем локальные сообщения
            const localMessages = JSON.parse(localStorage.getItem(localStorageKey)) || [];
            // Загружаем новые сообщения с сервера
            console.log(localStorage.getItem('token'));
            const response = await axios.get(`http://localhost:8080/chats/${chatId}/messages`, {
                headers: {
                    Authorization: `Bearer ${localStorage.getItem('token')}`,
                    'X-Client-Type': 'JS',
                    'X-Key': localStorage.getItem(`${userId}_${chatId}_shared_secret`).padStart(16, "0"),
                    'X-Algorithm': algorithm,
                },
            });

            console.log('Ответ от сервера:', response.data);
            const newMessages = response.data.messages || [];
            console.log('Новые сообщения с сервера:', newMessages);
            // Преобразуем сообщения с сервера к единому формату
            const formattedMessages = newMessages.map(msg => ({
                ...msg,
                timestamp: formatDate(msg.createdAt),
                createdAt: new Date(msg.createdAt),
            }));

            // Фильтруем новые сообщения, чтобы не дублировать уже сохранённые
            const existingMessageIds = new Set(localMessages.map(msg => msg.id));

            const uniqueNewMessages = formattedMessages.filter(msg => !existingMessageIds.has(msg.id));

            // Объединяем локальные и новые сообщения
            let updatedMessages = [...localMessages, ...uniqueNewMessages];

            // Сортируем все сообщения по времени (от старых к новым)
            updatedMessages.sort((a, b) => {
                const dateA = a.createdAt ? new Date(a.createdAt) : new Date(a.timestamp);
                const dateB = b.createdAt ? new Date(b.createdAt) : new Date(b.timestamp);
                return dateA - dateB;
            });
            updatedMessages = updatedMessages.filter(m => m.source !== "ws")
            // Сохраняем обновлённые и отсортированные сообщения локально
            localStorage.setItem(localStorageKey, JSON.stringify(updatedMessages));

            // Обновляем состояние
            setMessages(updatedMessages);
            setLoading(false);
        } catch (err) {
            setError('Ошибка загрузки сообщений');
            console.error('Ошибка загрузки сообщений:', err);
        }
    };

    // Отправка сообщения
    const sendMessage = async () => {
        if (!newMessage && !file) return;
        try {
            const senderName = "Вы"; // Имя отправителя (в данном случае — текущий пользователь)
            const timestamp = new Date().toLocaleString('ru-RU');

            const newLocalMessage = {
                id: Date.now(),
                senderName: senderName,
                timestamp: timestamp,
                createdAt: new Date().toISOString(),
                message: newMessage,
                file: file
            };

            if (ws) {
                ws.send(JSON.stringify({
                    msg: {
                        id: Date.now(),
                        senderName: username,
                        timestamp: timestamp,
                        createdAt: new Date().toISOString(),
                        message: newMessage,
                        file: file
                    },
                    chatId: chatId
                }));
                setNewMessage("");
                setFile(null);
            }

            setLocalMessages(newLocalMessage);
            // Отправляем сообщение на сервер
            const response = await axios.post(`http://localhost:8080/chats/${chatId}/messages`, {
                content: newMessage, 
                file: file?  file.data : '',
                file_name: file ? file.name : '',
                algorithm: algorithm,
                key: localStorage.getItem(`${userId}_${chatId}_shared_secret`).padStart(16, "0")
            }, {
                headers: {
                    Authorization: `Bearer ${localStorage.getItem('token')}`,
                },
            });

            if (response.status === 200) {
                setNewMessage('');
                // Загружаем новые сообщения после отправки
                //await loadMessages();
            }
        } catch (err) {
            setError('Ошибка отправки сообщения');
            console.error('Ошибка отправки сообщения:', err);
        }
    };

    const handleBackClick = () => {
        navigate('/chats')
    }

    if (loading) {
        return (
            <div style={{ textAlign: 'center', marginTop: '20%' }}>
                <CircularProgress />
                <Typography variant="h5" mt={2}>Загрузка сообщений...</Typography>
            </div>
        );
    }

    return (
        <Stack sx={{ height: '95vh', display: 'flex', flexDirection: 'column' }}>
            <Stack direction={'row'}>
                <Button onClick={handleBackClick}><ArrowBackIosNewIcon /></Button>
                <Stack direction={'row'} spacing={1}>
                    <Typography fontSize={40}>Чат</Typography>
                    <Typography fontSize={40} color='#1565c0'>{chatName}</Typography>
                </Stack>
            </Stack>
            <Divider></Divider>
            <List sx={{ flexGrow: 1, overflowY: 'auto' }}>
                {messages.map((msg, index) => {
                    const isMine = msg.senderName === "Вы";
                    //console.log('message', msg);
                    return (
                        <ListItem
                            key={index}
                            sx={{
                                display: 'flex',
                                justifyContent: isMine ? 'flex-end' : 'flex-start',

                            }}>
                            <Box
                                sx={{
                                    border: '1px solid #ccc',
                                    borderRadius: '5px',
                                    padding: '15px',
                                    borderColor: isMine ? '#1565c0' : '#7a7a7a',
                                    backgroundColor: isMine ? '#e3f2fd' : '#ffffff',
                                    alignSelf: isMine ? 'flex-end' : 'flex-start',
                                }}>
                                <Stack direction={'row'} spacing={2}>
                                    <Typography fontWeight={'bold'} sx={{ color: '#1565c0' }}>{msg.senderName}</Typography>
                                    <Box sx={{ flexGrow: 1 }} />
                                    <Typography>{msg.timestamp.slice(0, -3)}</Typography>
                                </Stack>
                                <Typography sx={{ mt: '5px' }}>{`${msg.message}`}</Typography>
                                {msg.file && (
                                    msg.file.data !== "" && (
                                    <Box sx={{ mt: 2 }}>
                                        <img
                                            src={msg.file.data}
                                            style={{
                                                maxWidth: '100%',
                                                maxHeight: '400px',
                                                objectFit: 'contain',
                                                borderRadius: '5px',
                                                border: '1px solid #ccc',
                                                marginTop: '10px',
                                            }}
                                        />
                                    </Box>)
                                )}
                            </Box>
                        </ListItem>)
                })}
            </List>

            <Stack>
                <Stack
                    direction="row"
                    alignItems="center"
                    spacing={1}
                    sx={{
                        position: 'sticky',
                        bottom: 0,
                        backgroundColor: 'white',
                        padding: '10px',
                        borderTop: '1px solid #ccc',
                    }}
                >
                    <input
                        type="file"
                        onChange={handleAddFile}
                        style={{ display: "none" }}
                        id="file-upload"
                        accept=".jpg,.jpeg,.png"
                    />
                    <label htmlFor="file-upload" style={{ height: "100%" }}>
                        <Button
                            variant="contained"
                            sx={{ height: "100%" }}
                            component="span"
                        >
                            <AttachFileIcon />
                        </Button>
                    </label>

                    <TextField
                        fullWidth
                        variant="outlined"
                        placeholder="Введите сообщение"
                        value={newMessage}
                        onChange={(e) => setNewMessage(e.target.value)}
                        onKeyDown={handleKeyDown}
                    />
                    <Button sx={{ height: '100%' }} variant="contained" color="primary" onClick={sendMessage}>
                        <SendIcon />
                    </Button>
                </Stack>

                {file && (
                    <Stack sx={{ ml: '20px' }} direction={'row'}>
                        <Typography>Выбран файл: </Typography>
                        <Typography color='#1565c0'>{file.name}</Typography>
                    </Stack>)}

            </Stack>


            {error && (
                <Typography color="error" style={{ marginTop: '10px' }}>
                    {error}
                </Typography>
            )}
        </Stack>
    );
}

export default Messages;