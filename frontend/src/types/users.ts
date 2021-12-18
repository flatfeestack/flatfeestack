import {CryptoCurrency} from "./crypto";

export type Users = {
  id: string;
  payment_method: string;
  last4: string;
  paymentCycleId: string;
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
  restTimeout: number;
  plans: Plan[];
  env: string;
  contractAddr: string;
  supportedCurrencies: CryptoCurrency[];
}

export type Plan = {
  title:      string;
  price:      number;
  freq:       number;
  desc:       string;
  disclaimer: string;
  feePRm:     number;
}

export type ClientSecret = {
  client_secret: string;
}

export type UserBalances = {
  paymentCycle: PaymentCycle;
  userBalances: UserBalance[];
  total: TotalUserBalance[];
  daysLeft: number;
}

export type TotalUserBalance = {
  currency: string;
  balance: number;
}

export type UserBalanceCore = {
  userId: string;
  balance: number;
  currency: string;
}

export type UserBalance = {
  paymentCycleId: string;
  userId: string;
  balance: number;
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
  daysLeft: number;
}

export type Repo = {
  uuid: string
  id: number;
  html_url: string;
  clone_url: string;
  default_branch: string;
  full_name: string;
  description: string;
  tags: string;
  score: number;
  source: string;
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
  meta: string;
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
  userEmail: string;
  userName: string;
  repoName: string;
  contributorEmail: string;
  contributorWeight: number;
  isFlatFeeStackUser: boolean;
  balance: number;
  balanceRepo: number;
  day: string;
}

export type PayoutAddress = {
  id: string;
  currency: string;
  address: string;
}

export type PayoutInfo = {
  currency: string;
  amount: number;
}