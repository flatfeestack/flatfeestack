import { writable } from "svelte/store";
import { Users, PaymentCycle, Repo, UserBalances } from "../types/users";

export const showSignin = writable(false);
export const loading = writable(false);
export const user = writable<Users>(<Users>{});
export const userBalances = writable<UserBalances>(<UserBalances>{});
export const token = writable<string>("");
export const sponsoredRepos = writable<Repo[]>([]);
export const route = writable("")

