import type { BigNumber, Contract, Event } from "ethers";
import { derived, get, type Readable, writable } from "svelte/store";
import { daaContract } from "./daaStore";

interface ProposalCreatedEvent {
  proposalId: string;
  event: Event;
}

const proposalEvents = writable<[] | ProposalCreatedEvent[]>([]);

export const proposalCreatedEvents = {
  // subscribe to the cart store
  subscribe: proposalEvents.subscribe,
  // custom logic
  async get(id: string, $daaContract): Promise<ProposalCreatedEvent> {
    let values = get(proposalEvents);
    const result = values.find(({ proposalId }) => proposalId === id);
    if (result) {
      return result;
    } else {
      return await Promise.resolve(
        $daaContract.queryFilter(
          $daaContract.filters.DAAProposalCreated(
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

export const votingSlots = derived<Readable<null | Contract>, null | Number[]>(
  daaContract,
  ($daaContract, set) => {
    if ($daaContract === null) {
      set(null);
    } else {
      Promise.resolve($daaContract.getSlotsLength())
        .then((votingSlotsLength: BigNumber) => {
          Promise.resolve(
            Promise.all(
              [...Array(votingSlotsLength.toNumber()).keys()].map((index) =>
                $daaContract.slots(index)
              )
            )
          ).then((slots: BigNumber[]) => {
            set(slots.map((slot) => slot.toNumber()));
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
