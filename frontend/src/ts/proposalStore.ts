import type { BigNumber, Contract, Event } from "ethers";
import { derived, get, writable, type Readable } from "svelte/store";
import { daoContract } from "./daoStore";
import { userEthereumAddress } from "./ethStore";

export interface ProposalCreatedEvent {
  proposalId: string;
  event: Event;
}

const proposalEvents = writable<[] | ProposalCreatedEvent[]>([]);

export const proposalCreatedEvents = {
  // subscribe to the cart store
  subscribe: proposalEvents.subscribe,
  // custom logic
  async get(id: string, $daoContract): Promise<ProposalCreatedEvent> {
    let values = get(proposalEvents);
    const result = values.find(({ proposalId }) => proposalId === id);
    if (result) {
      return result;
    } else {
      return await Promise.resolve(
        $daoContract.queryFilter(
          $daoContract.filters.DAOProposalCreated(
            id,
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
        let newItem: ProposalCreatedEvent = {
          proposalId: events[0].args[0].toString(),
          event: events[0],
        };
        proposalEvents.update((items) => {
          return [...items, newItem];
        });
        return newItem;
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
