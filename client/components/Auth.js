import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';
import { TextField, Typography } from '@mui/material';
import { useState } from 'react';
import Login from './Login';
import Register from './Register';


function Auth() {

  const [isLogin, setIsLogin] = useState(true);


  return (
    <Stack>
      <Stack m={'10px'} direction={'row'} spacing={2}>
        <Button onClick={() => setIsLogin(true)}>вход</Button>
        <Button onClick={() => setIsLogin(false)}>Регистрация</Button>
      </Stack>
      {isLogin ? <Login /> : <Register />}

    </Stack>

  )
}

export default Auth;
