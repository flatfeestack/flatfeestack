import { provider } from "../ts/daoStore";
import showMetaMaskRequired from "./showMetaMaskRequired";

const checkUndefinedProvider = () =>
  provider.subscribe((providerValue) => {
    if (providerValue === undefined) {
      showMetaMaskRequired();
    }
  });

export default checkUndefinedProvider;
