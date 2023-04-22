import { Web3Provider } from "@ethersproject/providers";
import detectEthereumProvider from "@metamask/detect-provider";
import { signer } from "../ts/ethStore";
import showMetaMaskRequired from "./showMetaMaskRequired";

async function setSigner(providerValue: Web3Provider | null) {
  if (providerValue === null || providerValue === undefined) {
    try {
      const ethProv = await detectEthereumProvider();
      providerValue = new Web3Provider(<any>ethProv);
    } catch (error) {
      console.error(error);
      providerValue = undefined;
    }
  }

  if (providerValue === undefined) {
    showMetaMaskRequired();
  } else {
    await providerValue.send("eth_requestAccounts", []);
    signer.set(providerValue.getSigner());
  }
}

export default setSigner;
