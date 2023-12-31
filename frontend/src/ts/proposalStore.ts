import type { BytesLike, Contract, EventLog, Log } from "ethers";
import { derived, get, writable, type Readable } from "svelte/store";
import { daoContract } from "./daoStore";
import { userEthereumAddress } from "./ethStore";

export interface ProposalCreatedEvent {
  proposalId: string;
  event: Event;
}

export interface Proposal {
  calldatas: BytesLike[];
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
  async get(proposalId: string, daoContract: Contract): Promise<Proposal> {
    let values = get(proposals);
    const result = values.find(({ id }) => proposalId === id);

    if (result) {
      return result;
    } else {
      return await Promise.resolve(
        daoContract.queryFilter(
          daoContract.filters.DAOProposalCreated(
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
      ).then((events: Array<EventLog>) => {
        let newProposal: Proposal = {
          calldatas: events[0].args[5],
          description: events[0].args[8],
          endBlock: Number(events[0].args[7]),
          id: proposalId,
          proposer: events[0].args[1],
          signatures: events[0].args[4],
          startBlock: Number(events[0].args[6]),
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
        .then((votingSlotsLength: bigint) => {
          Promise.resolve(
            Promise.all(
              [...Array(Number(votingSlotsLength)).keys()].map((index) =>
                $daoContract.slots(index)
              )
            )
          ).then((slots: bigint[]) => {
            const sortedSlots = slots
              .map((slot) => Number(slot))
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
  null | bigint[]
>(
  daoContract,
  ($daoContract, set) => {
    if ($daoContract === null) {
      set(null);
    } else {
      Promise.resolve($daoContract.getExtraOrdinaryProposalsLength())
        .then((extraOrdinaryProposalsLength: bigint) => {
          Promise.resolve(
            Promise.all(
              [...Array(Number(extraOrdinaryProposalsLength)).keys()].map(
                (index) => $daoContract.extraOrdinaryAssemblyProposals(index)
              )
            )
          ).then((extraOrdinaryAssemblyRequestProposalIds: bigint[]) => {
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
  EventLog[] | null
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
      ).then((events: EventLog[]) => {
        set(events);
      });
    }
  },
  null
);
