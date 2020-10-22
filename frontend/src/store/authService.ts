import { loading, token, user } from "./auth";
import { API } from "../api/api";

export const tryToAuthenticate = async () => {
  try {
    loading.set(true);
    const res = await API.auth.refresh();
    const t = res.headers["token"];
    if (!t) {
      loading.set(false);
      return;
    }
    token.set(t);
    const u = await API.api.user.get();
    user.set(u.data.data);
  } catch (e) {
    console.log("fetch user service error", String(e));
  }
};

export const login = async (email: string, password: string) => {
  try {
    const res = await API.auth.login(email, password);
    const t = res.headers["token"];
    if (t) {
      token.set(t);
    }
    const u = await API.api.user.get();
    console.log(u);
    user.set(u.data.data);
  } catch (e) {
    console.log(e);
  }
};
