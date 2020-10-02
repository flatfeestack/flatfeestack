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

apiInstance.interceptors.request.use((config) => {
  const t = get(token);
  console.log({ t });
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
      authInstance.post("/login", { email, password }),
  },
  test: {
    token: () => apiInstance.post("/test"),
  },
};
