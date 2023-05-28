import { get, writable } from "svelte/store";
import type { PublicUser } from "../types/backend";
import { API } from "./api";

const usersStore = writable<PublicUser[]>([
  {
    id: "00000000-0000-0000-0000-000000000000",
    image: null,
    name: "System",
  },
]);

export const users = {
  subscribe: usersStore.subscribe,
  async get(userId: string): Promise<PublicUser> {
    const values = get(usersStore);
    const result = values.find(({ id }) => userId === id);
    if (result) {
      return result;
    } else {
      return await Promise.resolve(
        API.user.getUser(userId).then((user: PublicUser) => {
          usersStore.update((items) => {
            return [...items, user];
          });
          return user;
        })
      );
    }
  },
};
