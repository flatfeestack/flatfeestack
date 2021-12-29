import { token, user, loginFailed, userBalances, config } from "./store";
import { API } from "./api";
import { Token, Users } from "../types/users";
import { get } from "svelte/store";

export const confirmReset = async(email: string, password: string, emailToken: string) => {
  const p1 = API.auth.confirmReset(email, password, emailToken);
  const p2 = API.config.config();

  const res = await p1;
  const p3 = API.user.get();

  const conf = await p2;
  config.set(conf);
  storeToken(res);

  const u = await p3;
  user.set(u);
};

export const confirmEmail = async(email: string, emailToken: string) => {
  const p1 = API.auth.confirmEmail(email, emailToken);
  const p2 = API.config.config();

  const res = await p1;
  const p3 = API.user.get();

  const conf = await p2;

  config.set(conf);
  storeToken(res);

  const u = await p3;
  user.set(u);
};

export const confirmInviteNew = async(email: string, password: string,
                                      emailToken: string, inviteEmail: string,
                                      expireAt: string, inviteToken: string, inviteMeta: string) => {
  const p1 = API.invite.confirmInvite(email);
  const p2 = API.config.config();

  const res = await p1;
  const p3 = API.user.get();

  const conf = await p2;
  config.set(conf);
  storeToken(res);

  const u = await p3;
  user.set(u);
}

export const login = async (email: string, password: string) => {
  const p1 = API.auth.login(email, password);
  const p2 = API.config.config();

  const res = await p1;
  const conf = await p2;
  config.set(conf);
  storeToken(res);

  const u = await API.user.get();
  user.set(u);
};

export const refresh = async ():Promise<string> => {
  const refresh = localStorage.getItem("ffs-refresh");
  if (refresh === null) {
    throw 'No refresh token not in local storage';
  }
  const p1 = API.auth.refresh(refresh);
  const p2 = API.config.config();

  const tok = await p1;
  const conf = await p2;
  config.set(conf);
  storeToken(tok);
  return tok.access_token;
}

export const removeSession = async () => {
  try {
    await API.authToken.logout();
  } finally {
    removeToken();
  }
}

export const removeToken = () => {
  localStorage.removeItem("ffs-refresh")
  user.set(<Users>{})
  token.set("");
  loginFailed.set(true);
}

export const storeToken = (tok: Token) => {
  const t = tok.access_token;
  const r = tok.refresh_token;
  if (!t || !r) {
    loginFailed.set(true);
    throw "No token in the request";
  }
  loginFailed.set(false);
  token.set(t);
  localStorage.setItem("ffs-refresh", r);
}

const connect = ():Promise<WebSocket> => {
  return new Promise(function(resolve, reject) {
    console.log("connect")
    const t = get(token);
    const c = get(config)
    const server = new WebSocket(`${c.wsBaseUrl}/ws/users/me/payment`, ["access_token", t]);
    server.onopen = function() {
      resolve(server);
    };
    server.onerror = function(err) {
      console.log(err)
      reject(err);
    };
  });
}

let timeoutOnclose;
let timeoutOnerror;

export const connectWs = async () => {
  if(timeoutOnclose) {
    clearTimeout(timeoutOnclose);
    timeoutOnclose = undefined;
  }
  if(timeoutOnerror) {
    clearTimeout(timeoutOnerror);
    timeoutOnerror = undefined;
  }
  try {
    const ws = await connect();
    ws.onmessage = function(event:MessageEvent) {
      try {
        userBalances.set(JSON.parse(event.data));
      } catch (e) {
        console.log(e);
      }
    };
    ws.onclose = async function(e:CloseEvent) {
      console.log('Socket is closed. Reconnect will be attempted in 3 second.', e);
      if (e.code === 4001) {
        await refresh();
        await connectWs();
      } else {
        timeoutOnclose = setTimeout(async function() {
          await connectWs();
        }, 3000);
      }
    };
  } catch (e) {
    console.log(e);
    timeoutOnerror = setTimeout(async function() {
      await connectWs();
    }, 5000);
  }
}

//https://stackoverflow.com/questions/38552003/how-to-decode-jwt-token-in-javascript-without-using-a-library
export const parseJwt = (token) => {
  const base64Url = token.split('.')[1];
  const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
  const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
    return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
  }).join(''));

  return JSON.parse(jsonPayload);
};

//https://stackoverflow.com/questions/3552461/how-to-format-a-javascript-date
export const formatDate = (d: Date):string => {
  return d.getFullYear()  + "-" + ("0"+(d.getMonth())).slice(-2) + "-" +
    ("0" + d.getDate()).slice(-2) + " " + ("0" + d.getHours()).slice(-2) + ":" + ("0" + d.getMinutes()).slice(-2);
}

export const formatDay = (d: Date):string => {
  return d.getFullYear()  + "-" + ("0"+(d.getMonth())).slice(-2) + "-" +
    ("0" + d.getDate()).slice(-2);
}

export const formatMUSD= (n: number):string => {
  if(n > 1000000) {
    return (n/1000000).toFixed(2);
  } else {
    return (n/10000).toFixed(2)+'¢';
  }
}

//https://stackoverflow.com/questions/3177836/how-to-format-time-since-xxx-e-g-4-minutes-ago-similar-to-stack-exchange-site
export const timeSince = (d: Date, now:Date):string => {
  const seconds = Math.floor((now.getTime() - d.getTime()) / 1000);
  let interval = seconds / 31536000;

  if (interval > 1) {
    return Math.floor(interval) + " years";
  }
  interval = seconds / 2592000;
  if (interval > 1) {
    return Math.floor(interval) + " months";
  }
  interval = seconds / 86400;
  if (interval > 1) {
    return Math.floor(interval) + " days";
  }
  interval = seconds / 3600;
  if (interval > 1) {
    return Math.floor(interval) + " hours";
  }
  interval = seconds / 60;
  if (interval > 1) {
    return Math.floor(interval) + " minutes";
  }
  return Math.floor(seconds) + " seconds";
}

export const stripePaymentMethod = async (stripe, cardElement) => {
  const cs = await API.user.setupStripe();
  const result = await stripe.confirmCardSetup(
    cs.client_secret,
    { payment_method: { card: cardElement } },
    { handleActions: false });
  user.set(await API.user.updatePaymentMethod(result.setupIntent.payment_method));
};

export const stripePayment = async (stripe, freq: number, seats: number, payment_method: string) => {
  const res = await API.user.stripePayment(freq, seats);
  await stripe.confirmCardPayment(res.client_secret, {payment_method })
};
