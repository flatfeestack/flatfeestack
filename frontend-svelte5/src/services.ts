import {appState} from "./ts/state.ts";
import { API } from "ts/api.ts";
import type { ClientSecret } from "types/backend";

import { get } from "svelte/store";
import type { Stripe, StripeCardElement } from "@stripe/stripe-js";
//import { formatUnits } from "ethers";

//https://stackoverflow.com/questions/38552003/how-to-decode-jwt-token-in-javascript-without-using-a-library
/*export const parseJwt = (token) => {
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
};*/

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

/*export const formatPaymentCycle = (c: string): string => {
  return c.substring(0, 10) + "…";
};*/

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
  const currency = appState.$state.config.supportedCurrencies[c.toUpperCase()];
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
    const currency = appState.$state.config.supportedCurrencies[c.toUpperCase()];
    if (!currency) {
      console.debug("Unknown currency: " + c);
      return n.toString(10);
    }
    //return formatUnits(n, currency.factorPow);
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

export const stripePaymentMethod = async (
  stripe: Stripe,
  cardElement: StripeCardElement
) => {
  const cs: ClientSecret = await API.user.setupStripe();
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
  appState.$state.user =await API.user.updatePaymentMethod(result.setupIntent.payment_method);
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

/*export function isIterable(value) {
  return Symbol.iterator in Object(value);
}*/

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
