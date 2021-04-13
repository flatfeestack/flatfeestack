import { token, user } from "./auth";
import { API } from "./api";
import { User } from "../types/user";
import { AxiosResponse } from "axios";

export const confirmReset = async(email: string, password: string, emailToken: string) => {
  const res = await API.auth.confirmReset(email, password, emailToken);
  const t = res.data.access_token;
  const r = res.data.refresh_token;
  if (!t || !r) {
    console.log("could not verify in reset");
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
    console.log("could not verify in email");
    return;
  }
  token.set(t);
  localStorage.setItem("ffs-refresh", r);
};

export const confirmInvite = async(email: string, password: string,
                                   emailToken: string, inviteEmail: string,
                                   inviteDate: string, inviteToken: string) => {
  const res = await API.auth.confirmInvite(email, password, emailToken, inviteEmail, inviteDate, inviteToken);
  const t = res.data.access_token;
  const r = res.data.refresh_token;
  if (!t || !r) {
    console.log("could not verify in invite");
    return;
  }
  token.set(t);
  localStorage.setItem("ffs-refresh", r);
}

export const login = async (email: string, password: string) => {
  const res = await API.auth.login(email, password);
  storeToken(res);
  const u = await API.user.get();
  user.set(u.data);
};

export const updateUser = async () => {
  const u = await API.user.get();
  user.set(u.data);
}

export const removeSession = async () => {
  try {
    await API.authToken.logout();
  } finally {
    localStorage.removeItem("ffs-refresh")
    user.set(<User>{})
    token.set("");
  }
}

export const refresh = async ():Promise<string> => {
  const refresh = localStorage.getItem("ffs-refresh");
  if (refresh === null) {
    throw 'No refresh token not in local storage';
  }
  const res = await API.auth.refresh(refresh);
  storeToken(res);
  return res.data.access_token;
}

const storeToken = (res: AxiosResponse) => {
  const t = res.data.access_token;
  const r = res.data.refresh_token;
  if (!t || !r) {
    throw "No token in the request";
  }
  token.set(t);
  localStorage.setItem("ffs-refresh", r);
}

//https://stackoverflow.com/questions/38552003/how-to-decode-jwt-token-in-javascript-without-using-a-library
/*const parseJwt = (token) => {
  const base64Url = token.split('.')[1];
  const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
  const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
    return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
  }).join(''));

  return JSON.parse(jsonPayload);
};*/
