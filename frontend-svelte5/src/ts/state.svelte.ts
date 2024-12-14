import type {User, Config, Repo, HealthValueThreshold} from "../types/backend.ts";

// Instead of stores, use state
export class AppState {
    error = $state("");
    isSubmitting = $state(false);
    user= $state(<User>{});
    config = $state(<Config>{});
    sponsoredRepos=  $state(<Repo[]>[]);
    multiplierSponsoredRepos=  $state(<Repo[]>[]);
    loadedSponsoredRepos= $state(false);
    route= $state("");
    accessToken= $state("");
    accessTokenExpire= $state(0);

    trustedRepos = $state(<Repo[]>[]);
    loadedTrustedRepos = $state(false);
    reposToUnTrustAfterTimeout = $state(<Repo[]>[]);
    reposInSearchResult = $state(<Repo[]>[]);
    loadedLatestThresholds = $state(false);
    latestThresholds =  $state(<HealthValueThreshold>{});
    reposWaitingForNewAnalysis = $state(<Repo[]>[]);
    reloadAdminSearchKey = $state(0);
    reloadHealthRepoCardKey = $state(0);

    isAccessTokenExpired(): boolean {
        const expireTime = this.accessTokenExpire;
        if (!expireTime) {
            return true;
        }

        try {
            const currentSeconds = Math.floor(Date.now() / 1000);
            const bufferSeconds = 30; // 30 second buffer

            return currentSeconds + bufferSeconds >= expireTime;
        } catch {
            return true;
        }
    }
    setError(e:any) {
        this.error = e instanceof Error ? e.message : String(e);
    }

    getAccessToken() {
        return this.accessToken;
    }

}
// Create and export a single instance
export const appState = new AppState();
