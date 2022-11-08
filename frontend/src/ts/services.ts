import { token, user, loginFailed, userBalances, config } from "./mainStore";
import { API } from "./api";
import type { Token, Users } from "../types/users";
import { get } from "svelte/store";
import { formatUnits } from "ethers/lib/utils";
import { BigNumber } from "ethers";

export const confirmReset = async (
  email: string,
  password: string,
  emailToken: string
) => {
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

export const confirmEmail = async (email: string, emailToken: string) => {
  const p1 = API.auth.confirmEmail(email, emailToken);
  const p2 = API.config.config();

  const res = await p1;
  storeToken(res);
  const p3 = API.user.get();

  const conf = await p2;
  config.set(conf);

  const u = await p3;
  user.set(u);
};

export const confirmInvite = async (
  email: string,
  password: string,
  emailToken: string,
  inviteByEmail: string
) => {
  const p1 = API.auth.confirmInvite(email, password, emailToken);
  const p2 = API.config.config();

  const res = await p1;
  storeToken(res);

  const p3 = API.invite.confirmInvite(inviteByEmail);
  const p4 = API.user.get();
  const conf = await p2;
  config.set(conf);

  await p3;

  const u = await p4;
  user.set(u);
};

export const login = async (email: string, password: string) => {
  const p1 = API.auth.login(email, password);
  const p2 = API.config.config();

  const res = await p1;
  storeToken(res);

  const conf = await p2;
  config.set(conf);

  const u = await API.user.get();
  user.set(u);
};

export const refresh = async (): Promise<string> => {
  const p2 = API.config.config();
  const refresh = localStorage.getItem("ffs-refresh");
  if (refresh === null) {
    throw "No refresh token not in local storage";
  }
  const p1 = API.auth.refresh(refresh);

  const tok = await p1;
  const conf = await p2;
  config.set(conf);
  storeToken(tok);
  return tok.access_token;
};

export const removeSession = async () => {
  try {
    await API.authToken.logout();
  } finally {
    removeToken();
  }
};

export const removeToken = () => {
  localStorage.removeItem("ffs-refresh");
  user.set(<Users>{});
  token.set("");
  loginFailed.set(true);
};

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
};

const connect = (): Promise<WebSocket> => {
  return new Promise(function (resolve, reject) {
    console.log("connect");
    const t = get(token);
    const c = get(config);
    console.log("ws to", `${c.wsBaseUrl}/ws/users/me/payment`);
    const server = new WebSocket(`${c.wsBaseUrl}/ws/users/me/payment`, [
      "access_token",
      t,
    ]);
    server.onopen = function () {
      console.log("open");
      resolve(server);
    };
    server.onerror = function (err) {
      console.log(err);
      reject(err);
    };
  });
};

let timeoutOnclose;
let timeoutOnerror;

export const connectWs = async () => {
  if (timeoutOnclose) {
    clearTimeout(timeoutOnclose);
    timeoutOnclose = undefined;
  }
  if (timeoutOnerror) {
    clearTimeout(timeoutOnerror);
    timeoutOnerror = undefined;
  }
  try {
    const ws = await connect();
    ws.onmessage = function (event: MessageEvent) {
      try {
        const json = JSON.parse(event.data);
        userBalances.set(json);
      } catch (e) {
        console.log(e);
      }
    };
    ws.onclose = async function (e: CloseEvent) {
      console.log(
        "Socket is closed. Reconnect will be attempted in 3 second.",
        e
      );
      if (e.code === 4001) {
        await refresh();
        await connectWs();
      } else {
        timeoutOnclose = setTimeout(async function () {
          await connectWs();
        }, 3000);
      }
    };
  } catch (e) {
    console.log(e);
    timeoutOnerror = setTimeout(async function () {
      await connectWs();
    }, 5000);
  }
};

//https://stackoverflow.com/questions/38552003/how-to-decode-jwt-token-in-javascript-without-using-a-library
export const parseJwt = (token) => {
  const base64Url = token.split(".")[1];
  const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
  const jsonPayload = decodeURIComponent(
    atob(base64)
      .split("")
      .map(function (c) {
        return "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2);
      })
      .join("")
  );

  return JSON.parse(jsonPayload);
};

//https://stackoverflow.com/questions/3552461/how-to-format-a-javascript-date
export const formatDate = (d: Date): string => {
  return (
    d.getFullYear() +
    "-" +
    ("0" + (1 + d.getMonth())).slice(-2) +
    "-" +
    ("0" + d.getDate()).slice(-2) +
    " " +
    ("0" + d.getHours()).slice(-2) +
    ":" +
    ("0" + d.getMinutes()).slice(-2) +
    ":" +
    ("0" + d.getSeconds()).slice(-2)
  );
};

export function formatNowUTC() {
  const date = new Date();
  const nowUtc = new Date(
    date.getUTCFullYear(),
    date.getUTCMonth(),
    date.getUTCDate(),
    date.getUTCHours(),
    date.getUTCMinutes(),
    date.getUTCSeconds()
  );
  return formatDate(nowUtc);
}

export const formatPaymentCycle = (c: string): string => {
  return c.substring(0, 10) + "…";
};

export const formatDay = (d: Date): string => {
  return (
    d.getFullYear() +
    "-" +
    ("0" + (d.getMonth() + 1)).slice(-2) +
    "-" +
    ("0" + d.getDate()).slice(-2)
  );
};

export function minBalanceName(c: string): string {
  const conf = get(config);
  const currency = conf.supportedCurrencies[c.toUpperCase()];
  if (!currency) {
    console.debug("Unknown currency: " + c);
    return c;
  }
  return currency.smallest;
}

export const formatBalance = (n: bigint, c: string): string => {
  if (c === "USD") {
    if (n > BigInt(1000000) || n <= BigInt(-1000000)) {
      const num = BigInt(n) / BigInt(1000000);
      return num.toString(10);
    } else if (n == BigInt(0)) {
      return "$0";
    } else {
      const num = BigInt(n) / BigInt(10000);
      return num.toString(10) + "¢";
    }
  } else {
    const conf = get(config);
    const currency = conf.supportedCurrencies[c.toUpperCase()];
    if (!currency) {
      console.debug("Unknown currency: " + c);
      return n.toString(10);
    }
    return formatUnits(BigNumber.from(n.toString()), currency.factorPow);
  }
};

//https://stackoverflow.com/questions/3177836/how-to-format-time-since-xxx-e-g-4-minutes-ago-similar-to-stack-exchange-site
export const timeSince = (d: Date, now: Date): string => {
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
};

export const stripePaymentMethod = async (stripe, cardElement) => {
  const cs = await API.user.setupStripe();
  const result = await stripe.confirmCardSetup(
    cs.clientSecret,
    { payment_method: { card: cardElement } },
    { handleActions: false }
  );
  if (result.error) {
    throw (
      "Card problem: " +
      result.error.code +
      (result.error.decline_code ? ", " + result.error.decline_code : "")
    );
  }
  user.set(
    await API.user.updatePaymentMethod(result.setupIntent.payment_method)
  );
};

export const stripePayment = async (
  stripe,
  freq: number,
  seats: number,
  paymentMethod: string
) => {
  const res = await API.user.stripePayment(freq, seats);
  const result = await stripe.confirmCardPayment(res.clientSecret, {
    payment_method: paymentMethod,
  });
  if (result.error) {
    throw (
      "Payment problem: " +
      result.error.code +
      (result.error.decline_code ? ", " + result.error.decline_code : "")
    );
  }
};

export function isIterable(value) {
  return Symbol.iterator in Object(value);
}

export function qrString(address, currency, value) {
  switch (currency) {
    case "ETH":
      //https://ethereum.stackexchange.com/questions/66508/ethereum-qr-code-with-amount
      return "ethereum:" + address + "?value=" + value;
    case "GAS":
      //https://github.com/nickfujita/neo-qrcode
      return "neo:" + address + "?asset=gas&amount=" + value;
  }
}
