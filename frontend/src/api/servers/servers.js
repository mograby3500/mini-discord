import { api } from '../core';

export const getServers = async () => {
    const res = await api.get('/servers');
    return res.data;
};
