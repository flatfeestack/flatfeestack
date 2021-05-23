export type Users = {
  id: string;
  payment_method: string;
  last4: string;
  paymentCycleId: string;
  email: string;
  name: string;
  image: string;
  username: string;
  payout_eth: string;
  token: string;
  role: string;
  login: boolean;
};

export type Config = {
  stripePublicApi: string;
  wsBaseUrl: string;
  restTimeout: number;
  plans: Plan[];
  env: string;
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
  total: number;
  daysLeft: number;
}

export type UserBalance = {
  paymentCycleId: string;
  userId: string;
  balance: number;
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
  payout_eth: string;
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
  userId: string;
  userEmail: string;
  repoId: string;
  repoName: string;
  contributorEmail: string;
  contributorWeight: number;
  contributorUserId: string|null;
  balance: number;
  balanceRepo: number;
  day: string;
}
