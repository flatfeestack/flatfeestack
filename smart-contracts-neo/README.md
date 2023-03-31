# payout-neo-contracts

This repository contains a PoC implementation of a smart contract for the Neo N3 blockchain. It introduces a mechanism
for transparent and scalable blockchain-based payments for the FlatFeeStack project.

### Dependencies

- Java 1.8
- [Docker](https://docs.docker.com/get-docker/)

### Compilation

To compile a SC, change to the variable `className` in the gradle task `neow3jCompiler` to the appropriate SC file. 

```gradle
neow3jCompiler {
    className = "io.flatfeestack.PayoutNeo"
    debug = true
}
```

Then run the following command to compile the PayoutNeo smart contract and find its compiled components as a `.nef` and
a `.manifest.json` file in the `/build` folder.

```bash
./gradlew neow3jCompile
```

### Testing

The smart contracts `PayoutNeo` and `PayoutNeoForEvaluation` have been tested thoroughly. The tests run with neow3j's
integrated test framework which requires docker.

### Evaluation

This project was constructed in the scope of a University thesis. Thus, it contains additional evaluation code among the test files. As 
well as the integration tests, the evaluation code utilizes [neo3-privatenet-docker](https://github.com/AxLabs/neo3-privatenet-docker) 
provided by [AxLabs](https://axlabs.com/) local Neo N3 network. Every fee calculation of a transaction is followed by
its execution to verify the expected state change to ensure each fee calculation matches the correct transaction.

The results of the evaluation can be found in the `/evaluation_results` folder, whereas the cleaned up final results can be found in the 
subfolders `/neo` and `/eth`. The `/eth` folder contains the results of the Ethereum https://github.
com/flatfeestack/payout-eth-contracts that has been implemented 
simultaneously ([payout-eth-contracts](https://github.com/flatfeestack/payout-eth-contracts)). It is included in this repository due to the 
jupyter notebook `visualise_evaluation.ipynb` that is used to create the 
needed plots for the thesis.

### Author
[mialbu](https://github.com/mialbu)
