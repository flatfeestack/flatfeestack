import type { PaymentCycle } from "./backend";

export type UserBalances = {
  paymentCycle: PaymentCycle;
  userBalances: UserBalance[];
  total: Map<string, bigint>;
  daysLeft: number;
};

export type UserBalance = {
  paymentCycleId: string;
  userId: string;
  balance: bigint;
  currency: string;
  balanceType: string;
  createdAt: string;
};
