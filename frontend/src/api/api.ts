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
  baseURL: "/backend",
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

apiInstance.interceptors.response.use(
  (response) => {
    return response;
  },
  async function (error) {
    const originalRequest = error.config;
    if (error.response.status === 418 && !originalRequest._retry) {
      console.log("referseh?")
      originalRequest._retry = true;
      const oldR = get(refresh);
      if (!oldR) {
        console.log("could not refresh");
        return;
      }
      const res = await API.auth.refresh(oldR);
      const t = res.data.access_token;
      const r = res.data.refresh_token;
      console.log("new toke: "+t)
      token.set(t);
      refresh.set(r);
      console.log("orig request")
      console.log(originalRequest)
      originalRequest.headers.Authorization = "Bearer " + t;
      console.log(originalRequest)
      return apiInstance(originalRequest);
    }
    return Promise.reject(error);
  }
);

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
      return authInstance.post("/refresh", { refresh_token: refresh });
    },
    logout: () => {
      const t = get(token);
      if (t) {
        token.set(null);
        refresh.set(null);
        const config = {
          headers: { Authorization: `Bearer ${t}` }
        };
        return authInstance.get(`/authen/logout?redirect_uri=/`, config)
      } else {
        throw new Error("t not found...");
      }
    },
    timeWarp: (hours: number) => authInstance.post(`/timewarp/${hours}`),
  },
  user: {
    get: () => apiInstance.get(`/users/me`),
    connectedEmails: () => apiInstance.get(`/users/me/connectedEmails`),
    addEmail: (email: string) =>
      apiInstance.post(`/users/me/connectedEmails`, { email }),
    removeEmail: (email: string) =>
      apiInstance.delete(`/users/me/connectedEmails/${encodeURI(email)}`),
    updatePayoutAddress: (address: string) =>
      apiInstance.put(`/users/me/payout/${address}`),
    getSponsored: () => apiInstance.get("/users/me/sponsored"),
  },
  payments: {
    createSubscription: (plan: string, paymentMethod: string) =>
      apiInstance.post("/payments/subscriptions", { plan, paymentMethod }),
  },
  repos: {
    search: (s: string) => apiInstance.get(`/repos/search?q=${encodeURI(s)}`),
    sponsor: (id: number) => apiInstance.post(`/repos/sponsor/github/${id}`),
    unsponsor: (id: number) => apiInstance.post(`/repos/${id}/unsponsor`),
  },
  search: {
    keywords: (keywords: string) => searchInstance.get(`/search/${keywords}`),
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
  payouts: {
    pending: () => apiInstance.post(`/admin/pending-payout`),
    time: () => apiInstance.get(`/admin/time`),
    fakeUser: () => apiInstance.post(`/admin/fake-user`),
    timeWarp: (hours: number) => apiInstance.post(`/admin/timewarp/${hours}`),
  }
};
