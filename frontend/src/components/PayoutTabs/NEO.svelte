<script lang="ts">
  import { ethers, providers } from "ethers";
  import { ABI } from "../../types/contract";
  import { error, user, config } from "../../ts/mainStore";
  import detectEthereumProvider from "@metamask/detect-provider";
  import { onMount } from "svelte";
  import Spinner from "../Spinner.svelte";
  import Dots from "../Dots.svelte";

  let storageContract;
  let viewContract;
  let balance = 0;
  let neoline;
  let account;

  const handleNeoLineEvent = () => {
    neoline = new window.NEOLineN3.Init()
    $error = "Please install <a href=\"https://neoline.io/en/\">NeoLine</a>";
  }

  const initNeolineAccount = async () => {
    try {
      account = await neoline.getAccount();
      /* Example */
      neoline.invoke({
        scriptHash: '0x8e5823c86999f283e440bb2e78f0a0c5784fc9bd',
        operation: 'balanceOf',
        args: [
          {
            type: 'Hash160',
            value: '91b83e96f2a7c4fdf0c1688441ec61986c7cae26'
          }
        ],
        attachedAssets: {
          NEO: '1',
          GAS: '0.0001'
        },
        fee: '0.001',
        network: 'TestNet',
        broadcastOverride: false,
        txHashAttributes: [
          {
            type: 'Boolean',
            value: true,
            txAttrUsage: 'Hash1'
          }
        ]
      })
      .then(result => {
        console.log('Invoke transaction success!');
        console.log('Transaction ID: ' + result.txid);
        console.log('RPC node URL: ' + result.nodeURL);
      })
      .catch((error) => {
        const {type, description, data} = error;
        switch(type) {
          case 'NO_PROVIDER':
            console.log('No provider available.');
            break;
          case 'RPC_ERROR':
            console.log('There was an error when broadcasting this transaction to the network.');
            break;
          case 'CANCELED':
            console.log('The user has canceled this transaction.');
            break;
          default:
            // Not an expected error object.  Just write the error to the console.
            console.error(error);
            break;
        }
      });

    } catch (error) {
      console.log({ message: "NeoLine", description: error.type });
    }
  };

  const requestFunds = async () => {
    try {
      await initNeolineAccount()
      // await storageContract.release();
    } catch (e) {
      $error = e;
    }
  };
</script>

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
<svelte:window on:NEOLine.NEO.EVENT.READY={handleNeoLineEvent}/>

<div class="container">
  <label class="px-2">Request funds:</label>
  {#await balance}
    <Dots /> NEO
  {:then res}
    {res} NEO
  {:catch err}
    {$error = err}
  {/await}
  <button class="button2" on:click="{requestFunds}">Request funds</button>
</div>
