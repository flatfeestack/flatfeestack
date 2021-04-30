import { writable } from "svelte/store";
import { Users, Repo, UserBalances, Config } from "../types/users";

export const showSignin = writable(false);
export const loading = writable(false);
export const user = writable<Users>(<Users>{});
export const config = writable<Config>(<Config>{});
export const userBalances = writable<UserBalances>(<UserBalances>{});
export const token = writable<string>("");
export const sponsoredRepos = writable<Repo[]>([]);
export const route = writable("")

