import type { BigNumber, Bytes, Contract, Event } from "ethers";
import { derived, get, writable, type Readable } from "svelte/store";
import { daoContract } from "./daoStore";
import { userEthereumAddress } from "./ethStore";

export interface ProposalCreatedEvent {
  proposalId: string;
  event: Event;
}

export interface Proposal {
  calldatas: Bytes[];
  description: string;
  endBlock: number;
  id: string;
  proposer: string;
  signatures: string[];
  startBlock: number;
  targets: string[];
  values: string[];
}

const proposals = writable<Proposal[]>([]);

export const proposalStore = {
  // subscribe to the cart store
  subscribe: proposals.subscribe,
  // custom logic
  async get(proposalId: string, $daoContract): Promise<Proposal> {
    let values = get(proposals);
    const result = values.find(({ id }) => proposalId === id);

    if (result) {
      return result;
    } else {
      return await Promise.resolve(
        $daoContract.queryFilter(
          $daoContract.filters.DAOProposalCreated(
            proposalId,
            null,
            null,
            null,
            null,
            null,
            null,
            null,
            null,
            null
          )
        )
      ).then((events: Event[]) => {
        let newProposal: Proposal = {
          calldatas: events[0].args[5],
          description: events[0].args[8],
          endBlock: events[0].args[7].toNumber(),
          id: proposalId,
          proposer: events[0].args[1],
          signatures: events[0].args[4],
          startBlock: events[0].args[6].toNumber(),
          targets: events[0].args[2],
          values: events[0].args[3],
        };
        proposals.update((items) => {
          return [...items, newProposal];
        });
        return newProposal;
      });
    }
  },
};

export const votingSlots = derived<Readable<null | Contract>, null | number[]>(
  daoContract,
  ($daoContract, set) => {
    if ($daoContract === null) {
      set(null);
    } else {
      Promise.resolve($daoContract.getSlotsLength())
        .then((votingSlotsLength: BigNumber) => {
          Promise.resolve(
            Promise.all(
              [...Array(votingSlotsLength.toNumber()).keys()].map((index) =>
                $daoContract.slots(index)
              )
            )
          ).then((slots: BigNumber[]) => {
            const sortedSlots = slots
              .map((slot) => slot.toNumber())
              .sort((slot1, slot2) => slot2 - slot1); // sort descending, slot with latest start to appear first

            set(sortedSlots);
          });
        })
        .catch((reason) => {
          console.error(reason);
          set(null);
        });
    }
  },
  null
);

export const extraOrdinaryAssemblyRequestProposalIds = derived<
  Readable<null | Contract>,
  null | BigNumber[]
>(
  daoContract,
  ($daoContract, set) => {
    if ($daoContract === null) {
      set(null);
    } else {
      Promise.resolve($daoContract.getExtraOrdinaryProposalsLength())
        .then((extraOrdinaryProposalsLength: BigNumber) => {
          Promise.resolve(
            Promise.all(
              [...Array(extraOrdinaryProposalsLength.toNumber()).keys()].map(
                (index) => $daoContract.extraOrdinaryAssemblyProposals(index)
              )
            )
          ).then((extraOrdinaryAssemblyRequestProposalIds: BigNumber[]) => {
            set(extraOrdinaryAssemblyRequestProposalIds);
          });
        })
        .catch((reason) => {
          console.error(reason);
          set(null);
        });
    }
  },
  null
);

export const votesCasted = derived<
  [Readable<Contract | null>, Readable<string | null>],
  Event[] | null
>(
  [daoContract, userEthereumAddress],
  ([$daoContract, $userEthereumAddress], set) => {
    if ($daoContract === null || $userEthereumAddress === null) {
      set(null);
    } else {
      Promise.resolve(
        $daoContract.queryFilter(
          $daoContract.filters.VoteCast(
            $userEthereumAddress,
            null,
            null,
            null,
            null
          )
        )
      ).then((events: Event[]) => {
        set(events);
      });
    }
  },
  null
);
