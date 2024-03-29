/**
 * This file was auto-generated by openapi-typescript.
 * Do not make direct changes to the file.
 */


export interface paths {
  "/users/me": {
    get: {
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["User"];
          };
        };
      };
    };
  };
  "/users/me/git-email": {
    get: {
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": (components["schemas"]["GitUser"])[];
          };
        };
      };
    };
    post: {
      requestBody: {
        content: {
          "application/json": {
            /** Format: email */
            email?: string;
          };
        };
      };
      responses: {
        /** @description OK */
        200: never;
      };
    };
    delete: {
      parameters: {
        query: {
          email: string;
        };
      };
      responses: {
        /** @description OK */
        200: never;
      };
    };
  };
  "/users/me/method/": {
    put: {
      parameters: {
        query: {
          method: string;
        };
      };
      responses: {
        /** @description OK */
        200: never;
      };
    };
    delete: {
      responses: {
        /** @description OK */
        200: never;
      };
    };
  };
  "/users/me/sponsored": {
    get: {
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": (components["schemas"]["Repo"])[];
          };
        };
      };
    };
  };
  "/users/me/name/{name}": {
    put: {
      parameters: {
        path: {
          name: {
            image?: string;
          };
        };
      };
      responses: {
        /** @description OK */
        200: never;
      };
    };
  };
  "/users/me/clear/name": {
    put: {
      responses: {
        /** @description OK */
        200: never;
      };
    };
  };
  "/users/me/image": {
    post: {
      requestBody: {
        content: {
          "application/json": {
            image?: string;
          };
        };
      };
      responses: {
        /** @description OK */
        200: never;
      };
    };
    delete: {
      responses: {
        /** @description OK */
        200: never;
      };
    };
  };
  "/users/me/stripe": {
    put: {
      parameters: {
        query: {
          freq: number;
          seats: number;
        };
      };
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["ClientSecret"];
          };
        };
      };
    };
    post: {
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["ClientSecret"];
          };
        };
        /** @description Bad Request */
        400: never;
      };
    };
    delete: {
      responses: {
        /** @description OK */
        200: never;
      };
    };
  };
  "/users/me/nowPayment": {
    post: {
      parameters: {
        query: {
          freq: number;
          seats: number;
        };
      };
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["PaymentResponse"];
          };
        };
        /** @description Internal Server Error */
        500: never;
      };
    };
  };
  "/users/me/payment": {
    get: {
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["PaymentEvent"];
          };
        };
      };
    };
  };
  "/users/me/sponsored-users": {
    post: {
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["UserStatus"];
          };
        };
      };
    };
  };
  "/users/me/balance": {
    get: {
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["UserStatus"];
          };
        };
      };
    };
  };
  "/users/contrib-snd": {
    post: {
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": (components["schemas"]["Contribution"])[];
          };
        };
      };
    };
  };
  "/users/contrib-rcv": {
    post: {
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": (components["schemas"]["Contribution"])[];
          };
        };
      };
    };
  };
  "/users/me/contributions-summary": {
    post: {
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/js  on": (components["schemas"]["RepoBalance"])[];
          };
        };
      };
    };
  };
  "/users/contributions-summary": {
    post: {
      parameters: {
        query: {
          uuid: string;
        };
      };
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": (components["schemas"]["RepoBalance"])[];
          };
        };
      };
    };
  };
  "/users/summary": {
    post: {
      parameters: {
        query: {
          uuid: string;
        };
      };
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["User"];
          };
        };
      };
    };
  };
  "/users/git-email": {
    post: {
      requestBody: {
        content: {
          "application/json": components["schemas"]["EmailToken"];
        };
      };
      responses: {
        /** @description OK */
        200: never;
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/users/me/request-payout": {
    post: {
      parameters: {
        query: {
          targetCurrency: "ETH" | "GAS" | "USD";
        };
      };
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["PayoutResponse"];
          };
        };
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/users/{id}": {
    get: {
      parameters: {
        path: {
          id: string;
        };
      };
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": (components["schemas"]["PublicUser"])[];
          };
        };
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/users/by/{email}": {
    get: {
      parameters: {
        path: {
          id: string;
        };
      };
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": {
              user?: components["schemas"]["User"];
            };
          };
        };
        /** @description No Content */
        204: never;
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/repos/search": {
    get: {
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": (components["schemas"]["Repo"])[];
          };
        };
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/repos/name": {
    get: {
      parameters: {
        query: {
          q: string;
        };
      };
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": (components["schemas"]["Repo"])[];
          };
        };
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/repos/{id}": {
    get: {
      parameters: {
        path: {
          id: string;
        };
      };
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["Repo"];
          };
        };
        /** @description Not Found */
        404: never;
      };
    };
  };
  "/repos/{id}/tag": {
    post: {
      parameters: {
        path: {
          id: string;
        };
      };
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["Repo"];
          };
        };
        /** @description Bad Request */
        400: never;
        /** @description Internal Server Error */
        500: never;
      };
    };
  };
  "/repos/{id}/untag": {
    post: {
      parameters: {
        path: {
          id: string;
        };
      };
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["Repo"];
          };
        };
        /** @description Bad Request */
        400: never;
        /** @description Internal Server Error */
        500: never;
      };
    };
  };
  "/repos/{id}/{offset}/graph": {
    get: {
      parameters: {
        path: {
          id: string;
          offset: number;
        };
      };
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["Data"];
          };
        };
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/hooks/stripe": {
    post: {
      responses: {
        /** @description OK */
        200: never;
        /** @description Service Unavailable */
        503: never;
      };
    };
  };
  "/hooks/nowpayments": {
    post: {
      responses: {
        /** @description OK */
        200: never;
        /** @description Internal Server Error */
        500: never;
      };
    };
  };
  "/hooks/analyzer": {
    post: {
      requestBody: {
        content: {
          "application/json": components["schemas"]["WebhookCallback"];
        };
      };
      responses: {
        /** @description OK */
        200: never;
        /** @description Internal Server Error */
        500: never;
      };
    };
  };
  "/admin/payout/{exchangeRate}": {
    post: {
      parameters: {
        path: {
          exchangeRate: number;
        };
      };
      responses: {
        /** @description OK */
        200: never;
        /** @description Bad Request */
        400: never;
        /** @description Internal Server Error */
        500: never;
      };
    };
  };
  "/admin/time": {
    get: {
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["Time"];
          };
        };
      };
    };
  };
  "/admin/users": {
    post: {
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": (string)[];
          };
        };
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/config": {
    get: {
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["Config"];
          };
        };
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/confirm/invite/{email}": {
    post: {
      parameters: {
        path: {
          email: string;
        };
      };
      responses: {
        /** @description OK */
        200: never;
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/invite": {
    get: {
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": (components["schemas"]["Invitation"])[];
          };
        };
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/invite/by/{email}": {
    delete: {
      parameters: {
        path: {
          email: string;
        };
      };
      responses: {
        /** @description OK */
        200: never;
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/invite/my/{email}": {
    delete: {
      parameters: {
        path: {
          email: string;
        };
      };
      responses: {
        /** @description OK */
        200: never;
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/invite/{email}": {
    post: {
      parameters: {
        path: {
          email: string;
        };
      };
      responses: {
        /** @description OK */
        200: never;
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/admin/fake/user/{email}": {
    post: {
      parameters: {
        path: {
          email: string;
        };
      };
      responses: {
        /** @description OK */
        200: never;
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/admin/fake/payment/{email}/{seats}": {
    post: {
      parameters: {
        path: {
          email: string;
          seats: number;
        };
      };
      responses: {
        /** @description OK */
        200: never;
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/admin/fake/contribution": {
    post: {
      requestBody: {
        content: {
          "application/json": components["schemas"]["FakeRepoMapping"];
        };
      };
      responses: {
        /** @description OK */
        200: never;
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/admin/timewarp/{hours}": {
    post: {
      parameters: {
        path: {
          hours: number;
        };
      };
      responses: {
        /** @description OK */
        200: never;
        /** @description Bad Request */
        400: never;
      };
    };
  };
  "/nowpayments/crontester": {
    post: {
      requestBody: {
        content: {
          "application/json": {
            data?: {
              [key: string]: string | undefined;
            };
          };
        };
      };
      responses: {
        /** @description OK */
        200: never;
        /** @description Bad Request */
        400: never;
      };
    };
  };
}

export type webhooks = Record<string, never>;

export interface components {
  schemas: {
    User: {
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
    };
    Claims: {
      iss?: string | null;
      sub?: string | null;
      aud?: string | null;
      /** Format: int64 */
      exp?: number | null;
      /** Format: int64 */
      nbf?: number | null;
      /** Format: int64 */
      iat?: number | null;
      jti?: string | null;
    };
    GitUser: {
      email: string;
      /** Format: date-time */
      confirmedAt?: string | null;
      /** Format: date-time */
      createdAt?: string | null;
    };
    Repo: {
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
    };
    PaymentEvent: ({
        /** Format: uuid */
        id: string;
        /** Format: uuid */
        externalId?: string;
        /** Format: uuid */
        userId?: string;
        /** Format: bigint */
        balance?: number;
        currency?: string;
        status?: string;
        /** Format: int64 */
        seats: number;
        /** Format: int64 */
        freq: number;
        /** Format: date-time */
        createdAt?: string;
      })[];
    UserStatus: {
      /** Format: uuid */
      userId?: string;
      /** Format: email */
      email?: string;
      name?: string | null;
      daysLeft?: number;
    };
    UserBalance: ({
        currency?: string;
        balance?: number;
      })[];
    Contribution: {
      repoName: string;
      repoUrl: string;
      sponsorName?: string | null;
      sponsorEmail: string;
      contributorName?: string | null;
      contributorEmail: string;
      /** Format: int64 */
      balance: string;
      currency: string;
      /** Format: uuid */
      paymentCycleInId: string;
      /** Format: date-time */
      day: string;
      /** Format: date-time */
      claimedAt?: string;
    };
    RepoBalance: {
      repo: components["schemas"]["Repo"];
      currencyBalance: {
        [key: string]: string | undefined;
      };
    };
    EmailToken: {
      email: string;
      token: string;
    };
    Data: {
      /** Format: int32 */
      days?: number;
      /** Format: int32 */
      total?: number;
      datasets?: (components["schemas"]["Dataset"])[];
      labels?: (string)[];
    };
    Dataset: {
      label?: string;
      data?: (number)[];
      fill?: boolean;
      backgroundColor?: string;
      borderColor?: string;
      /** Format: int32 */
      pointBorderWidth?: number;
    };
    WebhookCallback: {
      requestId?: string;
      error?: string | null;
      result?: (components["schemas"]["FlatFeeWeight"])[];
    };
    FlatFeeWeight: {
      names?: (string)[];
      email?: string;
      weight?: number;
    };
    Time: {
      /** Format: date-time */
      time?: string;
      offset?: string;
    };
    Config: {
      stripePublicApi?: string;
      plans?: (components["schemas"]["Plan"])[];
      env?: string;
      supportedCurrencies?: {
        [key: string]: components["schemas"]["Currency"] | undefined;
      };
    };
    /**
     * @example {
     *   "ETH": {
     *     "name": "Ethereum",
     *     "short": "ETH",
     *     "smallest": "wei",
     *     "factorPow": 18,
     *     "isCrypto": true
     *   }
     * }
     */
    Currency: {
      name: string;
      short: string;
      smallest: string;
      /** Format: int64 */
      factorPow: number;
      isCrypto: boolean;
    };
    Plan: {
      title: string;
      /** Format: float */
      price: number;
      /** Format: int64 */
      freq: number;
      desc?: string;
      disclaimer?: string;
      /** Format: int64 */
      feePrm?: number;
    };
    Invitation: {
      email: string;
      inviteEmail: string;
      /** Format: date-time */
      confirmedAt?: string | null;
      /** Format: date-time */
      createdAt: string;
    };
    ClientSecret: {
      clientSecret: string;
    };
    PaymentResponse: {
      payAddress?: string;
      /** Format: int64 */
      payAmount?: string;
      payCurrency?: string;
    };
    FakeRepoMapping: {
      startDate: string;
      endDate: string;
      name: string;
      url: string;
      weights?: (components["schemas"]["FlatFeeWeight"])[];
    };
    PayoutResponse: {
      /** Format: int64 */
      amount: string;
      currency?: string;
      /** Format: byte */
      encodedUserId: string;
      /** Format: byte */
      signature: string;
    };
    PublicUser: {
      /** Format: uuid */
      id?: string;
      name?: string | null;
      image?: string | null;
    };
  };
  responses: never;
  parameters: never;
  requestBodies: never;
  headers: never;
  pathItems: never;
}

export type external = Record<string, never>;

export type operations = Record<string, never>;
