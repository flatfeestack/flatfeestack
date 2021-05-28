<style>
main {
  text-align: center;
  padding: 1em;
  max-width: 240px;
  margin: 0 auto;
}
h1 {
  color: #ff3e00;
  text-transform: uppercase;
  font-size: 4em;
  font-weight: 100;
}
@media (min-width: 640px) {
  main {
    max-width: none;
  }
}
</style>

<script lang="ts">
import { ethers, providers } from "ethers";
import { ABI } from "./../types/contract";
import { error } from "../ts/store";

let storageContract
if (window.ethereum) {
  window.ethereum.enable();
  const provider = new providers.Web3Provider(window.ethereum);
  const signer = provider.getSigner();
  const contractAddress = "0x62Db0a2161e304819f4d54d54B90A3Feae6dDc72";
  storageContract = new ethers.Contract(contractAddress, ABI, signer);
} else {
  $error = 'Please install <a href="https://metamask.io/download.html">MetaMask</a>'
}

const requestFunds = async () => {
  storageContract.release();
};
</script>

<div class="container">
  <label class="px-2">Request funds:</label>
  <button class="button2" disabled={storageContract?"false":"true"} on:click="{requestFunds}">Request funds</button>
</div>
