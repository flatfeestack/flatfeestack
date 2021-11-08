package io.flatfeestack;

import io.neow3j.compiler.CompilationUnit;
import io.neow3j.compiler.Compiler;
import io.neow3j.contract.ContractManagement;
import io.neow3j.contract.GasToken;
import io.neow3j.contract.NefFile;
import io.neow3j.contract.PolicyContract;
import io.neow3j.contract.SmartContract;
import io.neow3j.contract.Token;
import io.neow3j.crypto.ECKeyPair;
import io.neow3j.crypto.Sign;
import io.neow3j.protocol.Neow3j;
import io.neow3j.protocol.core.response.ContractManifest;
import io.neow3j.protocol.core.response.NeoSendRawTransaction;
import io.neow3j.transaction.Transaction;
import io.neow3j.types.ContractParameter;
import io.neow3j.types.Hash160;
import io.neow3j.types.Hash256;
import io.neow3j.types.NeoVMStateType;
import io.neow3j.wallet.Account;

import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.math.BigDecimal;
import java.math.BigInteger;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

import static io.neow3j.contract.ContractUtils.writeContractManifestFile;
import static io.neow3j.contract.ContractUtils.writeNefFile;
import static io.neow3j.contract.SmartContract.calcContractHash;
import static io.neow3j.contract.Token.toFractions;
import static io.neow3j.crypto.Sign.signMessage;
import static io.neow3j.protocol.ObjectMapperFactory.getObjectMapper;
import static io.neow3j.transaction.AccountSigner.calledByEntry;
import static io.neow3j.transaction.AccountSigner.none;
import static io.neow3j.types.ContractParameter.hash160;
import static io.neow3j.types.ContractParameter.integer;
import static io.neow3j.types.ContractParameter.map;
import static io.neow3j.types.ContractParameter.publicKey;
import static io.neow3j.utils.ArrayUtils.concatenate;
import static io.neow3j.utils.ArrayUtils.reverseArray;
import static io.neow3j.utils.Await.waitUntilTransactionIsExecuted;
import static io.neow3j.wallet.Account.createMultiSigAccount;
import static java.util.Collections.singletonList;

public class EvaluationHelper {

    static final BigInteger TENTH_GAS = toFractions(new BigDecimal("0.1"), GasToken.DECIMALS);
    static final BigInteger ONE_GAS = toFractions(BigDecimal.ONE, GasToken.DECIMALS);
    static final BigInteger TEN_GAS = toFractions(BigDecimal.TEN, GasToken.DECIMALS);
    static final BigInteger HUNDRED_GAS = toFractions(BigDecimal.valueOf(100), GasToken.DECIMALS);
    static final BigInteger TSD_GAS = toFractions(BigDecimal.valueOf(1000), GasToken.DECIMALS);

    static final BigInteger FEE_PER_BYTE = BigInteger.valueOf(100);
    static final BigInteger STORAGE_PRICE = BigInteger.valueOf(10_000);
    static final BigInteger EXEC_FEE_FACTION = BigInteger.valueOf(3);

    private static final Path PAYOUT_EVALUATION_CONTRACT_NEF = Paths.get("./build/neow3j/PayoutNeoForEvaluation.nef");
    private static final Path PAYOUT_EVALUATION_CONTRACT_MANIFEST =
            Paths.get("./build/neow3j/PayoutNeoForEvaluation.manifest.json");

    static final Account defaultAccount = Account.fromWIF("L1eV34wPoj9weqhGijdDLtVQzUpWGHszXXpdU9dPuh2nRFFzFa7E");
    static final ECKeyPair.ECPublicKey defaultPubKey = defaultAccount.getECKeyPair().getPublicKey();
    static final Account committee = createMultiSigAccount(singletonList(defaultPubKey), 1);
    static final Account feePayAccount = Account.create();

    static final Account owner = Account.fromWIF("L3cNMQUSrvUrHx1MzacwHiUeCWzqK2MLt5fPvJj9mz6L2rzYZpok");
    static final ECKeyPair.ECPublicKey ownerPubKey = owner.getECKeyPair().getPublicKey();

    static Hash160[] getRandomHashes(int arrLength) {
        Hash160[] arr = new Hash160[arrLength];
        for (int i = 0; i < arrLength; i++) {
            arr[i] = Account.create().getScriptHash();
        }
        return arr;
    }

    static BigInteger[] getUniformTeas(int arrLength, BigInteger start, BigInteger step) {
        BigInteger[] arr = new BigInteger[arrLength];
        BigInteger tea = start;
        for (int i = 0; i < arrLength; i++) {
            arr[i] = tea;
            tea = tea.add(step);
        }
        return arr;
    }

    static BigInteger[] getRandomTeasToPreset(int nrAccounts, long min, long multiplier) {
        BigInteger[] arr = new BigInteger[nrAccounts];
        for (int i = 0; i < nrAccounts; i++) {
            BigInteger rand = BigInteger.valueOf((long) (Math.random() * multiplier) + min);
            arr[i] = rand;
        }
        return arr;
    }

    static BigInteger getSum(BigInteger[] arr) {
        BigInteger totalAmount = BigInteger.ZERO;
        for (BigInteger val : arr) {
            totalAmount = totalAmount.add(val);
        }
        return totalAmount;
    }

    static FileWriter getFileWriter(File file) throws IOException {
        return new FileWriter(file, false);
    }

    static File openFile(String filename) throws IOException {
        File file = new File("evaluation_results/" + filename);
        if (file.createNewFile()) {
            System.out.println("File created: " + file.getName());
        } else {
            System.out.println("Overwriting existing file.");
        }
        return file;
    }

    static void handleFeeFactors(Neow3j neow3j, Account committee, Account defaultAccount) throws Throwable {
        PolicyContract policyContract = new PolicyContract(neow3j);
        Transaction tx = policyContract.setFeePerByte(FEE_PER_BYTE).signers(calledByEntry(committee))
                .getUnsignedTransaction();
        tx.addMultiSigWitness(committee.getVerificationScript(), defaultAccount);
        Hash256 txHashFeePerByte = tx.send().getSendRawTransaction().getHash();

        tx = policyContract.setStoragePrice(STORAGE_PRICE).signers(calledByEntry(committee))
                .getUnsignedTransaction();
        tx.addMultiSigWitness(committee.getVerificationScript(), defaultAccount);
        Hash256 txHashStoragePrice = tx.send().getSendRawTransaction().getHash();

        tx = policyContract.setExecFeeFactor(EXEC_FEE_FACTION).signers(calledByEntry(committee))
                .getUnsignedTransaction();
        tx.addMultiSigWitness(committee.getVerificationScript(), defaultAccount);
        Hash256 txHashExecFeeFactor = tx.send().getSendRawTransaction().getHash();

        waitUntilTransactionIsExecuted(txHashFeePerByte, neow3j);
        waitUntilTransactionIsExecuted(txHashStoragePrice, neow3j);
        waitUntilTransactionIsExecuted(txHashExecFeeFactor, neow3j);

        BigDecimal actualFeePerByte = Token.toDecimals(policyContract.getFeePerByte(), GasToken.DECIMALS);
        BigInteger actualStoragePrice = policyContract.getStoragePrice();
        BigInteger actualExecFeeFactor = policyContract.getExecFeeFactor();

        System.out.println("\n##############feefactors#################");
        System.out.printf("Network Fee Per Byte: '%s'\n", actualFeePerByte);
        System.out.printf("Storage Fee Factor:   '%s'\n", actualStoragePrice);
        System.out.printf("Execution Fee Factor: '%s'\n", actualExecFeeFactor);
        System.out.println("##############feefactors#################\n");
    }

    static void compileContract(String canonicalName) throws IOException {
        CompilationUnit res = new Compiler().compile(canonicalName);
        // Write contract (compiled, NEF) to the disk
        Path buildNeow3jPath = Paths.get("build", "neow3j");
        buildNeow3jPath.toFile().mkdirs();
        writeNefFile(res.getNefFile(), res.getManifest().getName(), buildNeow3jPath);
        // Write manifest to the disk
        writeContractManifestFile(res.getManifest(), buildNeow3jPath);
    }

    static SmartContract deployPayoutNeoContract(Neow3j neow3j) throws Throwable {
        File nefFile = new File(PAYOUT_EVALUATION_CONTRACT_NEF.toUri());
        NefFile nef = NefFile.readFromFile(nefFile);

        File manifestFile = new File(PAYOUT_EVALUATION_CONTRACT_MANIFEST.toUri());
        ContractManifest manifest = getObjectMapper().readValue(manifestFile, ContractManifest.class);
        Hash256 txHash = new ContractManagement(neow3j)
                .deploy(nef, manifest, publicKey(ownerPubKey.getEncoded(true)))
                .signers(none(owner))
                .sign()
                .send()
                .getSendRawTransaction()
                .getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
        System.out.println("Deployed PayoutNeo contract");
        Hash160 hash = calcContractHash(owner.getScriptHash(), nef.getCheckSumAsInteger(),
                manifest.getName());
        return new SmartContract(hash, neow3j);
    }

    static void fundContract(Neow3j neow3j, SmartContract contract, BigDecimal fundContractAmount) throws Throwable {
        BigInteger fundAmountFractions = Token.toFractions(fundContractAmount, GasToken.DECIMALS);
        System.out.println(fundAmountFractions);
        Transaction tx = new GasToken(neow3j)
                .transfer(committee, contract.getScriptHash(), fundAmountFractions)
                .getUnsignedTransaction();
        tx.addMultiSigWitness(committee.getVerificationScript(), defaultAccount);
        Hash256 txHash = sendAndWaitUntilTransactionIsExecuted(tx, neow3j);
        NeoVMStateType state = neow3j.getApplicationLog(txHash).send()
                .getApplicationLog().getExecutions().get(0).getState();
        if (!state.equals(NeoVMStateType.HALT)) {
            throw new RuntimeException("Contract could not be funded.");
        }
        System.out.printf("Contract funded with %s GAS\n", fundContractAmount);
    }

    static void fundAccounts(Neow3j neow3j, BigInteger gasFractions, Account... accounts) throws Throwable {
        GasToken gasToken = new GasToken(neow3j);
        BigInteger minAmount = gasToken.toFractions(new BigDecimal("500"));
        List<Hash256> txHashes = new ArrayList<>();
        for (Account a : accounts) {
            if (gasToken.getBalanceOf(a).compareTo(minAmount) < 0) {
                NeoSendRawTransaction rawTx = gasToken
                        .transfer(committee, a.getScriptHash(), gasFractions)
                        .getUnsignedTransaction()
                        .addMultiSigWitness(committee.getVerificationScript(), defaultAccount)
                        .send();
                Hash256 txHash = rawTx.getSendRawTransaction()
                        .getHash();
                txHashes.add(txHash);
                System.out.println("Funded account " + a.getAddress());
            }
        }
        for (Hash256 h : txHashes) {
            waitUntilTransactionIsExecuted(h, neow3j);
        }
    }

    static BigInteger getSystemFee(Transaction tx) {
        return BigInteger.valueOf(tx.getSystemFee());
    }

    static BigInteger getNetworkFee(Transaction tx) {
        return BigInteger.valueOf(tx.getNetworkFee());
    }

    static void printFees(String method, Transaction tx) {
        System.out.println("\n############fees############");
        System.out.println(method);
        System.out.printf("System fees:  '%s'\n", tx.getSystemFee());
        System.out.printf("Network fees: '%s'\n", tx.getNetworkFee());
        System.out.println("############fees############\n");
    }

    // region helper methods

    public static Sign.SignatureData createSignature(Hash160 account, BigInteger tea, Account signer) {
        byte[] accountArray = account.toLittleEndianArray();
        byte[] teaArray = reverseArray(tea.toByteArray());
        byte[] message = concatenate(accountArray, teaArray);
        return signMessage(message, signer.getECKeyPair());
    }

    public static Hash160 getHash160FromPublicKey(String publicKey) {
        return Hash160.fromPublicKey(new ECKeyPair.ECPublicKey(publicKey).getEncoded(true));
    }

    public static Hash256 sendAndWaitUntilTransactionIsExecuted(Transaction tx, Neow3j neow3j) throws Throwable {
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
        return txHash;
    }

    public static ContractParameter createMapParam(Hash160[] devs, BigInteger[] teas) {
        HashMap<ContractParameter, ContractParameter> map = new HashMap<>();
        for (int i = 0; i < devs.length; i++) {
            map.put(hash160(devs[i]), integer(teas[i]));
        }
        return map(map);
    }

    static FileWriter getResultFileWriter(String filename) throws IOException {
        File file = openFile(filename + ".csv");
        return getFileWriter(file);
    }

    // endregion helper methods

}
