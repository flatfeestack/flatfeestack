import { derived, writable } from "svelte/store";
import { User } from "../types/user";

export const initialFetchDone = writable(false);
export const loading = writable(false);
export const user = writable<User>(<User>{});
export const token = writable<string>("");
