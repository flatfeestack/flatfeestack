# FlatFeeStack Frontend

![Pipeline](https://github.com/flatfeestack/frontend/actions/workflows/ci.yml/badge.svg)

## Development

Get the PNPM package manager and install all packages:

```shell
npm i -g pnpm
pnpm i
```

The development server can be run with:

```shell
pnpm run dev
```

Code formatting is automated with Prettier:

```shell
pnpm run prettify
```

## Integration with FlatFeeStack DAA

The frontend offers an interface to interact with the FlatFeeStack DAA. Right now this only works by starting a local Ethereum blockchain. Deployment to this local chain and exports of the contract's ABIs is automated from the [repository](https://github.com/flatfeestack/daa).

If you do a fresh deployment, restart the frontend development server as the contracts deployment addresses are injected via environment variables.
