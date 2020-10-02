import { derived, writable } from "svelte/store";

export const token = writable("");

export const loggedIn = derived(token, ($token) => $token !== "");
