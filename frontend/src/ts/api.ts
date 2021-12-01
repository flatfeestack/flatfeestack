import ky from "ky";
import { config, token } from "./store";
import { get } from "svelte/store";
import { refresh } from "./services";
import type {
  ClientSecret,
  Config,
  Invitation,
  Repo,
  Token,
  Users,
  Time,
  UserAggBalance,
  GitUser,
  RepoMapping,
  UserBalance, Contributions, UserBalanceCore,
  PayoutAddress
} from "../types/users";
import { PaymentCycle, PayoutInfo, UserStatus } from "../types/users";

async function addToken(request: Request) {
  const t = get(token);
  if (t) {
    request.headers.set('Authorization', "Bearer " + t);
  } else {
      const t = await refresh()
      if (t) {
        request.headers.set('Retry', "true");
        request.headers.set('Authorization', "Bearer " + t);
      } else {
        throw "could not set access token";
      }
  }
}

async function refreshToken(request: Request, options: any, response: Response) {
  if (response.status === 401 && request.headers.get('Retry') !== "true") {
    console.log("need to refresh due to:" + response);
    const t = await refresh();
    request.headers.set('Retry', "true");
    request.headers.set('Authorization', "Bearer " + t);
    return ky(request);
  }
}

const authToken = ky.create({
  prefixUrl: "/auth",
  timeout: get(config) ? get(config).restTimeout : 5000,
  hooks: {
    beforeRequest: [async request => addToken(request)],
    afterResponse: [async (request: Request, options: any, response: Response) => refreshToken(request, options, response)]
  }
});

const backendToken = ky.create({
  prefixUrl: "/backend",
  timeout: get(config) ? get(config).restTimeout : 5000,
  hooks: {
    beforeRequest: [async request => addToken(request)],
    afterResponse: [async (request: Request, options: any, response: Response) => refreshToken(request, options, response)]
  }
});

const auth = ky.create({
  prefixUrl: "/auth",
  timeout: get(config) ? get(config).restTimeout : 5000,
});

const backend = ky.create({
  prefixUrl: "/backend",
  timeout: get(config) ? get(config).restTimeout : 5000,
})

const search = ky.create({
  prefixUrl: "/search",
  timeout: get(config) ? get(config).restTimeout : 5000,
});


export const API = {
  authToken: {
    invites: () => authToken.get('invite').json<Invitation[]>(),
    invite: (email: string, name: string, expireAt: string, meta: string) => authToken.post('invite', {
      json: { email, name, expireAt, meta: meta.toString() } }),
    delInvite: (email: string) => authToken.delete(`invite/${email}`),
    logout: () => authToken.get(`authen/logout?redirect_uri=/`),
    timeWarp: (hours: number) => authToken.post(`timewarp/${hours}`),
    loginAs: (email: string) => authToken.post(`admin/login-as/${email}`).json<Token>()
  },
  auth: {
    signup: (email: string, password: string) => auth.post("signup", { json: { email, password } }),
    login: (email: string, password: string) => auth.post("login", { json: { email, password } }).json<Token>(),
    refresh: (refresh: string) => auth.post("refresh", { json: { refresh_token: refresh } }).json<Token>(),
    reset: (email: string) => auth.post(`reset/${email}`),
    confirmEmail: (email: string, emailToken: string) => auth.post("confirm/signup", { json: { email, emailToken } }).json<Token>(),
    confirmReset: (email: string, password: string, token: string) => auth.post("confirm/reset", {
      json: { email, password, email_token: token } }).json<Token>(),
    confirmInviteNew: (email: string, password: string, emailToken: string, inviteEmail: string, expireAt: string, inviteToken: string, inviteMeta: string) =>
      auth.post("confirm/invite-new", { json: { email, password, email_token: emailToken, inviteEmail, expireAt, inviteToken, inviteMeta }}).json<Token>(),
    confirmInvite: (email: string, inviteEmail: string, expireAt: string, inviteToken: string, inviteMeta: string) =>
      auth.post("confirm/invite", { json: { email, inviteEmail, expireAt, inviteToken, inviteMeta }}),
  },
  user: {
    get: () => backendToken.get(`users/me`).json<Users>(),
    gitEmails: () => backendToken.get(`users/me/git-email`).json<GitUser[]>(),
    confirmGitEmail: (email: string, token: string) => backendToken.post("users/git-email", {
      json: { email, token } }),
    addEmail: (email: string) => backendToken.post(`users/me/git-email`, { json: { email } }),
    removeGitEmail: (email: string) => backendToken.delete(`users/me/git-email/${encodeURI(email)}`),
    getPayoutAddresses: () => backendToken.get(`users/me/wallets`).json<PayoutAddress[]>(),
    addPayoutAddress: (currency: string, address: string) => backendToken.post(`users/me/wallets`, {json: {currency, address}}).json<PayoutAddress>(),
    removePayoutAddress: (id: number) => backendToken.delete(`users/me/wallets/${id}`),
    updatePaymentMethod: (method: string) => backendToken.put(`users/me/method/${method}`).json<Users>(),
    deletePaymentMethod: () => backendToken.delete(`users/me/method`),
    getSponsored: () => backendToken.get("users/me/sponsored").json<Repo[]>(),
    setName: (name: string) => backendToken.put(`users/me/name/${name}`),
    setImage: (image: string) => backendToken.post(`users/me/image`, { json: { image } }),
    setUserMode: (mode: string) => backendToken.put(`users/me/mode/${mode}`),
    setupStripe: () => backendToken.post(`users/me/stripe`).json<ClientSecret>(),
    stripePayment: (freq: number, seats: number) => backendToken.put(`users/me/stripe/${freq}/${seats}`).json<ClientSecret>(),
    nowpaymentsPayment: (currency: string, freq: number, seats: number) => backendToken.post(`users/me/nowpayments/${freq}/${seats}`, { json: { currency }}),
    cancelSub: () => backendToken.delete(`users/me/stripe`),
    timeWarp: (hours: number) => backendToken.post(`admin/timewarp/${hours}`),
    topup: () => backendToken.post(`users/me/topup`),
    paymentCycle: () => backendToken.post(`users/me/payment-cycle`).json<PaymentCycle>(),
    updateSeats: (seats: number)=> backendToken.post(`users/me/seats/${seats}`),
    statusSponsoredUsers: () => backendToken.post(`users/me/sponsored-users`).json<UserStatus[]>(),
    contributionsSend: () => backendToken.post(`users/me/contributions-send`).json<Contributions[]>(),
    contributionsRcv: () => backendToken.post(`users/me/contributions-receive`).json<Contributions[]>(),
    contributionsSummary: () => backendToken.post(`users/me/contributions-summary`).json<Repo[]>(),
    contributionsSummary2: (uuid: string) => backendToken.post(`users/contributions-summary/${uuid}`).json<Repo[]>(),
    summary: (uuid: string) => backendToken.post(`users/summary/${uuid}`).json<Users>(),
    pendingDailyUserPayouts: () => backendToken.post(`users/me/payout-pending`).json<UserBalanceCore>(),
    totalRealizedIncome: () => backendToken.post(`users/me/payout`).json<UserBalanceCore>(),
  },
  repos: {
    search: (s: string) => backendToken.get(`repos/search?q=${encodeURI(s)}`).json<Repo[]>(),
    get: (id: number) => backendToken.get(`repos/${id}`),
    tag: (repo: Repo) => backendToken.post(`repos/tag`, { json: repo }).json<Repo>(),
    untag: (id: string) => backendToken.post(`repos/${id}/untag`),
  },
  search: {
    keywords: (keywords: string) => search.get(`search/${keywords}`),
  },
  payouts: {
    pending: (type: string) => backendToken.post(`admin/pending-payout/${type}`).json<UserAggBalance[]>(),
    time: () => backendToken.get(`admin/time`).json<Time>(),
    fakeUser: (email: string) => backendToken.post(`admin/fake/user/${email}`),
    fakePayment: (email: string, seats: number) => backendToken.post(`admin/fake/payment/${email}/${seats}`),
    fakeContribution: (repo: RepoMapping) => backendToken.post(`admin/fake/contribution`, {json: repo}),
    payout: (exchangeRate: number) => backendToken.post(`admin/payout/${exchangeRate}`),
    payoutInfos: () => backendToken.get(`admin/payout`).json<PayoutInfo[]>(),
  },
  config: {
    config: () => backend.get(`config`).json<Config>()
  },
  admin: {
    users: () => backendToken.post(`admin/users`).json<Users[]>()
  }
};
