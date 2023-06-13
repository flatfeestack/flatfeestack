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

// usually, we have the EnsureChainId component that listens and redirects if the chain ids do not match
// however, in certain cases, we want to have a blocking function
// like in the Income component, where we want to make sure that the chain id is the same before preparing the transaction to withdraw funds
export function getChainId(): Promise<number | undefined> {
  const timeoutPromise = new Promise<undefined>((resolve) =>
    setTimeout(() => resolve(undefined), 5000)
  );
  const functionPromise = new Promise<number>((resolve) => {
    chainId.subscribe((chainIdValue) => {
      if (chainIdValue !== null) {
        resolve(chainIdValue);
      }
    });
  });

  return Promise.race([functionPromise, timeoutPromise]);
}

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
