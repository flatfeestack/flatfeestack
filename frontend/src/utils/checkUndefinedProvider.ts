import { provider } from "../ts/ethStore";
import showMetaMaskRequired from "./showMetaMaskRequired";

const checkUndefinedProvider = () =>
  provider.subscribe((providerValue) => {
    if (providerValue === undefined) {
      showMetaMaskRequired();
    }
  });

export default checkUndefinedProvider;
