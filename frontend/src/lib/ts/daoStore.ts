import type { JsonRpcSigner, Web3Provider } from "@ethersproject/providers";
import { BigNumber, Contract, ethers, Signer } from "ethers";
import { derived, writable, type Readable } from "svelte/store";
import { DAOABI } from "$lib/contracts/DAO";
import { MembershipABI } from "$lib/contracts/Membership";
import { WalletABI } from "$lib/contracts/Wallet";

// provider is null when it's not initialized
// undefined when we did not detect any provider
// this case should be handled by the components themselves
export const provider = writable<Web3Provider | null | undefined>(null);
export const signer = writable<JsonRpcSigner | null>(null);

export const currentBlockNumber = derived<
  Readable<null | Web3Provider>,
  number | null
>(
  provider,
  ($provider, set) => {
    if ($provider === null || $provider === undefined) {
      set(null);
    } else {
      $provider.getBlockNumber().then((blockNumber) => {
        set(blockNumber);
      });
    }
  },
  null
);

export const currentBlockTimestamp = derived<
  [Readable<null | Web3Provider>, Readable<number | null>],
  number | null
>(
  [provider, currentBlockNumber],
  ([$provider, $currentBlockNumber], set) => {
    if (
      $provider === null ||
      $provider === undefined ||
      $currentBlockNumber === null
    ) {
      set(null);
    } else {
      $provider.getBlock($currentBlockNumber).then((currentBlock) => {
        set(currentBlock.timestamp);
      });
    }
  },
  null
);

export const daoContract = derived(
  [provider, signer],
  ([$provider, $signer]) => {
    if ($provider === null || $provider === undefined) {
      return null;
    } else if ($signer === null) {
      return new ethers.Contract(
        import.meta.env.VITE_DAO_CONTRACT_ADDRESS,
        DAOABI,
        $provider
      );
    } else {
      return new ethers.Contract(
        import.meta.env.VITE_DAO_CONTRACT_ADDRESS,
        DAOABI,
        $signer
      );
    }
  }
);

export const userEthereumAddress = derived(
  signer,
  ($signer, set) => {
    if ($signer === null) {
      set(null);
    } else {
      Promise.resolve($signer.getAddress()).then((signerAddress: String) => {
        set(signerAddress);
      });
    }
  },
  null
);

export const membershipContract = derived(
  [provider, signer],
  ([$provider, $signer]) => {
    if ($provider === null || $provider === undefined || $signer === null) {
      return null;
    } else {
      return new ethers.Contract(
        import.meta.env.VITE_MEMBERSHIP_CONTRACT_ADDRESS,
        MembershipABI,
        $signer
      );
    }
  }
);

export const membershipStatusValue = derived(
  [userEthereumAddress, membershipContract],
  ([$userEthereumAddress, $membershipContract], set) => {
    if ($userEthereumAddress === null || $membershipContract === null) {
      set(null);
    } else {
      Promise.resolve(
        $membershipContract.getMembershipStatus($userEthereumAddress)
      ).then((membershipStatus) => {
        set(membershipStatus);
      });
    }
  },
  null
);

export const councilMembers = derived<
  Readable<Contract | null>,
  Signer[] | null
>(membershipContract, ($membershipContract, set) => {
  if ($membershipContract === null) {
    set(null);
  } else {
    Promise.resolve($membershipContract.getCouncilMembersLength()).then(
      (councilLength: BigNumber) => {
        Promise.all(
          [...Array(councilLength.toNumber()).keys()].map(
            async (index: Number) => {
              return await $membershipContract.councilMembers(index);
            }
          )
        ).then((councilMember) => {
          set(councilMember);
        });
      }
    );
  }
});

export const walletContract = derived(
  [provider, signer],
  ([$provider, $signer]) => {
    if ($provider === null || $provider === undefined || $signer === null) {
      return null;
    } else {
      return new ethers.Contract(
        import.meta.env.VITE_WALLET_CONTRACT_ADDRESS,
        WalletABI,
        $signer
      );
    }
  }
);

export const bylawsUrl = derived<Readable<Contract | null>, string | null>(
  daoContract,
  ($daoContract, set) => {
    if ($daoContract === null) {
      set(null);
    } else {
      // the empty bylaws URLs is a special scenario in the first week after DAO deployment
      // the DAO starts up without any bylaws attached, the first bylaws need to be confirmed in the first assembly
      // scheduled for a week after deployment
      $daoContract.bylawsUrl().then((retrievedBylawsUrl: string) => {
        set(retrievedBylawsUrl === "" ? null : retrievedBylawsUrl);
      });
    }
  }
);
