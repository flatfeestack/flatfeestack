import { token, user, showSignin, userBalances, config } from "./auth";
import { API } from "./api";
import { UserBalances, Users } from "../types/users";
import { AxiosResponse } from "axios";
import { get } from "svelte/store";

export const confirmReset = async(email: string, password: string, emailToken: string) => {
  const res = await API.auth.confirmReset(email, password, emailToken);
  storeToken(res);
};

export const confirmEmail = async(email: string, emailToken: string) => {
  const res = await API.auth.confirmEmail(email, emailToken);
  storeToken(res);
};

export const confirmInvite = async(email: string, password: string,
                                   emailToken: string, inviteEmail: string,
                                   inviteDate: string, inviteToken: string) => {
  const res = await API.auth.confirmInvite(email, password, emailToken, inviteEmail, inviteDate, inviteToken);
  storeToken(res);
}

export const login = async (email: string, password: string) => {
  console.log("MAOUOEMUOEMUSONEUOEu");
  const res = await API.auth.login(email, password);
  storeToken(res);
  const u = await API.user.get();
  user.set(u.data);
  const c = await API.user.config();
  config.set(c.data);
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
    user.set(<Users>{})
    token.set("");
    showSignin.set(true);
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
    showSignin.set(true);
    throw "No token in the request";
  }
  showSignin.set(false);
  token.set(t);
  localStorage.setItem("ffs-refresh", r);
}

function connect():Promise<WebSocket> {
  return new Promise(function(resolve, reject) {
    const t = get(token);
    const c = get(config)
    const server = new WebSocket(`${c.wsBaseUrl}/ws/users/me/payment`, ["access_token", t]);
    server.onopen = function() {
      resolve(server);
    };
    server.onerror = function(err) {
      reject(err);
    };

  });
}

export const connectWs = async () => {
  try {
    const ws = await connect();

    ws.onmessage = function(event) {
      console.log(event.data);
      try {
        userBalances.set(JSON.parse(event.data));
        console.log("current paymentCycleId: " + JSON.parse(event.data));
      } catch (e) {
        console.log(e);
      }
    };
    ws.onclose = function(e) {
      console.log('Socket is closed. Reconnect will be attempted in 1 second.', e.reason);
      setTimeout(function() {
        connectWs();
      }, 1000);
    };
    ws.onerror = function(err) {
      console.error('Socket encountered error: ', err, 'Closing socket');
      ws.close();
      setTimeout(async function() {
        await refresh();
        connectWs();
      }, 3000);
    };
  } catch (e) {
    console.log(e);
  }
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
