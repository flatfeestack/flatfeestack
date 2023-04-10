import type { SvelteComponentTyped } from "svelte";

export interface ProposalType {
  component: typeof SvelteComponentTyped<ProposalFormProps>;
  text: string;
}

export interface Call {
  target: string;
  value: number;
  transferCallData: string;
}

export interface ProposalFormProps {
  calls: Call[];
}
