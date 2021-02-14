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
import { onMount } from "svelte";
//import Web3 from "web3";
import { Contract, providers } from 'ethers';
//import { toHex, toWei } from "web3-utils";
//import { Common } from "web3-core";
import { ABI, CONTRACT_ID } from "../types/contract";
//const web3 = new Web3(
//  "https://ropsten.infura.io/v3/6d6c0e875d6c4becaec0e1b10d5bc3cc"
//);

let metamask = null;
const ethEnabled = () => {
/*  if (window.ethereum) {
    metamask = new Web3(window.ethereum);
    try {
      window.ethereum.enable();
      return true;
    } catch (e) {
      console.log("No MetaMask");
      return false;
    }
  }*/
  const provider = new providers.Web3Provider(window.ethereum);
  const signer = provider.getSigner();
  const contractAddress = "0x62Db0a2161e304819f4d54d54B90A3Feae6dDc72";
  const storageContract = new Contract(contractAddress, ABI, signer);
  return false;
};

const requestFunds = async () => {
  ethEnabled();
  const accounts = await metamask.eth.getAccounts();
  console.log("accounts", accounts[0]);

  const contract = new metamask.eth.Contract(ABI, CONTRACT_ID);
  contract.methods.release().send({ from: accounts[0] });
  /*  const res = await web3.eth.sendTransaction({
    gasPrice: toHex(toWei("5", "gwei")),
    gas: toHex("21000"),
    from: "0x2A9A56c5a16e7e0219BB7125C01ec4dF3105A502",
    to: "0x573ac7FFeeaBb98fDD2c0e05668637BBd8D6A104",
    value: toHex(toWei("1", "finney")),
  });*/
};
</script>

<button class="button" on:click="{requestFunds}">Request funds</button>
