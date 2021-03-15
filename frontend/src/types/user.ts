export type User = {
  id: string;
  payment_method: string;
  last4: string;
  balance: number
  email: string;
  name: string;
  image: string;
  username: string;
  payout_eth: string;
  freq: number;
  seats: number;
  token: string;
  role: string;
};

export type ClientSecret = {
  client_secret: string;
}
