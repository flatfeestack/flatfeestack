## ToDo's

### Tech:

Now:
- [x] Automatic release deployment (Guil)
  - [x] DNS
  - [x] Tag-based auto deployment (GitHub Actions, only master branch, tag)
  - [x] Infura node (Alchemy)
  - [x] Volume for analysis-engine
- [ ] Test containers: the goal is to execute tests BEFORE build
- [ ] Check if we can use podman?
- [ ] Create GitHub users
- [x] Create SendGrid account
- [x] Integrate / configure sendgrid to backend
- [x] Evaluate whether the "scheduler" should be deleted for good or not; If yes, delete.
- [ ] Sendgrid: change the if-statement in the backend + fastauth
- [ ] Deployment: volume for logs in all services (that we don't loose the logs on every deployment!)

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