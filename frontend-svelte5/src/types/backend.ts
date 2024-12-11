
export type Token = {
    access_token: string;
    refresh_token: string;
    expires_at: number;
};

// @ts-ignore
import { components } from "./generated-backend-types";

export type User = {
    /** Format: uuid */
    id: string;
    email: string;
    name?: string | null;
    /** Format: date-time */
    createdAt: string;
    /** Format: uuid */
    invitedId?: string;
    stripeId?: string | null;
    image?: string | null;
    paymentMethod?: string | null;
    last4?: string | null;
    /** Format: int64 */
    seats?: number | null;
    /** Format: int64 */
    freq?: number | null;
    role?: string | null;
    multiplier: boolean;
    multiplierDailyLimit: number;
};

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

export type Repo = {

        /** Format: uuid */
        uuid: string;
        url?: string | null;
        gitUrl?: string | null;
        name?: string | null;
        description?: string | null;
        /** Format: uint32 */
        score: number;
        source?: string | null;
        /** Format: date-time */
        createdAt: string;
        trustAt: string;
        /** Format: float */
        healthValue: number;
        analyzed: boolean;

};

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
    repo: Repo;
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

export type RepoHealthValue = ({
    repoid?: string;
    healthvalue?: number;
});

export type Threshold= ({
    /** Format: int64 */
    upper: number;
    /** Format: int64 */
    lower: number;
});

export type HealthValueThreshold = ({
    id?: string;
    /** Format: date-time */
    createdAt?: string;
    ThContributorCount?: Threshold;
    ThCommitCount?: Threshold;
    ThSponsorDonation?: Threshold;
    ThRepoStarCount?: Threshold;
    ThRepoMultiplier?: Threshold;
    ThActiveFFSUserCount?: Threshold;
});

export type RepoMetrics= ({
    /** Format: uuid */
    id?: string;
    /** Format: uuid */
    repoid?: string;
    /** Format: date-time */
    createdat?: string;
    /** Format: int64 */
    contributorcount?: number;
    /** Format: int64 */
    commitcount?: number;
    /** Format: int64 */
    sponsorcount?: number;
    /** Format: int64 */
    repostarcount?: number;
    /** Format: int64 */
    repomultipliercount?: number;
    /** Format: int64 */
    activeffsusercount?: number;
});
export type PartialHealthValues= ({
    /** Format: uuid */
    repoid?: string;
    /** Format: float */
    contributorvalue?: number;
    /** Format: float */
    commitvalue?: number;
    /** Format: float */
    sponsorvalue?: number;
    /** Format: float */
    repostarvalue?: number;
    /** Format: float */
    repomultipliervalue?: number;
    /** Format: float */
    activeffsuservalue?: number;
});