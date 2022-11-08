import { writable } from "svelte/store";
import type { Users, UserBalances, Config, Repos } from "../types/users";

export const error = writable("");
export const isSubmitting = writable(false);

export const loginFailed = writable(false);
export const user = writable<Users>(<Users>{});
export const config = writable<Config>(<Config>{});
export const userBalances = writable<UserBalances>(<UserBalances>{});
export const token = writable<string>("");
export const sponsoredRepos = writable<Repos[]>([]);
export const loadedSponsoredRepos = writable<boolean>(false);
export const route = writable("");
