import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Auth from './components/Auth';
import Chats from './components/Chats';
import CreateChat from './components/CreateChat.js'; 
import ChatKeys from './components/ChatKeys';
import Messages from './components/Messages';

function App() {
    return (
        <Router>
            <Routes>
                <Route path="" element={<Auth />} />
                <Route path="/login" element={<Auth />} />
                <Route path="/chats/:chatIdParam" element={<Chats />} />
                <Route path="/chats/" element={<Chats />} />
                <Route path="/create-chat" element={<CreateChat />} />
                <Route path="/chats/:chatId/keys" element={<ChatKeys />} />
                <Route path="/chats/:chatId/messages" element={<Messages />} />
            </Routes>
        </Router>
    );
}

export default App;