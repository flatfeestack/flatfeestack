import { appState } from "ts/state.ts";
import {API} from "ts/api.ts";
import type { User } from "types/backend";
import type { Token } from "types/auth";
import {HTTPError} from "ky";

export const login = async (email: string):Promise<boolean> => {
    const refreshToken = getRefreshToken();
    if (refreshToken) {
        if (!appState.isAccessTokenExpired()) {
            return appState.$state.user.email === email;
        }
        try {
            const { access_token, expires_at } = await refresh(refreshToken, email);
            appState.$state.accessToken = access_token;
            appState.$state.accessTokenExpire = expires_at;
            return true;
        } catch (e: unknown) {
            if (!(e instanceof HTTPError) || e.response.status !== 404) {
                throw e;
            }
        }
    }
    //here we actually need to do a login and send out the email
    const token = await API.auth.login(email);
    if (token) {
        storeToken(token);
        return true;
    }
    return false;
};

export const confirm = async (email: string, emailToken: string) => {
    const token = await API.auth.confirm(email, emailToken);
    storeToken(token);
    appState.$state.user  = await API.user.get();
};

export async function refresh(refreshToken: string, email: string = ""): Promise<{ access_token: string, expires_at: number }> {
    const token = await API.auth.refresh(refreshToken, email);
    storeToken(token);
    return { access_token: token.access_token, expires_at: token.expires_at };
}

export async function removeSession(){
    try {
        await API.authToken.logout();
    } finally {
        removeToken();
    }
}

export function removeToken(){
    localStorage.removeItem("ffs-refresh");
    appState.$state.user = <User>{};
    appState.$state.accessToken = "";
    appState.$state.accessTokenExpire = 0;
}

export function storeToken (token: Token){
    appState.$state.accessToken = token.access_token;
    appState.$state.accessTokenExpire = token.expires_at;
    localStorage.setItem("ffs-refresh", token.refresh_token);
}

export function getRefreshToken() {
    return localStorage.getItem("ffs-refresh");
}
