# analysis-engine for flatfeestack
This repository contains the engine, that evaluates the contribution 
for each contributer of an open source git repository. 
A simple ``GET`` request to the server with the parameters to configure 
the analyzation is all you need to analyze any publicly available repository. 

## Motivation
It is hard to get paid for working as an open source developer. 
Although there are some hurdles that have to be overcome, 
flatfeestack manages to find a solution to pay open source developers. 
Unlike conventional donation and sponsoring platforms, 
flatfeestack analyzes the repository independently and gives 
each contributor the appropriate amount of money according to their contribution.

## Tech
The code of this project is written in golang and is based on the 
libraries [gorilla/mux](https://github.com/gorilla/mux/) and [go-git/go-git](https://github.com/go-git/go-git/).

## Installation

1. Clone the repository into your GOPATH (e.g. src/github.com/flatfeestack/analysis-engine)
2. Get the dependencies with `go get`

## Usage

Run the server:
```
go run github.com/flatfeestack/analysis-engine
```
### API

The project is configured that way, that it runs the server at port 8080 of localhost. 
To call an endpoint simply make a `GET` request to the server at `localhost:8080` with 
one of the following endpoints.

#### /contributions

Request a listing of all contributors with their contributions:

##### parameters

* repositoryUrl
    + required
    + link to .git file or the repository
    + example: `repositoryUrlhttps://github.com/neow3j/neow3j.git`
* since
    + optional (default: date of first commit)
    + date at which the analysis should start in RFC3339 format.
    + example: `since=2020-01-22T15:04:05Z`
* until
    + optional (default: now)
    + date at which the analysis should end in RFC3339 format.
    + example: `until=2020-07-12T11:34:55Z`
* platformInformation
    + optional (default: false)
    + flag whether platform information such as issues and pull requests should be analyzed.
    + example: `platformInformation=true`
* branch
    + optional (default: master)
    + name of the branch that should be analyzed.
    + example: `branch=develop`

##### response

```
[{
        "gitInformation": {
            "contributor": {
                "name": "Guil. Sperb Machado",
                "email": "guil@axlabs.com"
            },
            "changes": {
                "addition": 44580,
                "deletion": 19538
            },
            "merges": 30,
            "commits": 272
        },
        "platformInformation": {
            "userName": "gsmachado",
            "issueInformation": {
                "author": [
                    1,
                    2,
                    0,
                    3,
                    5,
                    0
                ],
                "commenter": 97
            },
            "pullRequestInformation": {
                "author": [
                    {
                        "state": "MERGED",
                        "reviews": [
                            "CHANGES_REQUESTED",
                            "COMMENTED",
                            "APPROVED"
                        ]
                    },
                    {
                        "state": "MERGED",
                        "reviews": null
                    }
                ],
                "reviewer": 91
            }
        }
    }]
```

This is the structure of the response (JSON). 
* the information under `contributor` are the ones available from git
* `changes` represent the lines changed devided in additions and deletions
* `merges` represent the number of merges while `commits` represent the number of commits
* `platformInformation` will only be returned if the flag in the request is set.
* `username` is the username on the platform (e.g. GitHub)
* the array under `issueInformation/author` represents the amount of comments on an issue that this user was the author of
* `commenter` represents the amount of comments written
* the array under `pullRequestInformation/author` represents the state and the activities of pull request this user was the author of
* `reviewer` states how many code reviews this user made

#### /weights

Request a listing of the share of the total contribution by each contributor:

##### parameters

* repositoryUrl
    + required
    + link to .git file or the repository
    + example: `repositoryUrlhttps://github.com/neow3j/neow3j.git`
* since
    + optional (default: date of first commit)
    + date at which the analysis should start in RFC3339 format.
    + example: `since=2020-01-22T15:04:05Z`
* until
    + optional (default: now)
    + date at which the analysis should end in RFC3339 format.
    + example: `until=2020-07-12T11:34:55Z`
* platformInformation
    + optional (default: false)
    + flag whether platform information such as issues and pull requests should be analyzed.
    + example: `platformInformation=true`
* branch
    + optional (default: master)
    + name of the branch that should be analyzed.
    + example: `branch=develop`

##### response

```
[    {
         "contributor": {
             "name": "Claude Muller",
             "email": "claude@axlabs.com"
         },
         "weight": 0.5956763166051634
     },]
```

This is the structure of the response (JSON). 
* the information under `contributor` are the ones available from git
* `weight` represents the share of the total contribution of this user. (e.g. 0.59 means the user made 59% of all contributions) 

