import axios from "axios"

// Base for api requests
const API_BASE = "/api/"

// Sends search api request and returns json response
export const searchQuery = async (q: string, offset = 0, limit = 10) => {
    const res = await axios.get(`${API_BASE}/search`, {
        params: { q, offset, limit },
    });
    return res.data;
};

// Sends metrics api request and returns json response
export const fetchMetrics = async () => {
    const res = await axios.get(`${API_BASE}/metrics`);
    return res.data;
};
