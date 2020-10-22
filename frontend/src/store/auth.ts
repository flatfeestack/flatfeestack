import { derived, writable } from "svelte/store";
import { User } from "../types/user";

export const loading = writable(false);
export const user = writable<User | null>(null);
export const token = writable<string>("");
