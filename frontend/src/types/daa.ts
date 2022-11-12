import type { SvelteComponentTyped } from "svelte";

export interface ProposalType {
  component: typeof SvelteComponentTyped<ProposalFormProps>;
  text: string;
}

export interface ProposalFormProps {
  targets: string[];
  values: number[];
  transferCallData: string[];
}
