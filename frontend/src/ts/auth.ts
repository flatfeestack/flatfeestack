import { writable } from "svelte/store";
import { Users, PaymentCycle, Repo } from "../types/users";

export const loginFailed = writable(false);
export const loading = writable(false);
export const user = writable<Users>(<Users>{});
export const paymentCycle = writable<PaymentCycle>(<PaymentCycle>{});
export const token = writable<string>("");
export const sponsoredRepos = writable<Repo[]>([]);
export const route = writable("")

