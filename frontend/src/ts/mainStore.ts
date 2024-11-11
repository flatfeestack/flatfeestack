import { writable } from "svelte/store";
import type { User, Config, Repo } from "../types/backend";

export const error = writable("");
export const isSubmitting = writable(false);

export const loginFailed = writable(false);
export const user = writable<User>(<User>{});
export const config = writable<Config>(<Config>{});
export const token = writable<string>("");
export const sponsoredRepos = writable<Repo[]>([]);
export const loadedSponsoredRepos = writable<boolean>(false);
export const route = writable("");
export const activeRootRoute = writable<string | null>(null);
export const trustedRepos = writable<Repo[]>([]);
export const loadedTrustedRepos = writable<boolean>(false);

