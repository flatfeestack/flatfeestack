import { appState } from "ts/state.svelte.ts";
import {API} from "ts/api.ts";
import type {Token, User} from "types/backend.ts";

export const login = async (email: string):Promise<boolean> => {
    const refreshToken = getRefreshToken();
    if (refreshToken !== null) {
        if (!appState.isAccessTokenExpired()) {
            if(appState.user.email === email) {
                return true;
            }
        }
        try {
            const { access_token, expires_at } = await refresh(refreshToken, email);
            appState.accessToken = access_token;
            appState.accessTokenExpire = expires_at;
            return true;
        } catch (e: unknown) {
            //this can fail, if it does go on and send out email
        }
    }
    //here we actually need to do a login and send out the email
    const token = await API.auth.login(email);
    if (typeof token === 'object' && token !== null) {
        storeToken(token);
        return true;
    }
    return false;
};

export const confirm = async (email: string, emailToken: string) => {
    const token = await API.auth.confirm(email, emailToken);
    storeToken(token);
    appState.user  = await API.user.get();
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
    appState.user = <User>{};
    appState.accessToken = "";
    appState.accessTokenExpire = 0;
}

export function storeToken (token: Token){
    appState.accessToken = token.access_token;
    appState.accessTokenExpire = token.expires_at;
    localStorage.setItem("ffs-refresh", token.refresh_token);
}

export function getRefreshToken() {
    return localStorage.getItem("ffs-refresh");
}
