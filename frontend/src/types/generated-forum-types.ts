/**
 * This file was auto-generated by openapi-typescript.
 * Do not make direct changes to the file.
 */


export interface paths {
  "/metrics": {
    /** Get metrics */
    get: {
      responses: {
        /** @description ok */
        200: never;
      };
    };
  };
  "/posts": {
    /** Get all posts */
    get: {
      parameters: {
        query: {
          /** @description Only retrieve open or closed discussions */
          open?: boolean;
        };
      };
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": (components["schemas"]["Post"])[];
          };
        };
        204: components["responses"]["NoContent"];
        500: components["responses"]["InternalServerError"];
      };
    };
    /** Create a new post */
    post: {
      requestBody: {
        content: {
          "application/json": components["schemas"]["PostInput"];
        };
      };
      responses: {
        /** @description Created */
        201: {
          content: {
            "application/json": components["schemas"]["Post"];
          };
        };
        400: components["responses"]["BadRequest"];
        401: components["responses"]["Unauthorized"];
        500: components["responses"]["InternalServerError"];
      };
    };
  };
  "/posts/{postId}": {
    /** Get a specific post */
    get: {
      parameters: {
        path: {
          postId: components["parameters"]["PostId"];
        };
      };
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["Post"];
          };
        };
        404: components["responses"]["NotFound"];
        500: components["responses"]["InternalServerError"];
      };
    };
    /** Update a post */
    put: {
      parameters: {
        path: {
          postId: components["parameters"]["PostId"];
        };
      };
      requestBody: {
        content: {
          "application/json": components["schemas"]["PostInput"];
        };
      };
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["Post"];
          };
        };
        400: components["responses"]["BadRequest"];
        401: components["responses"]["Unauthorized"];
        403: components["responses"]["Forbidden"];
        404: components["responses"]["NotFound"];
        500: components["responses"]["InternalServerError"];
      };
    };
    /** Delete a Post */
    delete: {
      parameters: {
        path: {
          postId: components["parameters"]["PostId"];
        };
      };
      responses: {
        /** @description OK */
        200: never;
        204: components["responses"]["NoContent"];
        401: components["responses"]["Unauthorized"];
      };
    };
  };
  "/posts/{postId}/close": {
    /** Close a post for further edits and comments */
    put: {
      parameters: {
        path: {
          postId: components["parameters"]["PostId"];
        };
      };
      responses: {
        /** @description OK */
        200: never;
        401: components["responses"]["Unauthorized"];
        403: components["responses"]["Forbidden"];
        404: components["responses"]["NotFound"];
        500: components["responses"]["InternalServerError"];
      };
    };
  };
  "/posts/{postId}/comments": {
    /** Get all comments */
    get: {
      parameters: {
        path: {
          postId: components["parameters"]["PostId"];
        };
      };
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": (components["schemas"]["Comment"])[];
          };
        };
        204: components["responses"]["NoContent"];
        404: components["responses"]["NotFound"];
        500: components["responses"]["InternalServerError"];
      };
    };
    /** Add a comment to a post */
    post: {
      parameters: {
        path: {
          postId: components["parameters"]["PostId"];
        };
      };
      requestBody: {
        content: {
          "application/json": components["schemas"]["CommentInput"];
        };
      };
      responses: {
        /** @description Created */
        201: {
          content: {
            "application/json": components["schemas"]["Comment"];
          };
        };
        400: components["responses"]["BadRequest"];
        401: components["responses"]["Unauthorized"];
        404: components["responses"]["NotFound"];
        500: components["responses"]["InternalServerError"];
      };
    };
  };
  "/posts/{postId}/comments/{commentId}": {
    /** Update a comment */
    put: {
      parameters: {
        path: {
          postId: components["parameters"]["PostId"];
          commentId: components["parameters"]["CommentId"];
        };
      };
      requestBody: {
        content: {
          "application/json": components["schemas"]["CommentInput"];
        };
      };
      responses: {
        /** @description OK */
        200: {
          content: {
            "application/json": components["schemas"]["Comment"];
          };
        };
        400: components["responses"]["BadRequest"];
        401: components["responses"]["Unauthorized"];
        403: components["responses"]["Forbidden"];
        404: components["responses"]["NotFound"];
        500: components["responses"]["InternalServerError"];
      };
    };
    /** Delete a comment */
    delete: {
      parameters: {
        path: {
          postId: components["parameters"]["PostId"];
          commentId: components["parameters"]["CommentId"];
        };
      };
      responses: {
        /** @description OK */
        200: never;
        401: components["responses"]["Unauthorized"];
        404: components["responses"]["NotFound"];
      };
    };
  };
}

export type webhooks = Record<string, never>;

export interface components {
  schemas: {
    Post: {
      /** Format: uuid */
      id: string;
      title: string;
      content: string;
      /** Format: uuid */
      author: string;
      /** Format: date-time */
      created_at: string;
      /** Format: date-time */
      updated_at?: string;
      open: boolean;
      proposal_id?: number;
    };
    Comment: {
      /** Format: uuid */
      id: string;
      content: string;
      /** Format: uuid */
      author: string;
      /** Format: date-time */
      created_at: string;
      /** Format: date-time */
      updated_at?: string;
    };
    PostInput: {
      title: string;
      content: string;
    };
    CommentInput: {
      content: string;
    };
  };
  responses: {
    /** @description Bad Request */
    BadRequest: {
      content: {
        "application/json": {
          error: string;
        };
      };
    };
    /** @description Unauthorized */
    Unauthorized: {
      content: {
        "application/json": {
          error: string;
        };
      };
    };
    /** @description Forbidden */
    Forbidden: {
      content: {
        "application/json": {
          error: string;
        };
      };
    };
    /** @description Not Found */
    NotFound: {
      content: {
        "application/json": {
          error: string;
        };
      };
    };
    /** @description No Content */
    NoContent: {
      content: {
        "application/json": {
          info?: string;
        };
      };
    };
    /** @description Internal Server Error */
    InternalServerError: never;
  };
  parameters: {
    /** @description ID of the post */
    PostId: string;
    /** @description ID of the comment */
    CommentId: string;
  };
  requestBodies: never;
  headers: never;
  pathItems: never;
}

export type external = Record<string, never>;

export type operations = Record<string, never>;
