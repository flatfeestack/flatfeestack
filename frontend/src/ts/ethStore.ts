import type { BrowserProvider, JsonRpcSigner, Network } from "ethers";
import { derived, writable, type Readable } from "svelte/store";

// provider is null when it's not initialized
// undefined when we did not detect any provider
// this case should be handled by the components themselves
export const provider = writable<BrowserProvider | null | undefined>(null);
export const signer = writable<JsonRpcSigner | null>(null);

export const chainId = derived<Readable<BrowserProvider | null>, number | null>(
  provider,
  ($provider, set) => {
    if ($provider === null) {
      set(null);
    } else {
      set(null);
      Promise.resolve($provider.getNetwork()).then((network: Network) => {
        set(Number(network.chainId));
      });
    }
  }
);

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
