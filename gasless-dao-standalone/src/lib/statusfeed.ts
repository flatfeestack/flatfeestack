import { writable } from "svelte/store";

export const statusFeed = writable<string[]>([]);

function timestamp(): string {
    return new Date().toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
    });
}

export function pushStatus(message: string) {
    const entry = `[${timestamp()}] ${message}`;
    statusFeed.update((list) => [...list, entry]);
    console.log("Wrote Status: " + entry);
}

export function clearStatusFeed() {
    statusFeed.set([]);
}
