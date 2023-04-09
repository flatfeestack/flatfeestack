# FlatFeeStack Frontend

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

If the backend schema changes, the frontend needs to be regenerated:

```shell
pnpm run schema:backend
```

## Integration with FlatFeeStack DAO

The frontend offers an interface to interact with the FlatFeeStack DAO. Right now this only works by starting a local Ethereum blockchain. Deployment to this local chain and exports of the contract's ABIs is automated from the [repository](https://github.com/flatfeestack/daa).

If you do a fresh deployment, restart the frontend development server as the contracts deployment addresses are injected via environment variables.
