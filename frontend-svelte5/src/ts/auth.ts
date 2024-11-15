import { appState } from "ts/state.ts";
import { API } from "ts/api.ts";
import type { User } from "types/backend";
import type { Token } from "types/auth";

export const confirmReset = async (
    email: string,
    password: string,
    emailToken: string
) => {
    const p1 = API.auth.confirmReset(email, password, emailToken);
    const p2 = API.config.config();

    const res = await p1;
    storeToken(res);

    const p3 = API.user.get();

    appState.$state.config = await p2;
    appState.$state.user = await p3;
};

export const confirmEmail = async (email: string, emailToken: string) => {
    const p1 = API.auth.confirmEmail(email, emailToken);
    const p2 = API.config.config();

    const res = await p1;
    storeToken(res);

    const p3 = API.user.get();

    appState.$state.config = await p2;
    appState.$state.user  = await p3;
};

export const confirmInvite = async (
    email: string,
    password: string,
    emailToken: string,
    inviteByEmail: string
) => {
    const p1 = API.auth.confirmInvite(email, password, emailToken);
    const p2 = API.config.config();

    const res = await p1;
    storeToken(res);

    const p3 = API.invite.confirmInvite(inviteByEmail);
    const p4 = API.user.get();
    appState.$state.config = await p2;
    await p3;
    appState.$state.user = await p4;
};

export const login = async (email: string, password: string) => {
    const p1 = API.auth.login(email, password);
    const p2 = API.config.config();

    const res = await p1;
    storeToken(res);

    appState.$state.config = await p2;
    appState.$state.user = await API.user.get();
};

export const refresh = async (): Promise<string> => {
    const p2 = API.config.config();
    const refresh = localStorage.getItem("ffs-refresh");
    if (refresh === null) {
        throw "No refresh token not in local storage";
    }
    const p1 = API.auth.refresh(refresh);

    const token = await p1;
    appState.$state.config = await p2;
    storeToken(token);
    return token.access_token;
};

export const removeSession = async () => {
    try {
        await API.authToken.logout();
    } finally {
        removeToken();
    }
};

export const removeToken = () => {
    localStorage.removeItem("ffs-refresh");
    localStorage.removeItem("ffs-access-expiry");
    appState.$state.user = <User>{};
    appState.$state.accessToken = "";
    appState.$state.accessTokenExpire = "";
    appState.$state.loginFailed = true;
};

export const storeToken = (token1: Token) => {
    if (!token1 || !token1.access_token || !token1.refresh_token || !token1.expire) {
        appState.$state.loginFailed = true;
        throw "Invalid/empty token in the request";
    }
    appState.$state.loginFailed = false;
    appState.$state.accessToken = token1.access_token;
    appState.$state.accessTokenExpire = token1.expire;
    localStorage.setItem("ffs-refresh", token1.refresh_token);
};

export const hasToken= () => {
    const refresh = localStorage.getItem("ffs-refresh");
    return refresh !== null
}
