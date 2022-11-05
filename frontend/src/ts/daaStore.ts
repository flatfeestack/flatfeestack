import { writable } from "svelte/store";
import type { JsonRpcSigner } from "@ethersproject/providers"
import type {Web3Provider} from "@ethersproject/providers";

export const provider = writable<Web3Provider | null>(null);
export const membershipContract = writable<any | null>(null);
export const signer = writable<JsonRpcSigner | null>(null);
