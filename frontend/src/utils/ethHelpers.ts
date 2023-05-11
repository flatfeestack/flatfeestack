import { navigate } from "svelte-routing";
import { chainId, provider } from "../ts/ethStore";
import showMetaMaskRequired from "./showMetaMaskRequired";

export const checkUndefinedProvider = () =>
  provider.subscribe((providerValue) => {
    if (providerValue === undefined) {
      showMetaMaskRequired();
    }
  });

export const ensureSameChainId = (requiredChainId: number | undefined) => {
  if (requiredChainId === undefined) {
    return;
  }

  chainId.subscribe((chainIdValue) => {
    console.log(chainIdValue);

    if (chainIdValue !== null && chainIdValue !== requiredChainId) {
      navigate(
        `/differentChainId?required=${requiredChainId}&actual=${chainIdValue}`
      );
    }
  });
};
