import type {
  User,
  Config,
  Repo,
  HealthValueThreshold,
} from "../types/backend.ts";

export class AppState {
  error: string;
  isSubmitting: boolean;
  user: User;
  config: Config;
  sponsoredRepos: Repo[];
  multiplierSponsoredRepos: Repo[];
  multiplierCountByRepo: Map<string, number>;
  loadedSponsoredRepos: boolean;
  route: string;
  accessToken: string;
  accessTokenExpire: number;
  trustedRepos: Repo[];
  loadedTrustedRepos: boolean;
  reposToUnTrustAfterTimeout: Repo[];
  reposInSearchResult: Repo[];
  loadedLatestThresholds: boolean;
  latestThresholds: HealthValueThreshold;
  reposWaitingForNewAnalysis: Repo[];
  reloadAdminSearchKey: number;
  reloadHealthRepoCardKey: number;

  constructor() {
    this.error = $state("");
    this.isSubmitting = $state(false);
    this.user = $state({} as User);
    this.config = $state({} as Config);
    this.sponsoredRepos = $state([]);
    this.multiplierSponsoredRepos = $state([]);
    this.multiplierCountByRepo = $state(new Map());
    this.loadedSponsoredRepos = $state(false);
    this.route = $state("");
    this.accessToken = $state("");
    this.accessTokenExpire = $state(0);
    this.trustedRepos = $state([]);
    this.loadedTrustedRepos = $state(false);
    this.reposToUnTrustAfterTimeout = $state([]);
    this.reposInSearchResult = $state([]);
    this.loadedLatestThresholds = $state(false);
    this.latestThresholds = $state({} as HealthValueThreshold);
    this.reposWaitingForNewAnalysis = $state([]);
    this.reloadAdminSearchKey = $state(0);
    this.reloadHealthRepoCardKey = $state(0);
  }

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
  }

  setError(e: any) {
    this.error = e instanceof Error ? e.message : String(e);
  }

  getAccessToken() {
    return this.accessToken;
  }
}

export const appState = new AppState();
