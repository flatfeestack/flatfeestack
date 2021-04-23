import { writable } from "svelte/store";
import { Types, PaymentCycle, Repo } from "../types/types";

export const loading = writable(false);
export const user = writable<Types>(<Types>{});
export const paymentCycle = writable<PaymentCycle>(<PaymentCycle>{});
export const token = writable<string>("");
export const sponsoredRepos = writable<Repo[]>([]);
export const route = writable("")

