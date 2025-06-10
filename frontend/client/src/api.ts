import axios from "axios"

const API_BASE = "/api/"

export const searchQuery = async (q: string, offset = 0, limit = 10) => {
    const res = await axios.get(`${API_BASE}/search`, {
        params: { q, offset, limit },
    });
    return res.data;
};

export const fetchMetrics = async () => {
    const res = await axios.get(`${API_BASE}/metrics`);
    return res.data;
};
