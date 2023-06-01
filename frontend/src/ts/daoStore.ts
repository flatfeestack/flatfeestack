import type { Web3Provider } from "@ethersproject/providers";
import { BigNumber, Contract, ethers } from "ethers";
import { derived, readable, type Readable } from "svelte/store";
import { DAOABI } from "../contracts/DAO";
import { MembershipABI } from "../contracts/Membership";
import { WalletABI } from "../contracts/Wallet";
import type { DaoConfig } from "../types/payout";
import { API } from "./api";
import { provider, signer, userEthereumAddress } from "./ethStore";

export const daoConfig = readable<DaoConfig | null>(null, (set) => {
  API.payout.daoConfig().then((daoConfig) => {
    set(daoConfig);
  });
});

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
  [daoConfig, provider, signer],
  ([$daoConfig, $provider, $signer]) => {
    if ($provider === null || $provider === undefined || $daoConfig === null) {
      return null;
    } else if ($signer === null) {
      return new ethers.Contract($daoConfig.dao, DAOABI, $provider);
    } else {
      return new ethers.Contract($daoConfig.dao, DAOABI, $signer);
    }
  }
);

export const membershipContract = derived(
  [daoConfig, provider, signer],
  ([$daoConfig, $provider, $signer]) => {
    if ($provider === null || $provider === undefined || $daoConfig === null) {
      return null;
    } else if ($signer === null) {
      return new ethers.Contract(
        $daoConfig.membership,
        MembershipABI,
        $provider
      );
    } else {
      return new ethers.Contract($daoConfig.membership, MembershipABI, $signer);
    }
  }
);

export const membershipStatusValue = derived<
  [Readable<string | null>, Readable<Contract | null>],
  BigNumber | null
>(
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
  string[] | null
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
  [daoConfig, provider, signer],
  ([$daoConfig, $provider, $signer]) => {
    if ($provider === null || $provider === undefined || $daoConfig === null) {
      return null;
    } else if ($signer === null) {
      return new ethers.Contract($daoConfig.wallet, WalletABI, $provider);
    } else {
      return new ethers.Contract($daoConfig.wallet, WalletABI, $signer);
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
