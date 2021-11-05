package io.flatfeestack;

import io.neow3j.compiler.CompilationUnit;
import io.neow3j.compiler.Compiler;
import io.neow3j.contract.ContractManagement;
import io.neow3j.contract.GasToken;
import io.neow3j.contract.NefFile;
import io.neow3j.contract.PolicyContract;
import io.neow3j.contract.SmartContract;
import io.neow3j.crypto.ECKeyPair;
import io.neow3j.crypto.Sign;
import io.neow3j.protocol.Neow3j;
import io.neow3j.protocol.core.response.ContractManifest;
import io.neow3j.protocol.core.response.InvocationResult;
import io.neow3j.protocol.core.response.NeoApplicationLog;
import io.neow3j.protocol.core.response.NeoSendRawTransaction;
import io.neow3j.protocol.core.stackitem.StackItem;
import io.neow3j.protocol.http.HttpService;
import io.neow3j.serialization.NeoSerializableInterface;
import io.neow3j.transaction.Transaction;
import io.neow3j.transaction.Witness;
import io.neow3j.transaction.exceptions.TransactionConfigurationException;
import io.neow3j.types.ContractParameter;
import io.neow3j.types.Hash160;
import io.neow3j.types.Hash256;
import io.neow3j.types.NeoVMStateType;
import io.neow3j.wallet.Account;
import org.junit.Before;
import org.junit.BeforeClass;
import org.junit.ClassRule;
import org.junit.FixMethodOrder;
import org.junit.Test;
import org.junit.runners.MethodSorters;

import java.io.File;
import java.io.IOException;
import java.math.BigDecimal;
import java.math.BigInteger;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

import static io.flatfeestack.EvaluationHelper.getRandomHashes;
import static io.flatfeestack.EvaluationHelper.getRandomTeasToPreset;
import static io.flatfeestack.EvaluationHelper.getSum;
import static io.flatfeestack.EvaluationHelper.getUniformTeas;
import static io.neow3j.contract.ContractUtils.writeContractManifestFile;
import static io.neow3j.contract.ContractUtils.writeNefFile;
import static io.neow3j.contract.SmartContract.calcContractHash;
import static io.neow3j.contract.Token.toFractions;
import static io.neow3j.crypto.Sign.signMessage;
import static io.neow3j.protocol.ObjectMapperFactory.getObjectMapper;
import static io.neow3j.transaction.AccountSigner.calledByEntry;
import static io.neow3j.transaction.AccountSigner.none;
import static io.neow3j.types.ContractParameter.*;
import static io.neow3j.utils.ArrayUtils.*;
import static io.neow3j.utils.Await.waitUntilTransactionIsExecuted;
import static io.neow3j.wallet.Account.createMultiSigAccount;
import static java.util.Collections.singletonList;
import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.Matchers.containsString;
import static org.hamcrest.Matchers.hasSize;
import static org.hamcrest.core.Is.is;
import static org.junit.Assert.assertNull;
import static org.junit.Assert.assertTrue;

@SuppressWarnings("unchecked")
@FixMethodOrder(MethodSorters.NAME_ASCENDING)
public class PayoutNeoEvalIntTest {

    private static Neow3j neow3j;
    private static GasToken gasToken;
    private static SmartContract payoutContract;

    private static final BigInteger FEE_PER_BYTE = BigInteger.valueOf(100);
    private static final BigInteger STORAGE_PRICE = BigInteger.valueOf(10_000);
    private static final BigInteger EXEC_FEE_FACTION = BigInteger.valueOf(3);

    private static final Path PAYOUT_CONTRACT_NEF = Paths.get("./build/neow3j/PayoutNeoForEvaluation.nef");
    private static final Path PAYOUT_CONTRACT_MANIFEST =
            Paths.get("./build/neow3j/PayoutNeoForEvaluation.manifest.json");

    private static final Account defaultAccount =
            Account.fromWIF("L1eV34wPoj9weqhGijdDLtVQzUpWGHszXXpdU9dPuh2nRFFzFa7E");
    private static final ECKeyPair.ECPublicKey defaultPubKey = defaultAccount.getECKeyPair().getPublicKey();
    private static final Account committee =
            createMultiSigAccount(singletonList(defaultAccount.getECKeyPair().getPublicKey()), 1);

    private static final Account owner =
            Account.fromWIF("L3cNMQUSrvUrHx1MzacwHiUeCWzqK2MLt5fPvJj9mz6L2rzYZpok");
    private static final ECKeyPair.ECPublicKey ownerPubKey = owner.getECKeyPair().getPublicKey();

    private static final BigDecimal contractFundAmount = BigDecimal.valueOf(7000);
    private static final BigInteger devFundAmountFractions = toFractions(BigDecimal.valueOf(100), GasToken.DECIMALS);

    // Methods
    private static final String getOwner = "getOwner";
    private static final String setOwner = "setOwner";
    private static final String getTea = "getTea";
    private static final String setTea = "setTea";
    private static final String withdraw = "withdraw";
    private static final String batchPayout = "batchPayout";
    private static final String batchPayoutWithServiceFee = "batchPayoutWithServiceFee";
    private static final String batchPayoutWithTeas = "batchPayoutWithTeas";
    private static final String batchPayoutWithMap = "batchPayoutWithMap";
    private static final String batchPayoutWithMapAndServiceFee = "batchPayoutWithMapAndServiceFee";
    private static final String batchPayoutWithDoubleMap = "batchPayoutWithDoubleMap";

    private static int nrAccounts;
    private static BigInteger nrAccountsBigInt;
    private static Hash160[] devs;
    private static BigInteger presetTea;
    private static BigInteger[] teas;
    private static BigInteger[] presetTeas;
    private static BigInteger serviceFee;
    private static ContractParameter serviceFeeParam;

    private static BigInteger contractGasBalanceBeforePayout;

    @ClassRule
    public static NeoTestContainer neoTestContainer = new NeoTestContainer();

    // region setup

    @BeforeClass
    public static void setUp() throws Throwable {
        neow3j = Neow3j.build(new HttpService(neoTestContainer.getNodeUrl()));
        gasToken = new GasToken(neow3j);
        handleFeeFactors();
        setTestFactors();
        compileContract(PayoutNeoForEvaluation.class.getCanonicalName());
        System.out.println("\n##############setup#################");
        System.out.println("Owner hash:    " + owner.getScriptHash());
        System.out.println("Owner address: " + owner.getAddress());
        fundAccounts(gasToken.toFractions(BigDecimal.valueOf(10_000)), defaultAccount, owner);
        payoutContract = deployPayoutNeoContract();
        System.out.println("Payout contract hash: " + payoutContract.getScriptHash());
        fundPayoutContract();
        System.out.println("##############setup#################\n");
    }

    @Before
    public void setUpTest() throws IOException {
        devs = getRandomHashes(nrAccounts);
        contractGasBalanceBeforePayout = getContractGasBalance();
    }

    private static void setTestFactors() {
        nrAccounts = 5; // max 404 for all tests to pass
        nrAccountsBigInt = BigInteger.valueOf(nrAccounts);
        presetTea = BigInteger.valueOf(100);
        presetTeas = getRandomTeasToPreset(nrAccounts, 1000, 10);
        teas = getUniformTeas(nrAccounts, BigInteger.valueOf(10000), BigInteger.TEN);
        serviceFee = BigInteger.TEN;
        serviceFeeParam = integer(serviceFee);
    }

    private static void handleFeeFactors() throws Throwable {
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

        BigDecimal actualFeePerByte = gasToken.toDecimals(policyContract.getFeePerByte());
        BigInteger actualStoragePrice = policyContract.getStoragePrice();
        BigInteger actualExecFeeFactor = policyContract.getExecFeeFactor();

        System.out.println("\n##############feefactors#################");
        System.out.printf("Network Fee Per Byte: '%s'\n", actualFeePerByte);
        System.out.printf("Storage Fee Factor:   '%s'\n", actualStoragePrice);
        System.out.printf("Execution Fee Factor: '%s'\n", actualExecFeeFactor);
        System.out.println("##############feefactors#################\n");
    }

    // endregion setup
    // region helper methods

    private Sign.SignatureData createSignature(Hash160 account, BigInteger tea, Account signer) {
        byte[] accountArray = account.toLittleEndianArray();
        byte[] teaArray = reverseArray(tea.toByteArray());
        byte[] message = concatenate(accountArray, teaArray);
        return signMessage(message, signer.getECKeyPair());
    }

    private Hash160 getHash160FromPublicKey(String publicKey) {
        return Hash160.fromPublicKey(new ECKeyPair.ECPublicKey(publicKey).getEncoded(true));
    }

    private static void compileContract(String canonicalName) throws IOException {
        CompilationUnit res = new Compiler().compile(canonicalName);

        // Write contract (compiled, NEF) to the disk
        Path buildNeow3jPath = Paths.get("build", "neow3j");
        buildNeow3jPath.toFile().mkdirs();
        writeNefFile(res.getNefFile(), res.getManifest().getName(), buildNeow3jPath);

        // Write manifest to the disk
        writeContractManifestFile(res.getManifest(), buildNeow3jPath);
    }

    private static void fundPayoutContract() throws Throwable {
        BigInteger fundAmountFractions = gasToken.toFractions(contractFundAmount);
        Hash256 txHash = gasToken
                .transfer(owner, payoutContract.getScriptHash(), fundAmountFractions)
                .sign()
                .send()
                .getSendRawTransaction()
                .getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
        System.out.printf("Contract funded with %s GAS%n", contractFundAmount);
    }

    private static void fundAccounts(BigInteger gasFractions, Account... accounts) throws Throwable {
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

    private static SmartContract deployPayoutNeoContract() throws Throwable {
        File nefFile = new File(PAYOUT_CONTRACT_NEF.toUri());
        NefFile nef = NefFile.readFromFile(nefFile);

        File manifestFile = new File(PAYOUT_CONTRACT_MANIFEST.toUri());
        ContractManifest manifest = getObjectMapper().readValue(manifestFile, ContractManifest.class);
        Hash256 txHash = new ContractManagement(neow3j)
                .deploy(nef, manifest, publicKey(owner.getECKeyPair().getPublicKey().getEncoded(true)))
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

    private BigInteger getContractGasBalance() throws IOException {
        return getGasBalance(payoutContract.getScriptHash());
    }

    private BigInteger getGasBalance(Account account) throws IOException {
        return getGasBalance(account.getScriptHash());
    }

    private BigInteger getGasBalance(Hash160 account) throws IOException {
        return gasToken.getBalanceOf(account);
    }

    private BigInteger getTea(Account account) throws IOException {
        return getTea(account.getScriptHash());
    }

    private BigInteger getTea(Hash160 account) throws IOException {
        return payoutContract.callFuncReturningInt(getTea, hash160(account));
    }

    private Hash256 setTea(Hash160 scriptHash, BigInteger oldTea, BigInteger newTea) throws Throwable {
        Hash256 txHash = payoutContract
                .invokeFunction(setTea, hash160(scriptHash), integer(oldTea), integer(newTea))
                .signers(calledByEntry(owner))
                .sign()
                .send()
                .getSendRawTransaction()
                .getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
        return txHash;
    }

    private Hash256 sendAndWaitUntilTransactionIsExecuted(Transaction tx) throws Throwable {
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
        return txHash;
    }

    private BigInteger getSystemFee(Transaction tx) {
        return BigInteger.valueOf(tx.getSystemFee());
    }

    private BigInteger getNetworkFee(Transaction tx) {
        return BigInteger.valueOf(tx.getNetworkFee());
    }

    private void printFees(String method, Transaction tx) {
        System.out.println("\n############fees############");
        System.out.println(method);
        System.out.printf("System fees:  '%s'\n", tx.getSystemFee());
        System.out.printf("Network fees: '%s'\n", tx.getNetworkFee());
        System.out.println("############fees############\n");
    }

    private ContractParameter createMapParam(Hash160[] devs, BigInteger[] teas) {
        HashMap<ContractParameter, ContractParameter> map = new HashMap<>();
        for (int i = 0; i < devs.length; i++) {
            map.put(hash160(devs[i]), integer(teas[i]));
        }
        return map(map);
    }

    // endregion helper methods

    @Test
    public void verifyThatSecondSignerCannotCoverFees() throws Throwable {
        // Contains a dummy transaction with a random account as first signer with witness scope none and the contract
        // owner as second signer with witness scope calledByEntry. This is the order of the signers in a pre-signed
        // transaction. This test verifies that the contract owner cannot be charged for withdrawal fees with such
        // signature.
        NeoSendRawTransaction tx = payoutContract.invokeFunction("getOwner")
                .signers(none(Account.create()), calledByEntry(owner))
                .sign().send();
        assertNull(tx.getSendRawTransaction());
        assertThat(tx.getError().getMessage(), is("InsufficientFunds"));
    }

    // region test basic contract methods

    @Test
    public void testFundContract() throws Throwable {
        BigInteger contractBalance = getGasBalance(payoutContract.getScriptHash());
        BigInteger fundAmount = gasToken.toFractions(BigDecimal.valueOf(1500L));
        Hash256 txHash = gasToken.transfer(owner, payoutContract.getScriptHash(), fundAmount)
                .signers(calledByEntry(owner))
                .sign()
                .send()
                .getSendRawTransaction()
                .getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        BigInteger balanceAfterTransfer = getGasBalance(payoutContract.getScriptHash());
        assertThat(balanceAfterTransfer, is(contractBalance.add(fundAmount)));
    }

    @Test
    public void testOwner() throws IOException {
        InvocationResult res = payoutContract.callInvokeFunction(getOwner).getInvocationResult();
        String publicKey = res.getStack().get(0).getHexString();
        Hash160 o = getHash160FromPublicKey(publicKey);
        assertThat(o, is(owner.getScriptHash()));
    }

    @Test
    public void testSetOwner() throws Throwable {
        Hash256 txHash = payoutContract.invokeFunction(setOwner, publicKey(defaultPubKey.getEncoded(true)))
                .signers(calledByEntry(owner), calledByEntry(defaultAccount))
                .sign()
                .send()
                .getSendRawTransaction()
                .getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
        List<NeoApplicationLog.Execution> execs = neow3j.getApplicationLog(txHash).send()
                .getApplicationLog().getExecutions();
        assertThat(execs, hasSize(1));
        assertThat(execs.get(0).getState(), is(NeoVMStateType.HALT));

        StackItem item = payoutContract.callInvokeFunction(getOwner).getInvocationResult().getStack().get(0);
        assertThat(item.getByteArray(), is(defaultPubKey.getEncoded(true)));

        // Change owner back to maintain same state for other tests
        txHash = payoutContract.invokeFunction(setOwner, publicKey(ownerPubKey.getEncoded(true)))
                .signers(calledByEntry(owner), calledByEntry(defaultAccount))
                .sign()
                .send()
                .getSendRawTransaction()
                .getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
        execs = neow3j.getApplicationLog(txHash).send().getApplicationLog().getExecutions();
        assertThat(execs, hasSize(1));
        assertThat(execs.get(0).getState(), is(NeoVMStateType.HALT));

        item = payoutContract.callInvokeFunction(getOwner).getInvocationResult().getStack().get(0);
        assertThat(item.getByteArray(), is(ownerPubKey.getEncoded(true)));
    }

    @Test
    public void testGetSetTea() throws Throwable {
        Account randomAccount = Account.create();
        assertThat(getTea(randomAccount), is(BigInteger.ZERO));
        setTea(randomAccount.getScriptHash(), BigInteger.ZERO, BigInteger.valueOf(1234890));
        assertThat(getTea(randomAccount), is(BigInteger.valueOf(1234890)));
        setTea(randomAccount.getScriptHash(), BigInteger.valueOf(1234890), BigInteger.valueOf(1234891));
        assertThat(getTea(randomAccount), is(BigInteger.valueOf(1234891)));
    }

    @Test
    public void testSetTeaInvalidOldTea() throws Throwable {
        Account randomAccount = Account.create();
        assertThat(getTea(randomAccount), is(BigInteger.ZERO));
        boolean failed = false;
        try {
            setTea(randomAccount.getScriptHash(), BigInteger.valueOf(10), BigInteger.valueOf(11));
        } catch (TransactionConfigurationException e) {
            failed = true;
            assertThat(e.getMessage(), containsString("is not equal to the provided oldTea"));
        }
        assertTrue(failed);
    }

    @Test
    public void testSetTeaInvalidNewTea() throws Throwable {
        Account randomAccount = Account.create();
        assertThat(getTea(randomAccount), is(BigInteger.ZERO));
        boolean failed = false;
        try {
            setTea(randomAccount.getScriptHash(), BigInteger.ZERO, BigInteger.valueOf(-1));
        } catch (TransactionConfigurationException e) {
            failed = true;
            assertThat(e.getMessage(), containsString("is lower than or equal to the stored tea."));
        }
        assertTrue(failed);
    }

    // endregion test basic contract methods
    // region test withdraw with signature

    @Test
    public void testWithdrawWithSignature() throws Throwable {
        Account dev = Account.create();
        fundAccounts(devFundAmountFractions, dev);
        BigInteger balanceContractBefore = getContractGasBalance();
        // Dev earned 12 gas for the first time
        BigInteger teaDev = gasToken.toFractions(BigDecimal.valueOf(12));
        // Create a signature
        Sign.SignatureData sigData = createSignature(dev.getScriptHash(), teaDev, owner);
        // Dev invokes withdraw method with signatureData
        Transaction tx = payoutContract.invokeFunction(withdraw, hash160(dev), integer(teaDev), signature(sigData))
                .signers(none(dev)).sign();
        sendAndWaitUntilTransactionIsExecuted(tx);

        assertThat(getContractGasBalance(), is(balanceContractBefore.subtract(teaDev)));
        BigInteger totalFee = getSystemFee(tx).add(getNetworkFee(tx));
        assertThat(getGasBalance(dev), is(devFundAmountFractions.add(teaDev).subtract(totalFee)));
    }

    @Test
    public void testWithdrawAgain() throws Throwable {
        Account dev = Account.create();
        fundAccounts(devFundAmountFractions, dev);
        BigInteger balanceContractBefore = getContractGasBalance();
        BigInteger teaDev = gasToken.toFractions(BigDecimal.valueOf(10));
        Sign.SignatureData sigData = createSignature(dev.getScriptHash(), teaDev, owner);
        Transaction tx = payoutContract.invokeFunction(withdraw, hash160(dev), integer(teaDev), signature(sigData))
                .signers(none(dev)).sign();
        sendAndWaitUntilTransactionIsExecuted(tx);

        assertThat(getContractGasBalance(), is(balanceContractBefore.subtract(teaDev)));
        BigInteger totalFee = getSystemFee(tx).add(getNetworkFee(tx));
        assertThat(getGasBalance(dev), is(devFundAmountFractions.add(teaDev).subtract(totalFee)));
        assertThat(getTea(dev), is(teaDev));

        // Test that signature was invalidated and cannot be used again
        boolean tested = false;
        try {
            payoutContract.invokeFunction(withdraw, hash160(dev), integer(teaDev), signature(sigData))
                    .signers(none(dev)).sign();
        } catch (TransactionConfigurationException e) {
            assertThat(e.getMessage(), containsString("These funds have already been withdrawn."));
            tested = true;
        }
        assertTrue(tested);
    }

    // endregion withdraw with signature
    // region withdraw with witness

    @Test
    public void testWithdrawWithWitness() throws Throwable {
        Account dev = Account.create();
        fundAccounts(devFundAmountFractions, dev);
        BigInteger balanceContractBefore = getContractGasBalance();
        BigInteger teaDev = gasToken.toFractions(BigDecimal.valueOf(22));
        SmartContract smartContract = new SmartContract(payoutContract.getScriptHash(), neow3j);


        Transaction unsignedTransaction = smartContract.invokeFunction("withdraw", hash160(dev), integer(1))
                .signers(calledByEntry(owner))
                .getUnsignedTransaction();
        byte[] unsignedTxBytes = unsignedTransaction.toArray();
        Transaction tx = NeoSerializableInterface.from(unsignedTxBytes, Transaction.class);
        tx.setNeow3j(neow3j);

        Transaction txWithoutWitnesses = payoutContract.invokeFunction(withdraw, hash160(dev), integer(teaDev))
                .signers(none(dev), calledByEntry(owner).setAllowedContracts(payoutContract.getScriptHash()))
                .getUnsignedTransaction();

        Witness ownerWitness = Witness.create(txWithoutWitnesses.getHashData(), owner.getECKeyPair());
        byte[] witnessBytes = ownerWitness.toArray();
        byte[] preSignedTxBytes = txWithoutWitnesses.toArray();

        // The following steps are done by the dev after receiving the transaction and witness bytes.
        tx = NeoSerializableInterface.from(preSignedTxBytes, Transaction.class);
        tx.setNeow3j(neow3j);
        Witness devWitness = Witness.create(tx.getHashData(), dev.getECKeyPair());
        tx.addWitness(devWitness);
        Witness ownerWitnessFromBytes = NeoSerializableInterface.from(witnessBytes, Witness.class);
        tx.addWitness(ownerWitnessFromBytes);
        sendAndWaitUntilTransactionIsExecuted(tx);

        assertThat(getContractGasBalance(), is(balanceContractBefore.subtract(teaDev)));
        BigInteger totalFee = getSystemFee(tx).add(getNetworkFee(tx));
        assertThat(getGasBalance(dev), is(devFundAmountFractions.add(teaDev).subtract(totalFee)));
        assertThat(getTea(dev), is(teaDev));
    }

    @Test
    public void testWithdrawWithWitnessAgain() throws Throwable {
        Account dev = Account.create();
        fundAccounts(devFundAmountFractions, dev);
        BigInteger balanceContractBefore = getContractGasBalance();
        BigInteger teaDev = gasToken.toFractions(BigDecimal.valueOf(22));
        Transaction txToBePreSigned = payoutContract.invokeFunction(withdraw, hash160(dev), integer(teaDev))
                .signers(none(dev), calledByEntry(owner).setAllowedContracts(payoutContract.getScriptHash()))
                .getUnsignedTransaction();

        Witness ownerWitness = Witness.create(txToBePreSigned.getHashData(), owner.getECKeyPair());
        byte[] witnessBytes = ownerWitness.toArray();
        byte[] preSignedTxBytes = txToBePreSigned.toArray();

        // The following steps are done by the dev after receiving the transaction and witness bytes.
        Transaction tx = NeoSerializableInterface.from(preSignedTxBytes, Transaction.class);
        tx.setNeow3j(neow3j);
        Witness devWitness = Witness.create(tx.getHashData(), dev.getECKeyPair());
        tx.addWitness(devWitness);
        Witness ownerWitnessFromBytes = NeoSerializableInterface.from(witnessBytes, Witness.class);
        tx.addWitness(ownerWitnessFromBytes);
        sendAndWaitUntilTransactionIsExecuted(tx);

        assertThat(getContractGasBalance(), is(balanceContractBefore.subtract(teaDev)));
        BigInteger totalFee = getSystemFee(tx).add(getNetworkFee(tx));
        assertThat(getGasBalance(dev), is(devFundAmountFractions.add(teaDev).subtract(totalFee)));
        assertThat(getTea(dev), is(teaDev));

        // Create transaction and witness from bytes again and try to send it again.
        tx = NeoSerializableInterface.from(preSignedTxBytes, Transaction.class);
        tx.setNeow3j(neow3j);
        devWitness = Witness.create(tx.getHashData(), dev.getECKeyPair());
        tx.addWitness(devWitness);
        ownerWitnessFromBytes = NeoSerializableInterface.from(witnessBytes, Witness.class);
        tx.addWitness(ownerWitnessFromBytes);

        NeoSendRawTransaction rawTx = tx.send();
        assertNull(rawTx.getSendRawTransaction());
        assertThat(rawTx.getError().getMessage(), containsString("AlreadyExists"));
    }

    // endregion withdraw with witness
    // region test batch payout methods

    @Test
    public void test1_batchPayout() throws Throwable {
        ContractParameter accountsParam = array(devs);
        ContractParameter teasParam = array(teas);
        Transaction tx = payoutContract.invokeFunction(batchPayout, accountsParam, teasParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teas[i]));
            assertThat(getGasBalance(devs[i]), is(teas[i]));
        }
        BigInteger summedUpPayoutAmount = getSum(teas);
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test1 - list", tx);
    }

    @Test
    public void test2_withStoredTeas_batchPayoutList() throws Throwable {
        for (Hash160 dev : devs) {
            setTea(dev, BigInteger.ZERO, presetTea);
        }
        ContractParameter accountsParam = array(devs);
        ContractParameter teasParam = array(teas);
        Transaction tx = payoutContract.invokeFunction(batchPayout, accountsParam, teasParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teas[i]));
            assertThat(getGasBalance(devs[i]), is(teas[i].subtract(presetTea)));
        }
        BigInteger summedUpPayoutAmount = getSum(teas).subtract(presetTea.multiply(nrAccountsBigInt));
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test 2 - list", tx);
    }

    @Test
    public void test3_withDiverseStoredTeas_batchPayoutList() throws Throwable {
        for (int i = 0; i < nrAccounts; i++) {
            setTea(devs[i], BigInteger.ZERO, presetTeas[i]);
        }
        ContractParameter accountsParam = array(devs);
        ContractParameter teasParam = array(teas);
        Transaction tx = payoutContract.invokeFunction(batchPayout, accountsParam, teasParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teas[i]));
            assertThat(getGasBalance(devs[i]), is(teas[i].subtract(presetTeas[i])));
        }

        BigInteger summedUpPayoutAmount = getSum(teas).subtract(getSum(presetTeas));
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test 3 - list", tx);
    }

    // endregion batch payout
    // region batch payout with service fee

    @Test
    public void test1_batchPayoutListWithServiceFee() throws Throwable {
        ContractParameter accountsParam = array(devs);
        ContractParameter teasParam = array(teas);

        Transaction tx = payoutContract.invokeFunction(batchPayoutWithServiceFee, accountsParam, teasParam,
                        serviceFeeParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teas[i]));
            assertThat(getGasBalance(devs[i]), is(teas[i].subtract(serviceFee)));
        }
        BigInteger summedUpPayoutAmount = getSum(teas).subtract(serviceFee.multiply(nrAccountsBigInt));
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test1 - list with service fee", tx);
    }

    @Test
    public void test2_withStoredTeas_batchPayoutListWithServiceFee() throws Throwable {
        for (Hash160 dev : devs) {
            setTea(dev, BigInteger.ZERO, presetTea);
        }
        ContractParameter accountsParam = array(devs);
        ContractParameter teasParam = array(teas);
        Transaction tx = payoutContract.invokeFunction(batchPayoutWithServiceFee, accountsParam, teasParam,
                        serviceFeeParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teas[i]));
            assertThat(getGasBalance(devs[i]), is(teas[i].subtract(presetTea).subtract(serviceFee)));
        }
        BigInteger summedUpPayoutAmount = getSum(teas)
                .subtract(presetTea.multiply(nrAccountsBigInt).add(serviceFee.multiply(nrAccountsBigInt)));
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test2 - list with service fee", tx);
    }

    @Test
    public void test3_withDiverseStoredTeas_batchPayoutListWithServiceFee() throws Throwable {
        for (int i = 0; i < nrAccounts; i++) {
            setTea(devs[i], BigInteger.ZERO, presetTeas[i]);
        }
        ContractParameter accountsParam = array(devs);
        ContractParameter teasParam = array(teas);
        Transaction tx = payoutContract.invokeFunction(batchPayoutWithServiceFee, accountsParam, teasParam,
                        serviceFeeParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teas[i]));
            assertThat(getGasBalance(devs[i]), is(teas[i].subtract(presetTeas[i]).subtract(serviceFee)));
        }

        BigInteger summedUpPayoutAmount = getSum(teas).subtract(getSum(presetTeas))
                .subtract(serviceFee.multiply(nrAccountsBigInt));
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test3 - list with service fee", tx);
    }

    // endregion batch payout with service fee
    // region batch payout with two lists

    @Test
    public void test1_batchPayoutListList() throws Throwable {
        BigInteger[] teasToStore = getUniformTeas(nrAccounts, BigInteger.valueOf(10001), BigInteger.ONE);
        BigInteger[] teasForWithdrawal = getUniformTeas(nrAccounts, BigInteger.valueOf(10000), BigInteger.ONE);
        ContractParameter accountsParam = array(devs);
        ContractParameter teasToStoreParam = array(teasToStore);
        ContractParameter teasForWithdrawalParam = array(teasForWithdrawal);

        Transaction tx = payoutContract.invokeFunction(batchPayoutWithTeas, accountsParam, teasToStoreParam,
                        teasForWithdrawalParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teasToStore[i]));
            assertThat(getGasBalance(devs[i]), is(teasForWithdrawal[i]));
        }
        BigInteger summedUpPayoutAmount = getSum(teasForWithdrawal);
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test1 - list list", tx);
    }

    @Test
    public void test2_withStoredTeas_batchPayoutListList() throws Throwable {
        BigInteger[] teasToStore = getUniformTeas(nrAccounts, BigInteger.valueOf(10001), BigInteger.ONE);
        BigInteger[] teasForWithdrawal = getUniformTeas(nrAccounts, BigInteger.valueOf(10000), BigInteger.ONE);
        for (Hash160 dev : devs) {
            setTea(dev, BigInteger.ZERO, presetTea);
        }
        ContractParameter accountsParam = array(devs);
        ContractParameter teasToStoreParam = array(teasToStore);
        ContractParameter teasForWithdrawalParam = array(teasForWithdrawal);
        Transaction tx = payoutContract.invokeFunction(batchPayoutWithTeas, accountsParam, teasToStoreParam,
                        teasForWithdrawalParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teasToStore[i]));
            assertThat(getGasBalance(devs[i]), is(teasForWithdrawal[i].subtract(presetTea)));
        }
        BigInteger summedUpPayoutAmount = getSum(teasForWithdrawal).subtract(presetTea.multiply(nrAccountsBigInt));
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test2 - list list", tx);
    }

    @Test
    public void test3_withDiverseStoredTeas_batchPayoutListList() throws Throwable {
        BigInteger[] teasToStore = getUniformTeas(nrAccounts, BigInteger.valueOf(10001), BigInteger.valueOf(2));
        BigInteger[] teasForWithdrawal = getUniformTeas(nrAccounts, BigInteger.valueOf(10000), BigInteger.valueOf(5));
        for (int i = 0; i < nrAccounts; i++) {
            setTea(devs[i], BigInteger.ZERO, presetTeas[i]);
        }
        ContractParameter accountsParam = array(devs);
        ContractParameter teasToStoreParam = array(teasToStore);
        ContractParameter teasForWithdrawalParam = array(teasForWithdrawal);
        Transaction tx = payoutContract.invokeFunction(batchPayoutWithTeas, accountsParam, teasToStoreParam,
                        teasForWithdrawalParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teasToStore[i]));
            assertThat(getGasBalance(devs[i]), is(teasForWithdrawal[i].subtract(presetTeas[i])));
        }

        BigInteger summedUpPayoutAmount = getSum(teasForWithdrawal).subtract(getSum(presetTeas));
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test3 - list list", tx);
    }

    // endregion batch payout with two lists
    // region batch payout with map

    @Test
    public void test1_batchPayoutMap() throws Throwable {
        ContractParameter payoutMapParam = createMapParam(devs, teas);
        Transaction tx = payoutContract.invokeFunction(batchPayoutWithMap, payoutMapParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teas[i]));
            assertThat(getGasBalance(devs[i]), is(teas[i]));
        }
        BigInteger summedUpPayoutAmount = getSum(teas);
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test1 - map", tx);
    }

    @Test
    public void test2_withStoredTeas_batchPayoutMap() throws Throwable {
        for (Hash160 dev : devs) {
            setTea(dev, BigInteger.ZERO, presetTea);
        }
        ContractParameter payoutMapParam = createMapParam(devs, teas);
        Transaction tx = payoutContract.invokeFunction(batchPayoutWithMap, payoutMapParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teas[i]));
            assertThat(getGasBalance(devs[i]), is(teas[i].subtract(presetTea)));
        }
        BigInteger summedUpPayoutAmount = getSum(teas).subtract(presetTea.multiply(nrAccountsBigInt));
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test2 - map", tx);
    }

    @Test
    public void test3_withDiverseStoredTeas_batchPayoutMap() throws Throwable {
        BigInteger contractGasBalanceBeforePayout = getContractGasBalance();
        for (int i = 0; i < nrAccounts; i++) {
            setTea(devs[i], BigInteger.ZERO, presetTeas[i]);
        }
        ContractParameter payoutMapParam = createMapParam(devs, teas);
        Transaction tx = payoutContract.invokeFunction(batchPayoutWithMap, payoutMapParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teas[i]));
            assertThat(getGasBalance(devs[i]), is(teas[i].subtract(presetTeas[i])));
        }

        BigInteger summedUpPayoutAmount = getSum(teas).subtract(getSum(presetTeas));
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test3 - map", tx);
    }

    // endregion batch payout with map
    // region batch payout with map and service fee

    @Test
    public void test1_batchPayoutMapServiceFee() throws Throwable {
        ContractParameter payoutMapParam = createMapParam(devs, teas);
        Transaction tx = payoutContract.invokeFunction(batchPayoutWithMapAndServiceFee, payoutMapParam,
                        integer(serviceFee))
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teas[i]));
            assertThat(getGasBalance(devs[i]), is(teas[i].subtract(serviceFee)));
        }

        BigInteger summedUpPayoutAmount = getSum(teas).subtract(serviceFee.multiply(nrAccountsBigInt));
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test1 - map with service fee", tx);
    }

    @Test
    public void test2_withStoredTeas_batchPayoutMapServiceFee() throws Throwable {
        ContractParameter payoutMapParam = createMapParam(devs, teas);
        for (Hash160 dev : devs) {
            setTea(dev, BigInteger.ZERO, presetTea);
        }
        Transaction tx = payoutContract.invokeFunction(batchPayoutWithMapAndServiceFee, payoutMapParam,
                        integer(serviceFee))
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teas[i]));
            assertThat(getGasBalance(devs[i]), is(teas[i].subtract(presetTea).subtract(serviceFee)));
        }

        BigInteger summedUpPayoutAmount = getSum(teas).subtract(presetTea.multiply(nrAccountsBigInt))
                .subtract(serviceFee.multiply(nrAccountsBigInt));
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test2 - map with service fee", tx);
    }

    @Test
    public void test3_withDiverseStoredTeas_batchPayoutMapServiceFee() throws Throwable {
        ContractParameter payoutMapParam = createMapParam(devs, teas);
        for (int i = 0; i < nrAccounts; i++) {
            setTea(devs[i], BigInteger.ZERO, presetTeas[i]);
        }
        Transaction tx = payoutContract.invokeFunction(batchPayoutWithMapAndServiceFee, payoutMapParam,
                        integer(serviceFee))
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teas[i]));
            assertThat(getGasBalance(devs[i]), is(teas[i].subtract(presetTeas[i]).subtract(serviceFee)));
        }

        BigInteger summedUpPayoutAmount = getSum(teas).subtract(getSum(presetTeas))
                .subtract(serviceFee.multiply(nrAccountsBigInt));
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test3 - map with service fee", tx);
    }

    // endregion batch payout with map and service fee
    // region batch payout with two maps

    @Test
    public void test1_batchPayoutMapMap() throws Throwable {
        BigInteger[] teasToStore = getUniformTeas(nrAccounts, BigInteger.valueOf(10001), BigInteger.ONE);
        BigInteger[] teasForWithdrawal = getUniformTeas(nrAccounts, BigInteger.valueOf(10000), BigInteger.ONE);
        ContractParameter mapToStoreParam = createMapParam(devs, teasToStore);
        ContractParameter mapForWithdrawParam = createMapParam(devs, teasForWithdrawal);

        Transaction tx = payoutContract.invokeFunction(batchPayoutWithDoubleMap, mapToStoreParam, mapForWithdrawParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teasToStore[i]));
            assertThat(getGasBalance(devs[i]), is(teasForWithdrawal[i]));
        }

        BigInteger summedUpPayoutAmount = getSum(teasForWithdrawal);
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test1 - double map", tx);
    }

    @Test
    public void test2_withStoredTeas_batchPayoutMapMap() throws Throwable {
        BigInteger[] teasToStore = getUniformTeas(nrAccounts, BigInteger.valueOf(10001), BigInteger.TEN);
        BigInteger[] teasForWithdrawal = getUniformTeas(nrAccounts, BigInteger.valueOf(10000), BigInteger.TEN);
        ContractParameter mapToStoreParam = createMapParam(devs, teasToStore);
        ContractParameter mapForWithdrawParam = createMapParam(devs, teasForWithdrawal);

        for (Hash160 dev : devs) {
            setTea(dev, BigInteger.ZERO, presetTea);
        }

        Transaction tx = payoutContract.invokeFunction(batchPayoutWithDoubleMap, mapToStoreParam, mapForWithdrawParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teasToStore[i]));
            assertThat(getGasBalance(devs[i]), is(teasForWithdrawal[i].subtract(presetTea)));
        }

        BigInteger summedUpPayoutAmount = getSum(teasForWithdrawal).subtract(presetTea.multiply(nrAccountsBigInt));
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test2 - double map", tx);
    }

    @Test
    public void test3_withDiverseStoredTeas_batchPayoutMapMap() throws Throwable {
        BigInteger[] teasToStore = getUniformTeas(nrAccounts, BigInteger.valueOf(10001), BigInteger.TEN);
        BigInteger[] teasForWithdrawal = getUniformTeas(nrAccounts, BigInteger.valueOf(10000), BigInteger.TEN);
        ContractParameter mapToStoreParam = createMapParam(devs, teasToStore);
        ContractParameter mapForWithdrawParam = createMapParam(devs, teasForWithdrawal);

        for (int i = 0; i < nrAccounts; i++) {
            setTea(devs[i], BigInteger.ZERO, presetTeas[i]);
        }

        Transaction tx = payoutContract.invokeFunction(batchPayoutWithDoubleMap, mapToStoreParam, mapForWithdrawParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(devs[i]), is(teasToStore[i]));
            assertThat(getGasBalance(devs[i]), is(teasForWithdrawal[i].subtract(presetTeas[i])));
        }

        BigInteger summedUpPayoutAmount = getSum(teasForWithdrawal).subtract(getSum(presetTeas));
        BigInteger contractGasBalanceAfterPayout = getContractGasBalance();
        assertThat(contractGasBalanceAfterPayout,
                is(contractGasBalanceBeforePayout.subtract(summedUpPayoutAmount)));
        printFees("test3 - double map", tx);
    }

    // endregion batch payout with two maps

}
