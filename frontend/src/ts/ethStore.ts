import type { JsonRpcSigner, Web3Provider } from "@ethersproject/providers";
import { derived, writable, type Readable } from "svelte/store";

// provider is null when it's not initialized
// undefined when we did not detect any provider
// this case should be handled by the components themselves
export const provider = writable<Web3Provider | null | undefined>(null);
export const signer = writable<JsonRpcSigner | null>(null);

export const userEthereumAddress = derived<
  Readable<JsonRpcSigner | null>,
  string | null
>(
  signer,
  ($signer, set) => {
    if ($signer === null) {
      set(null);
    } else {
      Promise.resolve($signer.getAddress()).then((signerAddress: string) => {
        set(signerAddress);
      });
    }
  },
  null
);
