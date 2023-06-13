import { provider } from "../ts/ethStore";
import showMetaMaskRequired from "./showMetaMaskRequired";

export const checkUndefinedProvider = () =>
  provider.subscribe((providerValue) => {
    if (providerValue === undefined) {
      showMetaMaskRequired();
    }
  });
