import { writable } from "svelte/store";
import { User, PaymentCycle, Repo } from "../types/user";

export const loading = writable(false);
export const user = writable<User>(<User>{});
export const paymentCycle = writable<PaymentCycle>(<PaymentCycle>{});
export const token = writable<string>("");
export const sponsoredRepos = writable<Repo[]>([]);

