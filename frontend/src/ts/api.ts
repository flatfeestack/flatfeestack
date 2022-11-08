import ky from "ky";
import { config, token } from "./mainStore";
import { get } from "svelte/store";
import { refresh } from "./services";
import type {
  ClientSecret,
  Config,
  Invitation,
  Repos,
  Token,
  Users,
  Time,
  UserAggBalance,
  GitUser,
  RepoMapping,
  Contributions,
  PayoutAddress,
} from "../types/users";
import type {
  PaymentCycle,
  PaymentResponse,
  PayoutInfo,
  UserStatus,
} from "../types/users";
import type { ChartDataTotal } from "../types/users";

async function addToken(request: Request) {
  const t = get(token);
  if (t) {
    request.headers.set("Authorization", "Bearer " + t);
  } else {
    const t = await refresh();
    if (t) {
      request.headers.set("Retry", "true");
      request.headers.set("Authorization", "Bearer " + t);
    } else {
      throw "could not set access token";
    }
  }
}

async function refreshToken(
  request: Request,
  options: any,
  response: Response
) {
  if (response.status === 401 && request.headers.get("Retry") !== "true") {
    console.log("need to refresh due to:" + response);
    const t = await refresh();
    request.headers.set("Retry", "true");
    request.headers.set("Authorization", "Bearer " + t);
    return ky(request);
  }
}

const restTimeout = 5000;

const authToken = ky.create({
  prefixUrl: "/auth",
  timeout: restTimeout,
  hooks: {
    beforeRequest: [async (request) => addToken(request)],
    afterResponse: [
      async (request: Request, options: any, response: Response) =>
        refreshToken(request, options, response),
    ],
  },
});

const backendToken = ky.create({
  prefixUrl: "/backend",
  timeout: restTimeout,
  hooks: {
    beforeRequest: [async (request) => addToken(request)],
    afterResponse: [
      async (request: Request, options: any, response: Response) =>
        refreshToken(request, options, response),
    ],
  },
});

const auth = ky.create({
  prefixUrl: "/auth",
  timeout: restTimeout,
});

const backend = ky.create({
  prefixUrl: "/backend",
  timeout: restTimeout,
});

const search = ky.create({
  prefixUrl: "/search",
  timeout: restTimeout,
});

export const API = {
  authToken: {
    logout: () => authToken.get(`authen/logout?redirect_uri=/`),
    timeWarp: (hours: number) =>
      authToken.post(`timewarp/${hours}`).json<Token>(),
    loginAs: (email: string) =>
      authToken.post(`admin/login-as/${email}`).json<Token>(),
  },
  auth: {
    signup: (email: string, password: string) =>
      auth.post("signup", { json: { email, password } }),
    login: (email: string, password: string) =>
      auth.post("login", { json: { email, password } }).json<Token>(),
    refresh: (refresh: string) =>
      auth.post("refresh", { json: { refresh_token: refresh } }).json<Token>(),
    reset: (email: string) => auth.post(`reset/${email}`),
    confirmInvite: (email: string, password: string, emailToken: string) =>
      auth
        .post("confirm/invite", { json: { email, password, emailToken } })
        .json<Token>(),
    confirmEmail: (email: string, emailToken: string) =>
      auth
        .post("confirm/signup", { json: { email, emailToken } })
        .json<Token>(),
    confirmReset: (email: string, password: string, emailToken: string) =>
      auth
        .post("confirm/reset", { json: { email, password, emailToken } })
        .json<Token>(),
  },
  user: {
    get: () => backendToken.get(`users/me`).json<Users>(),
    gitEmails: () => backendToken.get(`users/me/git-email`).json<GitUser[]>(),
    confirmGitEmail: (email: string, token: string) =>
      backendToken.post("users/git-email", { json: { email, token } }),
    addEmail: (email: string) =>
      backendToken.post(`users/me/git-email`, { json: { email } }),
    removeGitEmail: (email: string) =>
      backendToken.delete(`users/me/git-email/${encodeURIComponent(email)}`),
    getPayoutAddresses: () =>
      backendToken.get(`users/me/wallets`).json<PayoutAddress[]>(),
    addPayoutAddress: (currency: string, address: string) =>
      backendToken
        .post(`users/me/wallets`, { json: { currency, address } })
        .json<PayoutAddress>(),
    removePayoutAddress: (id: number) =>
      backendToken.delete(`users/me/wallets/${id}`),
    updatePaymentMethod: (method: string) =>
      backendToken.put(`users/me/method/${method}`).json<Users>(),
    deletePaymentMethod: () => backendToken.delete(`users/me/method`),
    getSponsored: () => backendToken.get("users/me/sponsored").json<Repos[]>(),
    setName: (name: string) => backendToken.put(`users/me/name/${name}`),
    setImage: (image: string) =>
      backendToken.post(`users/me/image`, { json: { image } }),
    setupStripe: () =>
      backendToken.post(`users/me/stripe`).json<ClientSecret>(),
    stripePayment: (freq: number, seats: number) =>
      backendToken.put(`users/me/stripe/${freq}/${seats}`).json<ClientSecret>(),
    nowPayment: (currency: string, freq: number, seats: number) =>
      backendToken
        .post(`users/me/nowPayment/${freq}/${seats}`, { json: { currency } })
        .json<PaymentResponse>(),
    cancelSub: () => backendToken.delete(`users/me/stripe`),
    timeWarp: (hours: number) => backendToken.post(`admin/timewarp/${hours}`),
    paymentCycle: () =>
      backendToken.post(`users/me/payment-cycle`).json<PaymentCycle>(),

    statusSponsoredUsers: () =>
      backendToken.post(`users/me/sponsored-users`).json<UserStatus[]>(),
    contributionsSend: () =>
      backendToken.post(`users/contrib-snd`).json<Contributions[]>(),
    contributionsRcv: () =>
      backendToken.post(`users/contrib-rcv`).json<Contributions[]>(),
    contributionsSummary: () =>
      backendToken.post(`users/me/contributions-summary`).json<Repos[]>(),
    contributionsSummary2: (uuid: string) =>
      backendToken.post(`users/contributions-summary/${uuid}`).json<Repos[]>(),
    summary: (uuid: string) =>
      backendToken.post(`users/summary/${uuid}`).json<Users>(),
  },
  repos: {
    search: (s: string) =>
      backendToken
        .get(`repos/search?q=${encodeURIComponent(s)}`)
        .json<Repos[]>(),
    searchName: (s: string) =>
      backendToken.get(`repos/name?q=${encodeURIComponent(s)}`).json<Repos[]>(),
    linkGitUrl: (repoId: string, gitUrl: string) =>
      backendToken
        .post(`repos/link/${repoId}`, { json: { gitUrl } })
        .json<Repos[]>(),
    makeRoot: (repoId: string, rootUuid: string) =>
      backendToken.get(`repos/root/${repoId}/${rootUuid}`).json<Repos[]>(),
    get: (id: number) => backendToken.get(`repos/${id}`),
    tag: (repoId: string) =>
      backendToken.post(`repos/${repoId}/tag`).json<Repos>(),
    untag: (repoId: string) => backendToken.post(`repos/${repoId}/untag`),
    graph: (repoId: string, offset: number) =>
      backendToken
        .get(`repos/${repoId}/${offset}/graph`)
        .json<ChartDataTotal>(),
  },
  invite: {
    invites: () => backendToken.get("invite").json<Invitation[]>(),
    invite: (email: string) => backendToken.post(`invite/${email}`),
    inviteAuth: (email: string) => authToken.post(`invite/${email}`),
    delMyInvite: (email: string) => backendToken.delete(`invite/my/${email}`),
    delByInvite: (email: string) => backendToken.delete(`invite/by/${email}`),
    confirmInvite: (email: string) =>
      backendToken.post(`confirm/invite/${email}`),
  },
  search: {
    keywords: (keywords: string) => search.get(`search/${keywords}`),
  },
  payouts: {
    pending: (type: string) =>
      backendToken
        .post(`admin/pending-payout/${type}`)
        .json<UserAggBalance[]>(),
    time: () => backendToken.get(`admin/time`).json<Time>(),
    fakeUser: (email: string) => backendToken.post(`admin/fake/user/${email}`),
    fakePayment: (email: string, seats: number) =>
      backendToken.post(`admin/fake/payment/${email}/${seats}`),
    fakeContribution: (repo: RepoMapping) =>
      backendToken.post(`admin/fake/contribution`, { json: repo }),
    payout: (exchangeRate: number) =>
      backendToken.post(`admin/payout/${exchangeRate}`),
    payoutInfos: () => backendToken.get(`admin/payout`).json<PayoutInfo[]>(),
  },
  config: {
    config: () => backend.get(`config`).json<Config>(),
  },
  admin: {
    users: () => backendToken.post(`admin/users`).json<Users[]>(),
  },
};
