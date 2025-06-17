import { api } from './core';

export const register = async (data) => {
  const res = await api.post('/signup', data);
  return res.data;
};

export const login = async (data) => {
  const res = await api.post('/login', data);
  return res.data;
};

export const getUser = async () => {
  const res = await api.get('/user');
  return res.data;
};

export const logout = async () => {
  localStorage.removeItem('token');
};
