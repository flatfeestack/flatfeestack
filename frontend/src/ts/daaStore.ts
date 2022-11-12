import { derived, readable, writable } from "svelte/store";
import type { JsonRpcSigner } from "@ethersproject/providers";
import type { Web3Provider } from "@ethersproject/providers";
import { ethers } from "ethers";
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

export const ethereumAddress = derived(
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
  [ethereumAddress, membershipContract],
  ([$ethereumAddress, $membershipContract], set) => {
    if ($ethereumAddress === null || $membershipContract === null) {
      set(null);
    } else {
      Promise.resolve(
        $membershipContract.getMembershipStatus($ethereumAddress)
      ).then((membershipStatus) => {
        set(membershipStatus);
      });
    }
  },
  null
);
