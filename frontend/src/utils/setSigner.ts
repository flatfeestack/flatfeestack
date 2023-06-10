import detectEthereumProvider from "@metamask/detect-provider";
import { signer } from "../ts/ethStore";
import showMetaMaskRequired from "./showMetaMaskRequired";
import { BrowserProvider } from "ethers";

async function setSigner(providerValue: BrowserProvider | null) {
  if (providerValue === null || providerValue === undefined) {
    try {
      const ethProv = await detectEthereumProvider();
      providerValue = new BrowserProvider(<any>ethProv);
    } catch (error) {
      console.error(error);
      providerValue = undefined;
    }
  }

  if (providerValue === undefined) {
    showMetaMaskRequired();
  } else {
    await providerValue.send("eth_requestAccounts", []);
    signer.set(await providerValue.getSigner());
  }
}

export default setSigner;
