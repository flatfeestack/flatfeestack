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
