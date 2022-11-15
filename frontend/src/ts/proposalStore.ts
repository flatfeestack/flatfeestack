import type { BigNumber, Contract, Event } from "ethers";
import { derived, type Readable } from "svelte/store";
import { daaContract } from "./daaStore";

export const proposalCreatedEvents = derived<
  Readable<null | Contract>,
  null | Event[]
>(
  daaContract,
  ($daaContract, set) => {
    if ($daaContract === null) {
      set(null);
    } else {
      Promise.resolve(
        $daaContract.queryFilter(
          $daaContract.filters.ProposalCreated(
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
        set(events);
      });
    }
  },
  null
);

export const votingSlots = derived<Readable<null | Contract>, null | Number[]>(
  daaContract,
  ($daaContract, set) => {
    if ($daaContract === null) {
      set(null);
    } else {
      Promise.resolve($daaContract.getSlotsLength()).then(
        (votingSlotsLength: BigNumber) => {
          Promise.resolve(
            Promise.all(
              [...Array(votingSlotsLength.toNumber()).keys()].map((index) =>
                $daaContract.slots(index)
              )
            )
          ).then((slots: BigNumber[]) => {
            set(slots.map((slot) => slot.toNumber()));
          });
        }
      );
    }
  },
  null
);
