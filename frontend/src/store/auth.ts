import { derived, writable } from "svelte/store";
import { User } from "../types/user";

export const initialFetchDone = writable(false);
export const loading = writable(false);
export const user = writable<User | null>(null);
export const token = writable<string>("");
export const refresh = writable<string>("");
