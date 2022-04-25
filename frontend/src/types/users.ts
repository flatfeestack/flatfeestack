import type {ChartData} from "chart.js";

export type Users = {
  id: string;
  paymentMethod: string;
  last4: string;
  paymentCycleInId: string;
  paymentCycleOutId: string;
  email: string;
  name: string;
  image: string;
  username: string;
  token: string;
  login: boolean;
  role: string;
};

export type Config = {
  stripePublicApi: string;
  wsBaseUrl: string;
  plans: Plan[];
  env: string;
  contractAddr: string;
  supportedCurrencies: Map<string, Currencies>;
}

export type Plan = {
  title:      string;
  price:      number;
  freq:       number;
  desc:       string;
  disclaimer: string;
  feePrm:     number;
}

export type ClientSecret = {
  clientSecret: string;
}

export type UserBalances = {
  paymentCycle: PaymentCycle;
  userBalances: UserBalance[];
  total: Map<string, bigint>;
  daysLeft: number;
}

export type UserBalanceCore = {
  userId: string;
  balance: bigint;
  currency: string;
}

export type UserBalance = {
  paymentCycleId: string;
  userId: string;
  balance: bigint;
  currency: string;
  balanceType: string;
  createdAt: string;
}

export type UserStatus = {
  userId: string;
  name: string;
  email: string;
  daysLeft: number;
}

export type PaymentCycle = {
  id: string;
  seats: number;
  freq: number;
}

export type Repos = {
  uuid: string;
  repos: Repo[];
  balances: Map<string, bigint>;
}

export type Repo = {
  uuid: string;
  id: number; //this comes from the github search
  url: string;
  gitUrl: string;
  name: string;
  description: string;
  score: number;
  source: string;
  link: string|null;
};

export type Token = {
  access_token: string
  refresh_token: string
  expire: string
}

export type GitUser = {
  email: string;
  confirmedAt: string|null;
  createdAt: string|null;
};

export type Invitation = {
  email: string;
  inviteEmail: string;
  confirmedAt: string|null;
  createdAt: string;
};

export type Time = {
  time: string;
}

export type UserAggBalance = {
  balance: number;
  email_list: string[];
  daily_user_payout_id_list: string[];
}

export type RepoMapping = {
  startDate: string;
  endDate: string;
  name: string;
  weights: FlatFeeWeight;
}

export type FlatFeeWeight = {
  email: string;
  weight: number;
}

export type Contributions = {
  repoName: string;
  repoUrl: string;
  sponsorName: string;
  sponsorEmail: string;
  contributorName: string;
  contributorEmail: string;
  balance: bigint;
  currency: string;
  paymentCycleInId: string;
  day: string;
}

export type FilteredContributions = {
  repoName: string;
  repoUrl: string;
  sponsorName: string[];
  sponsorEmail: string[];
  contributorName: string[];
  contributorEmail: string[];
  balance: bigint;
  currency: string;
  paymentCycleInId: string;
  dayFrom: string;
  dayTo: string;
}

export type PayoutAddress = {
  id: string;
  currency: string;
  address: string;
}

export type PayoutInfo = {
  currency: string;
  amount: bigint;
}

export type Currencies = {
  name: string;
  short: string;
  smallest: string;
  factorPow: number;
  isCrypto: boolean;
}

export type PaymentResponse = {
  payAddress: string;
  payAmount: bigint;
  payCurrency: string;
}

export type ChartDataTotal = ChartData & {
  total: number;
  days: number;
}