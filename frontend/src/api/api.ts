import axios from "axios";
import { refresh, token } from "../store/auth";
import { get } from "svelte/store";
import { Exchange } from "../types/exchange.type";
import { PayoutAddress } from "../types/payout-address.type";

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
  if (t) {
    config.headers.Authorization = "Bearer " + t;
    return config;
  }
  return config;
});

apiInstance.interceptors.response.use((response) => {
  return response
}, async function (error) {
  const originalRequest = error.config;
  if (error.response.status === 418 && !originalRequest._retry) {
    originalRequest._retry = true;
    const oldR = get(refresh);
    if (!oldR) {
      console.log("could not refresh");
      return;
    }
    const res = await API.auth.refresh(oldR)
    const t = res.data.access_token
    const r = res.data.refresh_token
    token.set(t)
    refresh.set(r)
    axios.defaults.headers.common['Authorization'] = 'Bearer ' + t;
    return apiInstance(originalRequest);
  }
  return Promise.reject(error);
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
    refresh: (refresh: string) => {
      return authInstance.post("/refresh", {"refresh_token":refresh})
    },
  },
  user: {
    get: () => apiInstance.get(`/users/me`),
    connectedEmails: () => apiInstance.get(`/users/me/connectedEmails`),
    addEmail: (email: string) =>
      apiInstance.post(`/users/me/connectedEmails`, { email }),
    removeEmail: (email: string) =>
      apiInstance.delete(`/users/me/connectedEmails/${encodeURI(email)}`),
    updatePayoutAddress: (p: PayoutAddress) =>
      apiInstance.put("/users/me/payout", p),
    getPayoutAddresses: () => apiInstance.get("/users/me/payout"),
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
  search: {
    keywords: (keywords: string) =>
      searchInstance.get(`/search/${keywords}`),
  },
  exchanges: {
    get: () => apiInstance.get(`/exchanges`),
    update: (e: Exchange) =>
      apiInstance.put(`/exchanges/${e.id}`, {
        date: e.date,
        price: e.price,
        id: e.id,
      }),
  },
};
