export interface DaoConfig {
  chainId: number;
  dao: string;
  membership: string;
  wallet: string;
}

interface PayoutAddresses {
  eth: string;
  neo: string;
  usdc: string;
}

export interface PayoutConfig {
  payoutContractAddresses: PayoutAddresses;
  chainId: number;
  env: string;
}
