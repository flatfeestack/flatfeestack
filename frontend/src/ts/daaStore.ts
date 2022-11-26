import { derived, readable, writable, type Readable } from "svelte/store";
import type { JsonRpcSigner } from "@ethersproject/providers";
import type { Web3Provider } from "@ethersproject/providers";
import { BigNumber, Contract, ethers, Signer } from "ethers";
import { MembershipABI } from "../contracts/Membership";
import { DAAABI } from "../contracts/DAA";

export const provider = writable<Web3Provider | null>(null);
export const signer = writable<JsonRpcSigner | null>(null);

export const daaContract = derived(
  [provider, signer],
  ([$provider, $signer]) => {
    if ($provider === null) {
      return null;
    } else if ($signer === null) {
      return new ethers.Contract(
        import.meta.env.VITE_DAA_CONTRACT_ADDRESS,
        DAAABI,
        $provider
      );
    } else {
      return new ethers.Contract(
        import.meta.env.VITE_DAA_CONTRACT_ADDRESS,
        DAAABI,
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

export const chairmanAddress = derived<
  Readable<Contract | null>,
  Signer | null
>(membershipContract, ($membershipContract, set) => {
  if ($membershipContract === null) {
    set(null);
  } else {
    Promise.resolve($membershipContract.chairman()).then((chairmanAddress) => {
      set(chairmanAddress);
    });
  }
});

export const whitelisters = derived<Readable<Contract | null>, Signer[] | null>(
  membershipContract,
  ($membershipContract, set) => {
    if ($membershipContract === null) {
      set(null);
    } else {
      Promise.resolve($membershipContract.whitelisterListLength()).then(
        (whitelisterLength: BigNumber) => {
          Promise.all(
            [...Array(whitelisterLength.toNumber()).keys()].map(
              async (index: Number) => {
                return await $membershipContract.whitelisterList(index);
              }
            )
          ).then((whitelisters) => {
            set(whitelisters);
          });
        }
      );
    }
  }
);
