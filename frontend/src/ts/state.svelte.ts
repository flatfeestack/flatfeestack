import type {
  User,
  Config,
  Repo,
  HealthValueThreshold,
} from "../types/backend.ts";

export const appState = $state({
  error: null as string | null ,
  isSubmitting: false,
  user: null as User | null,
  config: null as Config | null,
  sponsoredRepos: [] as Repo[],
  multiplierSponsoredRepos: [] as Repo[],
  multiplierCountByRepo: new Map<string, number>(),
  loadedSponsoredRepos: false,
  route: "",
  accessToken: "",
  accessTokenExpire: 0,
  trustedRepos: [] as Repo[],
  loadedTrustedRepos: false,
  reposToUnTrustAfterTimeout: [] as Repo[],
  reposInSearchResult: [] as Repo[],
  loadedLatestThresholds: false,
  latestThresholds: {} as HealthValueThreshold,
  reposWaitingForNewAnalysis: [] as Repo[],
  reloadAdminSearchKey: 0,
  reloadHealthRepoCardKey: 0,

  isAccessTokenExpired(): boolean {
    const expireTime = this.accessTokenExpire;
    if (!expireTime) {
      return true;
    }
    try {
      const currentSeconds = Math.floor(Date.now() / 1000);
      const bufferSeconds = 30;
      return currentSeconds + bufferSeconds >= expireTime;
    } catch {
      return true;
    }
  },

  setError(e: any) {
    this.error = e instanceof Error ? e.message : String(e);
  },

  getAccessToken() {
    return this.accessToken;
  },
});
