export type User = {
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
};

export type ClientSecret = {
  client_secret: string;
}

export type UserBalances = {
  paymentCycleId: string
  userBalances: UserBalance[]
  total: number
}

export type UserBalance = {
  paymentCycleId: string;
  userId: string;
  balance: number;
  balanceType: string;
  day: string;
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
