# analysis-engine
This repository contains the engine, that evaluates the contribution for each contributer of an open source git repository. 

## Installation

1. Clone the repository into your GOPATH (e.g. src/github.com/flatfeestack/analysis-engine)
2. Get the dependencies with `go get`

## Usage

Run the server:
```
go run github.com/flatfeestack/analysis-engine
```

Request a listing of all contributors with their contributions:
1. Make a GET request to http://localhost:8080
2. Pass the repository url as a route parameter
    ```
    repositoryUrl=https://github.com/styled-components/styled-components.git
    ```
3. Pass the timestamp from when you want the analysis to start as a route parameter (optional)
   ```
   since=1577836800
   ```
3. Pass the timestamp from when you want the analysis to end as a route parameter (optional)
   ```
   until=1588776384
   ```
   
## Example

```
GET http://localhost:8080?repositoryUrl=https://github.com/styled-components/styled-components.git&since=1577836800&until=1588776384
```

This will return a JSON with all contributiors and their contributions:
```
[
    {
        "contributor": {
            "name": "macintoshhelper",
            "email": "macintoshhelper@users.noreply.github.com"
        },
        "changes": {
            "addition": 128,
            "deletion": 9
        }
    },
    {
        "contributor": {
            "name": "Kristóf Poduszló",
            "email": "kripod@protonmail.com"
        },
        "changes": {
            "addition": 9,
            "deletion": 2
        }
    },
    {
        "contributor": {
            "name": "dependabot[bot]",
            "email": "49699333+dependabot[bot]@users.noreply.github.com"
        },
        "changes": {
            "addition": 2020,
            "deletion": 4667
        }
    },
    {
        "contributor": {
            "name": "Phil Plückthun",
            "email": "phil@kitten.sh"
        },
        "changes": {
            "addition": 167,
            "deletion": 4
        }
    },
    {
        "contributor": {
            "name": "Matt Lubner",
            "email": "matt@mattlubner.com"
        },
        "changes": {
            "addition": 3031,
            "deletion": 3302
        }
    },
    {
        "contributor": {
            "name": "Evan Jacobs",
            "email": "probablyup@gmail.com"
        },
        "changes": {
            "addition": 16843,
            "deletion": 15117
        }
    },
    {
        "contributor": {
            "name": "Keegan Street",
            "email": "keeganstreet@gmail.com"
        },
        "changes": {
            "addition": 43,
            "deletion": 1
        }
    },
    {
        "contributor": {
            "name": "egdbear",
            "email": "1176374+egdbear@users.noreply.github.com"
        },
        "changes": {
            "addition": 1,
            "deletion": 1
        }
    },
    {
        "contributor": {
            "name": "Phil Pluckthun",
            "email": "phil@kitten.sh"
        },
        "changes": {
            "addition": 1,
            "deletion": 1
        }
    },
    {
        "contributor": {
            "name": "Samuli Ulmanen",
            "email": "s.ulmanen@gmail.com"
        },
        "changes": {
            "addition": 9,
            "deletion": 9
        }
    },
    {
        "contributor": {
            "name": "Jacob Duval",
            "email": "jladuval@gmail.com"
        },
        "changes": {
            "addition": 4,
            "deletion": 1
        }
    }
]
```