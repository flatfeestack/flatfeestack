import ky from "ky";

import { appState } from "ts/state.ts";
import {refresh} from "ts/auth.ts";
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
} from "types/backend.ts";
import type { Token } from "types/auth.ts";
import type { DaoConfig, PayoutConfig } from "types/payout.ts";
/*import type {
  Comment,
  CommentId,
  CommentInput,
  Post,
  PostId,
  PostInput,
} from "types/forum.ts";*/

const timeout = 5000;

export type ApiError = {
  type: string;
  values: Map<string, string>;
}

type ApiErrorRaw = {
  type: string;
  values: { [key: string]: string };
}

function convertToApiError(rawError: ApiErrorRaw): ApiError {
  return {
    type: rawError.type,
    values: new Map(Object.entries(rawError.values))
  };
}

function createAuthenticatedApi(prefix: string, timeout: number) {
  return ky.create({
    prefixUrl: prefix,
    timeout: timeout,
    throwHttpErrors: false,
    hooks: {
      beforeRequest: [
        async (request) => {
          // If no token or token is expired, refresh immediately
          if (!appState.$state.accessToken || appState.isAccessTokenExpired()) {
            const accessToken = await refresh();
            request.headers.set('Authorization', `Bearer ${accessToken}`);
            return;
          }
          request.headers.set('Authorization', `Bearer ${appState.$state.accessToken}`);
        }
      ],
      afterResponse: [
        async (_request, _options, response) => {
          if (response.status !== 200) {
            if (response.status === 403 || response.status === 429) {
              appState.$state.accessToken = "";
              appState.$state.accessTokenExpire = "";
            }
            const rawError = await response.json();
            const apiError = convertToApiError(rawError as ApiErrorRaw);
            throw new Error(apiError.type);
          }
          return response;
        }
      ]
    },
    retry: {
      limit: 2,
      statusCodes: [401, 429]
    }
  });
}

const authToken = createAuthenticatedApi("/auth", timeout);
const backendToken = createAuthenticatedApi("/backend", timeout);
//const forumToken = createAuthenticatedApi("/forum", timeout);

const auth = ky.create({prefixUrl: "/auth", timeout});
const backend = ky.create({prefixUrl: "/backend", timeout});
const payout = ky.create({prefixUrl: "/payout", timeout});
const search = ky.create({prefixUrl: "/search", timeout});
//const forum = ky.create({prefixUrl: "/forum", timeout});

export const API = {
  authToken: {
    logout: () =>
        authToken.get(`authen/logout?redirect_uri=/`),
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
    reset: (email: string) =>
        auth.post(`reset/${email}`),
    confirmInvite: (email: string, password: string, emailToken: string) =>
        auth.post("confirm/invite", { json: { email, password, emailToken } }).json<Token>(),
    confirmEmail: (email: string, emailToken: string) =>
      auth.post("confirm/signup", { json: { email, emailToken } }).json<Token>(),
    confirmReset: (email: string, password: string, emailToken: string) =>
      auth.post("confirm/reset", { json: { email, password, emailToken } }).json<Token>(),
  },
  user: {
    get: () =>
        backendToken.get(`users/me`).json<User>(),
    gitEmails: () =>
        backendToken.get(`users/me/git-email`).json<GitUser[]>(),
    confirmGitEmail: (email: string, token: string) =>
        backendToken.post("users/me/git-email/confirm", {json: { email, token },}),
    addEmail: (email: string) =>
        backendToken.post(`users/me/git-email`, { json: { email } }),
    removeGitEmail: (email: string) =>
        backendToken.delete(`users/me/git-email/${email}`),
    updatePaymentMethod: (method: string) =>
        backendToken.put(`users/me/method/${method}`).json<User>(),
    deletePaymentMethod: () =>
        backendToken.delete(`users/me/method`),
    getSponsored: () =>
        backendToken.get("users/me/sponsored").json<Repo[]>(),
    setName: (name: string) =>
        backendToken.put(`users/me/name/${name}`),
    clearName: () =>
        backendToken.put(`users/me/clear/name`),
    setImage: (image: string) =>
        backendToken.post(`users/me/image`, { json: { image } }),
    deleteImage: () =>
        backendToken.delete(`users/me/image`),
    setupStripe: () =>
        backendToken.post(`users/me/stripe`).json<ClientSecret>(),
    stripePayment: (freq: number, seats: number) =>
        backendToken.put(`users/me/stripe/${freq}/${seats}`).json<ClientSecret>(),
    nowPayment: (currency: string, freq: number, seats: number) =>
        backendToken.post(`users/me/nowPayment/${freq}/${seats}`, { json: { currency } }).json<PaymentResponse>(),
    cancelSub: () =>
        backendToken.delete(`users/me/stripe`),
    timeWarp: (hours: number) =>
        backendToken.post(`admin/timewarp/${hours}`),
    payment: () =>
        backendToken.get(`users/me/payment`).json<PaymentEvent[]>(),
    statusSponsoredUsers: () =>
        backendToken.post(`users/me/sponsored-users`).json<UserStatus[]>(),
    contributionsSend: () =>
        backendToken.post(`users/contrib-snd`).json<Contribution[]>(),
    contributionsRcv: () =>
        backendToken.post(`users/contrib-rcv`).json<Contribution[]>(),
    contributionsSummary: () =>
        backendToken.post(`users/me/contributions-summary`).json<ContributionSummary[]>(),
    contributionsSummary2: (uuid: string) =>
        backend.get(`users/contributions-summary/${uuid}`).json<ContributionSummary[]>(),
    summary: (uuid: string) =>
        backend.get(`users/summary/${uuid}`).json<User>(),
    requestPayout: (targetCurrency: "ETH" | "GAS" | "USD") =>
        backendToken.post(`users/me/request-payout/${targetCurrency}`).json<PayoutResponse>(),
    getUser: (userId: string) =>
        backend.get(`users/${userId}`).json<PublicUser>(),
    userBalance: () =>
        backendToken.get(`users/me/balance`).json<UserBalance[]>(),
  },
  repos: {
    search: (s: string) =>
        backendToken.get(`repos/search?q=${encodeURIComponent(s)}`).json<Repo[] | null>(),
    searchName: (s: string) =>
        backendToken.get(`repos/name?q=${encodeURIComponent(s)}`).json<Repo[]>(),
    get: (id: number) =>
        backendToken.get(`repos/${id}`),
    tag: (repoId: string) =>
        backendToken.post(`repos/${repoId}/tag`).json<Repo>(),
    untag: (repoId: string) =>
        backendToken.post(`repos/${repoId}/untag`),
    graph: (repoId: string, offset: number) =>
        backendToken.get(`repos/${repoId}/${offset}/graph`).json<ChartDataTotal>(),
  },
  invite: {
    invites: () =>
        backendToken.get("invite").json<Invitation[]>(),
    invite: (email: string) =>
        backendToken.post(`invite/${email}`),
    inviteAuth: (email: string) =>
        authToken.post(`invite/${email}`),
    delMyInvite: (email: string) =>
        backendToken.delete(`invite/my/${email}`),
    delByInvite: (email: string) =>
        backendToken.delete(`invite/by/${email}`),
    confirmInvite: (email: string) =>
        backendToken.post(`confirm/invite/${email}`),
  },
  search: {
    keywords: (keywords: string) =>
        search.get(`search/${keywords}`),
  },
  payouts: {
    time: () =>
        backendToken.get(`admin/time`).json<Time>(),
    fakeUser: (email: string) =>
        backendToken.post(`admin/fake/user/${email}`),
    fakePayment: (email: string, seats: number) =>
        backendToken.post(`admin/fake/payment/${email}/${seats}`),
    fakeContribution: (repo: RepoMapping) =>
        backendToken.post(`admin/fake/contribution`, { json: repo }),
    payout: (exchangeRate: number) =>
        backendToken.post(`admin/payout/${exchangeRate}`),
  },
  config: {
    config: () =>
        backend.get(`config`).json<Config>(),
  },
  admin: {
    users: () =>
        backendToken.post(`admin/users`).json<string[]>(),
  },
  payout: {
    daoConfig: () =>
        payout.get(`config/dao`).json<DaoConfig>(),
    payoutConfig: () =>
        payout.get(`config/payout`).json<PayoutConfig>(),
  },
  /*forum: {
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
  },*/
};
