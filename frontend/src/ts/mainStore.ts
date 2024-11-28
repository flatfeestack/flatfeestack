import { writable } from "svelte/store";
import type {
  User,
  Config,
  Repo,
  HealthValueThreshold,
} from "../types/backend";

export const error = writable("");
export const isSubmitting = writable(false);

export const loginFailed = writable(false);
export const user = writable<User>(<User>{});
export const config = writable<Config>(<Config>{});
export const token = writable<string>("");
export const sponsoredRepos = writable<Repo[]>([]);
export const multiplierSponsoredRepos = writable<Repo[]>([]);
export const loadedSponsoredRepos = writable<boolean>(false);
export const loadedMultiplierRepos = writable<boolean>(false);
export const route = writable("");
export const trustedRepos = writable<Repo[]>([]);
export const loadedTrustedRepos = writable<boolean>(false);
export const reposToUnTrustAfterTimeout = writable<Repo[]>([]);
// export const undoDuration: number = 5000;
export const abortUntrustEvent = writable(false);
export const reposInSearchResult = writable<Repo[]>([]);
export const loadedLatestThresholds = writable<boolean>(false);
export const latestThresholds = writable<HealthValueThreshold>();
