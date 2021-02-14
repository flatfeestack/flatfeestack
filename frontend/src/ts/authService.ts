import { initialFetchDone, loading, token, user } from "./auth";
import { API } from "./api";
import { User } from "../types/user";
import { get } from "svelte/store";

export const confirmReset = async(email: string, password: string, emailToken: string) => {
  const res = await API.auth.confirmReset(email, password, emailToken);
  const t = res.data.access_token;
  const r = res.data.refresh_token;
  if (!t || !r) {
    console.log("could not verify");
    return;
  }
  token.set(t);
  localStorage.setItem("ffs-refresh", r);
};

export const confirmEmail = async(email: string, emailToken: string) => {
  const res = await API.auth.confirmEmail(email, emailToken);
  const t = res.data.access_token;
  const r = res.data.refresh_token;
  if (!t || !r) {
    console.log("could not verify");
    return;
  }
  token.set(t);
  localStorage.setItem("ffs-refresh", r);
};

export const login = async (email: string, password: string) => {
  const res = await API.auth.login(email, password);
  const t = res.data.access_token;
  const r = res.data.refresh_token;
  if (!t || !r) {
    console.log("could not login");
    return;
  }
  token.set(t);
  localStorage.setItem("ffs-refresh", r);
  const u = await API.user.get();
  console.log(u);
  user.set(u.data);
};

export const updateUser = async () => {
  try {
    const u = await API.user.get();
    user.set(u.data);
  } catch (e) {
    console.log("could not fetch user", e);
  }
};

export const removeSession = async () => {
  localStorage.removeItem("ffs-refresh")
  user.set(<User>{})
  token.set("");
}

export const refreshSession = async () => {
  try {
    loading.set(true);
    const r = localStorage.getItem("ffs-refresh");
    const res = await API.auth.refresh(r);
    const t = res.data.access_token;
    const new_r = res.data.refresh_token;
    token.set(t);
    localStorage.setItem("ffs-refresh", new_r);
    await updateUser();
  } catch (e) {
    console.log("could not refresh session");
  } finally {
    loading.set(false);
  }
};

//https://stackoverflow.com/questions/38552003/how-to-decode-jwt-token-in-javascript-without-using-a-library
/*const parseJwt = (token) => {
  const base64Url = token.split('.')[1];
  const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
  const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
    return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
  }).join(''));

  return JSON.parse(jsonPayload);
};*/
