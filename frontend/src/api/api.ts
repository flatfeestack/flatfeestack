import axios from "axios";
import { token } from "../store/auth";
import { get } from "svelte/store";

const authInstance = axios.create({
  baseURL: "/auth",
  timeout: 5000,
});

const apiInstance = axios.create({
  baseURL: "/api",
  timeout: 5000,
});

const searchInstance = axios.create({
  baseURL: "/search",
  timeout: 5000,
});

apiInstance.interceptors.request.use((config) => {
  const t = get(token);
  console.log(t);
  if (t) {
    config.headers.Authorization = "Bearer " + t;
    return config;
  }
  return config;
});

export const API = {
  auth: {
    signup: (email: string, password: string) =>
      authInstance.post("/signup", { email, password }),
    login: (email: string, password: string) =>
      authInstance.post(
        "/login",
        { email, password },
        {
          withCredentials: true,
        }
      ),
    refresh: () => authInstance.post("/refresh", null),
  },
  user: {
    get: () => apiInstance.get(`/users/me`),
  },
  payments: {
    createSubscription: (plan: string, paymentMethod: string) =>
      apiInstance.post("/payments/subscriptions", { plan, paymentMethod }),
  },
  repos: {
    search: (q: string) => apiInstance.get(`/repos/search?q=${encodeURI(q)}`),
    sponsor: (id: number) => apiInstance.post(`/repos/${id}/sponsor`),
    unsponsor: (id: number) => apiInstance.post(`/repos/${id}/unsponsor`),
    getSponsored: () => apiInstance.get("/users/sponsored"),
  },
};

export const SEARCH = {
  search: (keywords: string) =>
    searchInstance.get(`/search/${keywords}`)
};
