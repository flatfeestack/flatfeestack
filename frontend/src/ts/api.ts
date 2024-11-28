import ky from "ky";
import { get } from "svelte/store";
import type {
  ChartDataTotal,
  ClientSecret,
  Config,
  Contribution,
  ContributionSummary,
  GitUser,
  Invitation,
  PaymentEvent,
  PaymentResponse,
  Repo,
  RepoMapping,
  Time,
  User,
  UserStatus,
  PayoutResponse,
  PublicUser,
  UserBalance,
  RepoHealthValue,
  HealthValueThreshold,
  RepoMetrics,
} from "../types/backend";
import type { Token } from "../types/auth";
import { token } from "./mainStore";
import { refresh } from "./services";
import type { DaoConfig, PayoutConfig } from "../types/payout";
import type {
  Comment,
  CommentId,
  CommentInput,
  Post,
  PostId,
  PostInput,
} from "../types/forum";

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
  throwHttpErrors: false,
  hooks: {
    beforeRequest: [async (request) => addToken(request)],
    afterResponse: [
      async (request: Request, options: any, response: Response) => {
        refreshToken(request, options, response);
        if (response.status !== 200) {
          const body = await response.json();
          if (body.error) {
            throw new Error(body.error);
          }
        }
      },
    ],
  },
});

const backendToken = ky.create({
  prefixUrl: "/backend",
  timeout: restTimeout,
  throwHttpErrors: false,
  hooks: {
    beforeRequest: [async (request) => addToken(request)],
    afterResponse: [
      async (request: Request, options: any, response: Response) => {
        refreshToken(request, options, response);
        if (response.status !== 200) {
          const body = await response.json();
          if (body.error) {
            throw new Error(body.error);
          }
        }
      },
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

const payout = ky.create({
  prefixUrl: "/payout",
  timeout: restTimeout,
});

const search = ky.create({
  prefixUrl: "/search",
  timeout: restTimeout,
});

const forum = ky.create({
  prefixUrl: "/forum",
  timeout: restTimeout,
});

const forumToken = ky.create({
  prefixUrl: "/forum",
  timeout: restTimeout,
  hooks: {
    beforeRequest: [async (request) => addToken(request)],
    afterResponse: [
      async (request: Request, options: any, response: Response) =>
        refreshToken(request, options, response),
    ],
    beforeError: [
      async (error) => {
        const { response } = error;
        if (response) {
          try {
            const r = await response.json();
            if (r.error) {
              error.message = r.error;
            }
          } catch (e) {
            // Handle any errors that occur during the response parsing
            console.error("Error parsing response:", e);
          }
        }
        return error;
      },
    ],
  },
});

export const API = {
  authToken: {
    logout: () => authToken.get(`authen/logout?redirect_uri=/`),
    timeWarp: (hours: number) =>
      authToken.post(`admin/timewarp/${hours}`).json<Token>(),
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
    get: () => backendToken.get(`users/me`).json<User>(),
    gitEmails: () => backendToken.get(`users/me/git-email`).json<GitUser[]>(),
    confirmGitEmail: (email: string, token: string) =>
      backendToken.post("users/me/git-email/confirm", {
        json: { email, token },
      }),
    addEmail: (email: string) =>
      backendToken.post(`users/me/git-email`, { json: { email } }),
    removeGitEmail: (email: string) =>
      backendToken.delete(`users/me/git-email/${email}`),
    updatePaymentMethod: (method: string) =>
      backendToken.put(`users/me/method/${method}`).json<User>(),
    deletePaymentMethod: () => backendToken.delete(`users/me/method`),
    getSponsored: () => backendToken.get("users/me/sponsored").json<Repo[]>(),
    getMultiplier: () => backendToken.get("users/me/multiplied").json<Repo[]>(),
    setName: (name: string) => backendToken.put(`users/me/name/${name}`),
    clearName: () => backendToken.put(`users/me/clear/name`),
    setMultiplier: (isSet: boolean) =>
      backendToken.put(`users/me/multiplier/${isSet}`),
    setMultiplierDailyLimit: (amount: number) =>
      backendToken.put(`users/me/multiplierDailyLimit/${amount}`),
    setImage: (image: string) =>
      backendToken.post(`users/me/image`, { json: { image } }),
    deleteImage: () => backendToken.delete(`users/me/image`),
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
    payment: () => backendToken.get(`users/me/payment`).json<PaymentEvent[]>(),

    statusSponsoredUsers: () =>
      backendToken.post(`users/me/sponsored-users`).json<UserStatus[]>(),
    contributionsSend: () =>
      backendToken.post(`users/contrib-snd`).json<Contribution[]>(),
    contributionsRcv: () =>
      backendToken.post(`users/contrib-rcv`).json<Contribution[]>(),
    contributionsSummary: () =>
      backendToken
        .post(`users/me/contributions-summary`)
        .json<ContributionSummary[]>(),
    contributionsSummary2: (uuid: string) =>
      backend
        .get(`users/contributions-summary/${uuid}`)
        .json<ContributionSummary[]>(),
    summary: (uuid: string) =>
      backend.get(`users/summary/${uuid}`).json<User>(),
    requestPayout: (targetCurrency: "ETH" | "GAS" | "USD") =>
      backendToken
        .post(`users/me/request-payout/${targetCurrency}`)
        .json<PayoutResponse>(),
    getUser: (userId: string) =>
      backend.get(`users/${userId}`).json<PublicUser>(),
    userBalance: () =>
      backendToken.get(`users/me/balance`).json<UserBalance[]>(),
  },
  repos: {
    search: (s: string) =>
      backendToken
        .get(`repos/search?q=${encodeURIComponent(s)}`)
        .json<Repo[] | null>(),
    searchName: (s: string) =>
      backendToken.get(`repos/name?q=${encodeURIComponent(s)}`).json<Repo[]>(),
    get: (id: number) => backendToken.get(`repos/${id}`),
    tag: (repoId: string) =>
      backendToken.post(`repos/${repoId}/tag`).json<Repo>(),
    untag: (repoId: string) => backendToken.post(`repos/${repoId}/untag`),
    setMultiplier: (repoId: string) =>
      backendToken.post(`repos/${repoId}/setMultiplier`).json<Repo>(),
    unsetMultiplier: (repoId: string) =>
      backendToken.post(`repos/${repoId}/unsetMultiplier`),
    graph: (repoId: string, offset: number) =>
      backendToken
        .get(`repos/${repoId}/${offset}/graph`)
        .json<ChartDataTotal>(),
    trust: (repoId: string) =>
      backendToken.post(`repos/${repoId}/trust`).json<Repo>(),
    untrust: (repoId: string) => backendToken.post(`repos/${repoId}/untrust`),
    getTrusted: () => backendToken.get("repos/trusted").json<Repo[]>(),
    getHealthValue: (repoId: string) =>
      backendToken.get(`repos/${repoId}/healthvalue`).json<RepoHealthValue>(),
    getLatestHealthValueThresholds: () =>
      backendToken
        .get(`repos/healthvaluethreshold`)
        .json<HealthValueThreshold>(),
    getRepoMetricsById: (repoId: string) =>
      backendToken
        .get(`repos/${repoId}/healthvalue/metrics`)
        .json<RepoMetrics>(),
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
    time: () => backendToken.get(`admin/time`).json<Time>(),
    fakeUser: (email: string) => backendToken.post(`admin/fake/user/${email}`),
    fakePayment: (email: string, seats: number) =>
      backendToken.post(`admin/fake/payment/${email}/${seats}`),
    fakeContribution: (repo: RepoMapping) =>
      backendToken.post(`admin/fake/contribution`, { json: repo }),
    payout: (exchangeRate: number) =>
      backendToken.post(`admin/payout/${exchangeRate}`),
  },
  config: {
    config: () => backend.get(`config`).json<Config>(),
  },
  admin: {
    users: () => backendToken.post(`admin/users`).json<string[]>(),
  },
  payout: {
    daoConfig: () => payout.get(`config/dao`).json<DaoConfig>(),
    payoutConfig: () => payout.get(`config/payout`).json<PayoutConfig>(),
  },
  forum: {
    getAllPosts: (open?: boolean) => {
      let url = `posts`;

      if (open) {
        url = `${url}?open=${open}`;
      }

      return forum.get(url).json<Post[]>();
    },
    createPost: (postInput: PostInput) =>
      forumToken.post(`posts`, { json: postInput }).json<Post>(),
    getPost: (postId: PostId) => forum.get(`posts/${postId}`).json<Post>(),
    getPostByProposalId: (proposalId: string) =>
      forum.get(`posts/byProposalId/${proposalId}`).json<Post>(),
    deletePost: (postId: PostId) => forumToken.delete(`posts/${postId}`),
    updatePost: (postId: PostId, postInput: PostInput) =>
      forumToken.put(`posts/${postId}`, { json: postInput }).json<Post>(),
    getAllComments: (postId: PostId) =>
      forum.get(`posts/${postId}/comments`).json<Comment[]>(),
    createComment: (postId: PostId, commentInput: CommentInput) =>
      forumToken
        .post(`posts/${postId}/comments`, { json: commentInput })
        .json<Comment>(),
    updateComment: (
      postId: PostId,
      commentId: CommentId,
      commentInput: CommentInput
    ) =>
      forumToken
        .put(`posts/${postId}/comments/${commentId}`, { json: commentInput })
        .json<Comment>(),
    closePost: (postId) =>
      forumToken.put(`posts/${postId}/close`).json<String>(),
    deleteComment: (postId, commentId) =>
      forumToken.delete(`posts/${postId}/comments/${commentId}`),
  },
};
