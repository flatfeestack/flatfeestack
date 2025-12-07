import {
  type Address,
  type Hex,
} from 'viem';

export interface ProposalView {
    id: bigint;
    description: string;
    startTime: bigint;
    endTime: bigint;
    proposalState: number;
    againstVotes: bigint;
    forVotes: bigint;
    abstainVotes: bigint;
    proposer: `0x${string}`;
}

export interface ProposalViewExtended extends ProposalView {
    startDate?: Date;
    endDate?: Date;
}

export type MembershipTokenInfo = {
  tokenId: bigint;
  membershipPaidUntil: bigint;
};

export interface ProposalDetails {
  targets: Address[];
  values: bigint[];
  calldatas: Hex[];
  description: string;
}