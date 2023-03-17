import type { JsonRpcSigner, Web3Provider } from "@ethersproject/providers";
import { BigNumber, Contract, ethers, Signer } from "ethers";
import { derived, writable, type Readable } from "svelte/store";
import { DAOABI } from "../contracts/DAO";
import { MembershipABI } from "../contracts/Membership";
import { WalletABI } from "../contracts/Wallet";

export const provider = writable<Web3Provider | null>(null);
export const signer = writable<JsonRpcSigner | null>(null);

export const currentBlockNumber = derived<
  Readable<null | Web3Provider>,
  number | null
>(
  provider,
  ($provider, set) => {
    if ($provider === null) {
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
    if ($provider === null || $currentBlockNumber === null) {
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
    if ($provider === null) {
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
    if ($provider === null || $signer === null) {
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
    if ($provider === null || $signer === null) {
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
