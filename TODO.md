## ToDo's

### Tech:

Now:
- [x] Automatic release deployment (Guil)
  - [x] DNS
  - [x] Tag-based auto deployment (GitHub Actions, only master branch, tag)
  - [x] Infura node (Alchemy)
  - [x] Volume for aanalyzer
- [ ] Test containers: the goal is to execute tests BEFORE build
- [ ] Check if we can use podman?
- [ ] Create GitHub users
- [x] Create SendGrid account
- [x] Integrate / configure sendgrid to backend
- [x] Evaluate whether the "scheduler" should be deleted for good or not; If yes, delete.
- [ ] Sendgrid: change the if-statement in the backend + fastauth
- [ ] Deployment: volume for logs in all services (that we don't loose the logs on every deployment!)
- [ ] Automatically add the user registration email (flatfeestrack account) to the git emails -- since that email is already confirmed! 

Future plans:
- [ ] Automatic conversion from fiat (stripe) to ETH
- [ ] Additional blockchains? (NEO, ...)
- [ ] Monitoring dashboard (Guil)
- [ ] Front-end Design
- [ ] Migration DB engine for backend + fastauth
- [ ] Sanity check and alerting if payments are not correct (sum, etc)
- [ ] What if a user donates to a SINGLE repo, but no developer registers 
in flatfeestack after 3 months? Should the amount be returned to the user, or be 
split to other projects? or what's the flow? 

### Marketing:

- [x] Landing Page (Thomas)
- [ ] 99Designs for logo and brand identity
- [ ] Create badges for sponsors (repos, etc)
- [ ] 3 minute pitch deck

Questionnaire / Interview:

- What is the tech stack used in your company?
- Do the engineers contribute to OSS (e.g., PRs)?
- What is the most exciting OSS project that you're currently aware of?
- Can you imagine developing your software solutions without any OSS? For example, no `Linux` or `npm`.
- If you rely on OSS, how do you ensure that the software/lib/project is properly maintained over time?
- If you realize that a library is not being maintained, what would you do?
  -- Fork and maintain (open source, or closed source?)
  -- Switch to an alternative solution
- What if you're missing a feature in an OSS project?
  -- Fork and maintain (open source, or closed source?)
  -- Switch to an alternative solution
  -- Engage to developers/maintainers
- How do you engage with developers/maintainers?
  -- PRs?
  -- Donations?
  -- Reporting issues and commenting
- Does donation would be an option?
  -- GitHub sponsors?
  -- OpenCollective
  -- Patreon
  -- others
- If no, what is the main barrier to NOT contribute/donate?
  -- Organization-level decision on what to contribute?
  -- Budget allocation?
  -- Accounting?
  -- others?
  