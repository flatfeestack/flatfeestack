// @ts-ignore
import { components } from "./generated-backend-types";

export type User = components["schemas"]["User"];

export type Config = {
    stripePublicApi: string;
    plans: (components["schemas"]["Plan"])[];
    env: string;
    supportedCurrencies: {
        [key: string]: components["schemas"]["Currency"] | undefined;
    };
};

export type Plan = components["schemas"]["Plan"];

export type ClientSecret = components["schemas"]["ClientSecret"];

export type UserStatus = components["schemas"]["UserStatus"];

export type PaymentEvent = components["schemas"]["PaymentEvent"];

export type Repo = components["schemas"]["Repo"];

export type GitUser = components["schemas"]["GitUser"];

export type Invitation = components["schemas"]["Invitation"];

export type Time = components["schemas"]["Time"];

export type RepoMapping = components["schemas"]["FakeRepoMapping"];

export type Contribution = {
    repoName: string;
    repoUrl: string;
    sponsorName?: string | null;
    sponsorEmail: string;
    contributorName?: string | null;
    contributorEmail: string;
    /** Format: int64 */
    balance: bigint;
    currency: string;
    /** Format: uuid */
    paymentCycleInId: string;
    /** Format: date-time */
    day: string;
    /** Format: date-time */
    claimedAt?: string;
};

export type ContributionSummary = {
    repo: components["schemas"]["Repo"];
    currencyBalance: {
        [key: string]: bigint;
    };
};

export type Currency = components["schemas"]["Currency"];

export type PaymentResponse = components["schemas"]["PaymentResponse"];

export type ChartDataTotal = {
    /** Format: int32 */
    days: number;
    /** Format: int32 */
    total: number;
    datasets: (components["schemas"]["Dataset"])[];
    labels: (string)[];
};

export type PayoutResponse = components["schemas"]["PayoutResponse"];

export type PublicUser = components["schemas"]["PublicUser"];

export type  UserBalance = ({
    currency: string;
    balance: number;
})[];
