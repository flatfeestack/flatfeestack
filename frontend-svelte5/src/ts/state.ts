import type { User, Config, Repo } from "../types/backend.ts";

type AppStateType = {
    error: string;
    isSubmitting: boolean;
    loginFailed: boolean;
    user: User;
    config: Config;
    sponsoredRepos: Repo[];
    loadedSponsoredRepos: boolean;
    route: string;
    accessToken: string;
    accessTokenExpire: string;
}

// Instead of stores, use state
export class AppState {
    $state: AppStateType = {
        error: "",
        isSubmitting: false,
        loginFailed: false,
        user: <User>{},
        config: <Config>{},
        sponsoredRepos: <Repo[]>[],
        loadedSponsoredRepos: false,
        route: "",
        accessToken: "",
        accessTokenExpire: ""
    };

    isAccessTokenExpired(): boolean {
        const expireTime = this.$state.accessTokenExpire;
        if (!expireTime) {
            return true;
        }

        try {
            const expirationSeconds = parseInt(expireTime);
            const currentSeconds = Math.floor(Date.now() / 1000);
            const bufferSeconds = 30; // 30 second buffer

            return currentSeconds + bufferSeconds >= expirationSeconds;
        } catch {
            return true;
        }
    }
    setError(e:any) {
        this.$state.error = e instanceof Error ? e.message : String(e);
    }
}

// Create and export a single instance
export const appState = new AppState();