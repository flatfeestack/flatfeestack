import type { User, Config, Repo } from "../types/backend.ts";

type AppStateType = {
    error: string;
    isSubmitting: boolean;
    user: User;
    config: Config;
    sponsoredRepos: Repo[];
    loadedSponsoredRepos: boolean;
    route: string;
    accessToken: string;
    accessTokenExpire: number;
}
// Instead of stores, use state
export class AppState {
    $state: AppStateType = {
        error: "",
        isSubmitting: false,
        user: <User>{},
        config: <Config>{},
        sponsoredRepos: <Repo[]>[],
        loadedSponsoredRepos: false,
        route: "",
        accessToken: "",
        accessTokenExpire: 0
    };

    isAccessTokenExpired(): boolean {
        const expireTime = this.$state.accessTokenExpire;
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
        this.$state.error = e instanceof Error ? e.message : String(e);
    }
}
// Create and export a single instance
export const appState = new AppState();


