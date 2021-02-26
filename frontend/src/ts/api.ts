import axios from "axios";
import { token } from "./auth";
import { get } from "svelte/store";
import { Exchange } from "../types/exchange.type";
import { PayoutAddress } from "../types/payout-address.type";
import { removeSession } from "./authService";
import { Repo } from "../types/repo.type";

const authInstance = axios.create({
  baseURL: "/auth",
  timeout: 5000000
});

const apiInstance = axios.create({
  baseURL: "/backend",
  timeout: 5000000
});

const searchInstance = axios.create({
  baseURL: "/search",
  timeout: 5000000
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
      originalRequest._retry = true;
      const refresh = localStorage.getItem("ffs-refresh");
      if (refresh) {
        console.log("could not refresh");
        return;
      } else {
        console.log(refresh)
      }
      const res = await API.auth.refresh(refresh);
      const t = res.data.access_token;
      const r = res.data.refresh_token;
      token.set(t);
      localStorage.setItem("ffs-refresh", r);
      originalRequest.headers.Authorization = "Bearer " + t;
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
    reset: (email: string) => {
      return authInstance.post(`/reset/${email}`);
    },
    confirmEmail: (email: string, token: string) => {
      return authInstance.post("/confirm/signup", {email, token})
    },
    confirmReset: (email: string, password: string, token: string) => {
      return authInstance.post("/confirm/reset", {email, password, email_token_reset: token})
    },
    logout: () => {
      const t = get(token);
      if (t) {
        removeSession();
        const config = {
          headers: { Authorization: `Bearer ${t}` }
        };
        //TODO: logout also with refreshToken
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
    addEmail: (email: string) => apiInstance.post(`/users/me/connectedEmails`, { email }),
    removeEmail: (email: string) => apiInstance.delete(`/users/me/connectedEmails/${encodeURI(email)}`),
    updatePayoutAddress: (address: string) => apiInstance.put(`/users/me/payout/${address}`),
    getSponsored: () => apiInstance.get("/users/me/sponsored"),
    setName: (name: string) => apiInstance.put(`/users/me/name/${name}`),
    setImage: (image: string) => apiInstance.post(`/users/me/image`, {image}),
  },
  payments: {
    createSubscription: (plan: string, paymentMethod: string) =>
      apiInstance.post("/payments/subscriptions", { plan, paymentMethod }),
  },
  repos: {
    search: (s: string) => apiInstance.get(`/repos/search?q=${encodeURI(s)}`),
    get: (id: number) => apiInstance.get(`/repos/${id}`),
    tag: (repo: Repo) => apiInstance.post(`/repos/tag`, repo),
    untag: (id: string) => apiInstance.post(`/repos/${id}/untag`),
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
    pending: (type: string) => apiInstance.post(`/admin/pending-payout/${type}`),
    time: () => apiInstance.get(`/admin/time`),
    fakeUser: () => apiInstance.post(`/admin/fake-user`),
    timeWarp: (hours: number) => apiInstance.post(`/admin/timewarp/${hours}`),
    payout: (exchangeRate: number) => apiInstance.post(`/admin/payout/${exchangeRate}`),
  }
};
