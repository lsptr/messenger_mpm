import React, { useEffect, useState } from 'react';
import axios from 'axios';
import { Stack, Typography, Paper, List, ListItem, ListItemText, Divider, Button, Box, ListItemButton } from '@mui/material';
import { useNavigate, Link, useParams } from 'react-router-dom'; // Добавляем Link
import DeleteForeverIcon from '@mui/icons-material/DeleteForever';
import Messages from './Messages';
import RefreshIcon from '@mui/icons-material/Refresh';

function Chats() {
    const { chatIdParam } = useParams();
    const [chats, setChats] = useState([]);
    const [selectedChat, setSelectedChat] = useState(null);
    const [error, setError] = useState('');
    const navigate = useNavigate();

    const fetchChats = async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await axios.get('http://localhost:8080/chats', {
                headers: {
                    Authorization: `Bearer ${token}`,
                },
            });
            if (!response.data) return
            setChats(response.data);
            if (chatIdParam) {
                const chatId = parseInt(chatIdParam, 10);  // Преобразуем chatIdParam в число
                setSelectedChat({
                    id: chatId,
                    name: response.data.find(c => c.id === chatId).name,
                    algorithm: response.data.find(c => c.id === chatId).algorithm
                });
            } else setSelectedChat(null);
        } catch (err) {
            setError('Не удалось загрузить чаты');
            console.error('Ошибка при загрузке чатов:', err);
        }
    };

    const handleLogout = () => {
        localStorage.removeItem('token');
        navigate('/login');
    };

    const handleCreateChat = () => {
        navigate('/create-chat');
    };

    const handleChatClick = (id, name) => {

        navigate(`/chats/${id}`)
    };

    const handleChatDelete = async (id) => {
        try {
            const token = localStorage.getItem('token');
            await axios.delete(`http://localhost:8080/chats/${id}`, {
                headers: {
                    Authorization: `Bearer ${token}`,
                },
            });
            fetchChats();
            console.log(id === parseInt(chatIdParam, 10));
            if (id === parseInt(chatIdParam, 10)) {
                navigate("/chats/")
            }
        } catch (err) {
            console.error('Ошибка при удалении чата:', err);
        }
    }
    useEffect(() => {
        const token = localStorage.getItem('token');
        if (!token) {
            navigate('/login');
            return;
        }
        fetchChats();

    }, [navigate]);

    return (
        <Stack mt={4} spacing={2} width="100%">
            <Box position="absolute" top={16} right={16}>
                <Button variant="contained" color="error" onClick={handleLogout}>
                    Выход
                </Button>
            </Box>

            <Stack ml={'20px !important'} direction={'row'} spacing={4}>
                <Stack alignItems={'center'}>

                    {error && (
                        <Typography color="error" textAlign="center">
                            {error}
                        </Typography>
                    )}
                    <Typography fontSize={24} fontWeight={'bold'} ml={'10px'}>Чаты</Typography>
                    <Paper
                        elevation={3}
                        sx={{ width: '100%', padding: '16px', textAlign: 'center' }}
                    >
                        {chats ? (
                            <List
                                sx={{
                                    maxHeight: '750px', // Максимальная высота контейнера
                                    overflowY: 'auto', // Включаем вертикальный скроллинг, если контент выходит за пределы
                                    paddingBottom: '8px', // Небольшой отступ снизу, чтобы не прилипали элементы
                                }}
                            >
                                {chats.map((chat) => (
                                    <ListItem
                                        key={chat.id}
                                        button="true"
                                    >
                                        <Stack
                                            direction={'row'}
                                            alignItems={'center'}
                                            width={'100%'}
                                            height={'100%'}
                                        >
                                            <ListItemText
                                                sx={{ width: '80%', cursor: 'pointer' }}
                                                onClick={() => handleChatClick(chat.id, chat.name)}
                                                button="true"
                                                primary={chat.name}
                                                secondary={
                                                    <Typography color='textDisabled' variant="body2" component="span">
                                                        Алгоритм: {chat.algorithm}
                                                        <br />
                                                        Участник: {chat.user2_name}
                                                    </Typography>
                                                }
                                            />
                                            <ListItemButton
                                                onClick={() => handleChatDelete(chat.id)}
                                            >
                                                <DeleteForeverIcon sx={{ color: 'red', fontSize: 30 }} />
                                            </ListItemButton>
                                        </Stack>
                                    </ListItem>
                                ))}
                            </List>) : (<></>)}
                    </Paper>

                    <Stack direction={'row'} spacing={1} mt={'10px'} alignItems={'center'}>
                        <Button variant="contained" color="primary" onClick={handleCreateChat}>
                            Создать
                        </Button>
                        <Button variant='contained' onClick={fetchChats}>
                            <RefreshIcon />
                        </Button>
                    </Stack>

                </Stack>
                <Stack width={'100%'}>
                    {selectedChat ? (
                        <Messages
                            chatName={selectedChat.name}
                            chatId={selectedChat.id}
                            algorithm={selectedChat.algorithm}
                        />) : <></>}
                </Stack>
            </Stack>

        </Stack>
    );
}

export default Chats;