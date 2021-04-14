import axios, { AxiosError, AxiosInstance, AxiosRequestConfig, AxiosResponse, AxiosStatic } from "axios";
import { token } from "./auth";
import { get } from "svelte/store";
import { refresh, removeSession } from "./authService";
import { Repo } from "../types/repo.type";
import { ClientSecret, User } from "../types/user";

const auth = axios.create({
  baseURL: "/auth",
  timeout: 5000000
});

const authToken = axios.create({
  baseURL: "/auth",
  timeout: 5000000
});

const backend = axios.create({
  baseURL: "/backend",
  timeout: 5000000
});

const search = axios.create({
  baseURL: "/search",
  timeout: 5000000
});

function addToken(config:AxiosRequestConfig) {
  const t = get(token);
  if (t) {
    config.headers.Authorization = "Bearer " + t;
  }
  //TODO: refresh session if we see that the token expired, no need to do a request
  return config;
}

async function refreshSession(error: AxiosError, call: AxiosInstance) {
  const originalRequest = error.config;
  if (error.response.status === 401 && originalRequest.headers.Retry !== "true") {
    originalRequest.headers.Retry = "true";
    const t = await refresh();
    originalRequest.headers.Authorization = "Bearer " + t;
    return call(originalRequest);
  }
  return Promise.reject(error);
}

backend.interceptors.request.use(config => addToken(config), error => {return Promise.reject(error)});
backend.interceptors.response.use(response => {return response;}, error => {return refreshSession(error, backend)});
authToken.interceptors.request.use(config => addToken(config), error => {return Promise.reject(error)});
authToken.interceptors.request.use(response => {return response;}, error => {return refreshSession(error, authToken)});

export const API = {
  authToken: {
    invites: () => authToken.get('/invite'),
    invite: (email: string, inviteEmail: string, name: string, invitedAt: string) => authToken.post('/invite', { email, invite_email: inviteEmail, name, invitedAt }),
    delInvite: (email: string) => authToken.delete(`/invite/${email}`),
    logout: () => authToken.get(`/authen/logout?redirect_uri=/`),
  },
  auth: {
    signup: (email: string, password: string) => auth.post("/signup", { email, password }),
    login: (email: string, password: string) => auth.post("/login", { email, password }),
    refresh: (refresh: string) =>  auth.post("/refresh", { refresh_token: refresh }),
    reset: (email: string) => auth.post(`/reset/${email}`),
    confirmEmail: (email: string, token: string) => auth.post("/confirm/signup", {email, token}),
    confirmReset: (email: string, password: string, token: string) => auth.post("/confirm/reset", {email, password, email_token: token}),
    confirmInvite: (email: string, password: string, emailToken: string, inviteEmail: string, inviteDate:string, inviteToken: string) =>
      auth.post("/confirm/invite", {email, password, email_token: emailToken, invite_email: inviteEmail, invite_date: inviteDate, invite_token: inviteToken}),
    timeWarp: (hours: number) => auth.post(`/timewarp/${hours}`),
  },
  user: {
    get: () => backend.get(`/users/me`),
    gitEmails: () => backend.get(`/users/me/git-email`),
    confirmGitEmail: (email: string, token: string) => backend.post("/users/git-email", {email, token}),
    addEmail: (email: string) => backend.post(`/users/me/git-email`, { email }),
    removeGitEmail: (email: string) => backend.delete(`/users/me/git-email/${encodeURI(email)}`),
    updatePayoutAddress: (address: string) => backend.put(`/users/me/payout/${address}`),
    updatePaymentMethod: (method: string) => backend.put(`/users/me/method/${method}`),
    getSponsored: () => backend.get("/users/me/sponsored"),
    setName: (name: string) => backend.put(`/users/me/name/${name}`),
    setImage: (image: string) => backend.post(`/users/me/image`, {image}),
    setUserMode: (mode: string) => backend.put(`/users/me/mode/${mode}`),
    setupStripe: () => backend.post<ClientSecret>(`/users/me/stripe`),
    stripePayment: (freq: string, seats: number) => backend.put(`/users/me/stripe/${freq}/${seats}`),
    cancelSub: () => backend.delete(`/users/me/stripe`)
  },
  repos: {
    search: (s: string) => backend.get(`/repos/search?q=${encodeURI(s)}`),
    get: (id: number) => backend.get(`/repos/${id}`),
    tag: (repo: Repo) => backend.post(`/repos/tag`, repo),
    untag: (id: string) => backend.post(`/repos/${id}/untag`),
  },
  search: {
    keywords: (keywords: string) => search.get(`/search/${keywords}`),
  },
  payouts: {
    pending: (type: string) => backend.post(`/admin/pending-payout/${type}`),
    time: () => backend.get(`/admin/time`),
    fakeUser: () => backend.post(`/admin/fake-user`),
    timeWarp: (hours: number) => backend.post(`/admin/timewarp/${hours}`),
    payout: (exchangeRate: number) => backend.post(`/admin/payout/${exchangeRate}`),
  }
};
