import { writable } from "svelte/store";
import { Repo } from "../types/repo.type";

export const sponsoredRepos = writable<Repo[]>([]);
export const initialFetch = writable(false);
