import { provider } from "$lib/ts/daoStore";
import showMetaMaskRequired from "$lib/utils/showMetaMaskRequired";

const checkUndefinedProvider = () =>
  provider.subscribe((providerValue) => {
    if (providerValue === undefined) {
      showMetaMaskRequired();
    }
  });

export default checkUndefinedProvider;
