import axios from 'axios';

const API = axios.create({
  baseURL: 'http://localhost:8080/',
});

// Attach token if present
API.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = token;
  }
  return config;
});

export const register = async (data) => {
  const res = await API.post('/signup', data);
  return res.data;
};

export const login = async (data) => {
  const res = await API.post('/login', data);
  return res.data;
};

export const getUser = async () => {
  const res = await API.get('/user');
  return res.data;
};

export const logout = async () => {
  localStorage.removeItem('token');
};
