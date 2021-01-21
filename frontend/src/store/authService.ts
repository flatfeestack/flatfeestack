import { initialFetchDone, loading, refresh, token, user } from "./auth";
import { API } from "../api/api";

export const login = async (email: string, password: string) => {
  const res = await API.auth.login(email, password);
  const t = res.data.access_token;
  const r = res.data.refresh_token;
  if (!t || !r) {
    console.log("could not login");
    return;
  }
  token.set(t);
  localStorage.setItem("ffs-refresh", r);
  refresh.set(r);
  const u = await API.user.get();
  console.log(u);
  user.set(u.data);
};

export const updateUser = async () => {
  try {
    const u = await API.user.get();
    user.set(u.data);
  } catch (e) {
    console.log("could not fetch user", e);
  }
};

export const refreshSession = async () => {
  try {
    const r = localStorage.getItem("ffs-refresh");
    const res = await API.auth.refresh(r);
    const t = res.data.access_token;
    const new_r = res.data.refresh_token;
    token.set(t);
    refresh.set(new_r);
    await updateUser();
  } catch (e) {
    console.log("could not refresh session");
  }
};
