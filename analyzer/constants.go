package main

// Represents the factor of the total changed lines
// with witch the merger gets rewarded for merging the branch.
// Changed lines in normal commits are considered with factor 1
// while the changed lines in merges (summary of the size of the merge)
// are considered with this factor.
const mergedLinesWeight = 0.1

//////////////////////
// Weights
//////////////////////

// ---- Weights inside categories ----

// Category "Changes" devided into additions and deletions.
// Both must sum up to 1
const additionWeight = 0.7
const deletionWeight = 0.3

// Category "GitHistory" devided into commits and merges.
// Both must sum up to 1
const commitWeight = 0.7
const mergeWeight = 0.3

// Category "Issues" divided into issues (created by the user),
// issue comments (on issues created by the user), and commenter (comments of the user).
// All must sum up to 1.
const issueWeight = 0.5
const issueCommentsWeight = 0.2
const issueCommenterWeight = 0.3

// Category "PullRequests"
const pullRequestAuthorWeight = 0.7
const pullRequestReviewerWeight = 0.3

// ---- Weights between categories ----

// Intercategory weights between categories Changes and Githistory.
// All must sum up to 1.
// Only when platformInformation IS NOT considered
const changesWeight = 0.55
const gitHistoryWeight = 0.45

// Intercategory weights between categories Changes, Githistory, Issues and Pull Requests.
// All must sum up to 1.
// Only when platformInformation IS considered
const changesWeightPlatformInfo = 0.36
const gitHistoryWeightPlatformInfo = 0.3
const issueCategoryWeightPlatformInfo = 0.14
const pullRequestCategoryWeightPlatformInfo = 0.2

//////////////////////
// Pull Request Value
//////////////////////

// Value of the pull request when the state is one of the following
const pullRequestClosedValue = 0.6
const pullRequestMergedValue = 1.5
const pullRequestOpenValue = 1.0

// Multiplier when there is an approval inside the pull request
const approvedMultiplier = 1.4
