package io.flatfeestack;

import io.neow3j.contract.GasToken;
import io.neow3j.contract.SmartContract;
import io.neow3j.crypto.Sign;
import io.neow3j.protocol.Neow3j;
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
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.ExpectedException;
import org.junit.runners.MethodSorters;

import java.io.IOException;
import java.math.BigDecimal;
import java.math.BigInteger;
import java.util.Arrays;
import java.util.List;

import static io.flatfeestack.EvaluationHelper.committee;
import static io.flatfeestack.EvaluationHelper.compileContract;
import static io.flatfeestack.EvaluationHelper.createSignature;
import static io.flatfeestack.EvaluationHelper.defaultAccount;
import static io.flatfeestack.EvaluationHelper.defaultPubKey;
import static io.flatfeestack.EvaluationHelper.deployPayoutNeoContract;
import static io.flatfeestack.EvaluationHelper.fundAccounts;
import static io.flatfeestack.EvaluationHelper.fundContract;
import static io.flatfeestack.EvaluationHelper.getHash160FromPublicKey;
import static io.flatfeestack.EvaluationHelper.getNetworkFee;
import static io.flatfeestack.EvaluationHelper.getRandomHashes;
import static io.flatfeestack.EvaluationHelper.getRandomTeasToPreset;
import static io.flatfeestack.EvaluationHelper.getSum;
import static io.flatfeestack.EvaluationHelper.getSystemFee;
import static io.flatfeestack.EvaluationHelper.getUniformTeas;
import static io.flatfeestack.EvaluationHelper.handleFeeFactors;
import static io.flatfeestack.EvaluationHelper.owner;
import static io.flatfeestack.EvaluationHelper.ownerPubKey;
import static io.flatfeestack.EvaluationHelper.printFees;
import static io.flatfeestack.EvaluationHelper.sendAndWaitUntilTransactionIsExecuted;
import static io.neow3j.contract.Token.toFractions;
import static io.neow3j.transaction.AccountSigner.calledByEntry;
import static io.neow3j.transaction.AccountSigner.none;
import static io.neow3j.types.ContractParameter.array;
import static io.neow3j.types.ContractParameter.hash160;
import static io.neow3j.types.ContractParameter.integer;
import static io.neow3j.types.ContractParameter.publicKey;
import static io.neow3j.types.ContractParameter.signature;
import static io.neow3j.utils.Await.waitUntilTransactionIsExecuted;
import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.Matchers.containsString;
import static org.hamcrest.Matchers.hasSize;
import static org.hamcrest.Matchers.not;
import static org.hamcrest.core.Is.is;
import static org.junit.Assert.assertNull;
import static org.junit.Assert.assertTrue;

@SuppressWarnings("unchecked")
@FixMethodOrder(MethodSorters.NAME_ASCENDING)
public class PayoutNeoIntegrationTest {

    private static Neow3j neow3j;
    private static GasToken gasToken;
    private static SmartContract payoutContract;

    private static final BigDecimal contractFundAmount = BigDecimal.valueOf(7000);
    private static final BigInteger devFundAmountFractions = toFractions(BigDecimal.valueOf(100), GasToken.DECIMALS);

    // Methods
    private static final String getOwner = "getOwner";
    private static final String changeOwner = "changeOwner";
    private static final String getTea = "getTea";
    private static final String setTea = "setTea";
    private static final String setTeas = "setTeas";
    private static final String withdraw = "withdraw";
    private static final String batchPayout = "batchPayout";

    private static int nrAccounts;
    private static BigInteger nrAccountsBigInt;
    private static Hash160[] devs;
    private static BigInteger presetTea;
    private static BigInteger[] teas;
    private static BigInteger[] presetTeas;

    private static BigInteger contractGasBalanceBeforePayout;

    @Rule
    public ExpectedException exceptionRule = ExpectedException.none();

    @ClassRule
    public static NeoTestContainer neoTestContainer = new NeoTestContainer();

    // region setup

    @BeforeClass
    public static void setUp() throws Throwable {
        neow3j = Neow3j.build(new HttpService(neoTestContainer.getNodeUrl()));
        gasToken = new GasToken(neow3j);
        handleFeeFactors(neow3j, committee, defaultAccount);
        setTestFactors();
        compileContract(PayoutNeo.class.getCanonicalName());
        System.out.println("\n##############setup#################");
        System.out.printf("Owner hash:    '%s'\n", owner.getScriptHash());
        System.out.printf("Owner address: '%s'\n", owner.getAddress());
        fundAccounts(neow3j, gasToken.toFractions(BigDecimal.valueOf(10_000)), defaultAccount, owner);
        payoutContract = deployPayoutNeoContract(neow3j, true);
        System.out.printf("Payout contract hash: '%s'", payoutContract.getScriptHash());
        fundContract(neow3j, payoutContract, contractFundAmount);
        System.out.println("##############setup#################\n");
    }

    @Before
    public void setUpTest() throws IOException {
        devs = getRandomHashes(nrAccounts);
        contractGasBalanceBeforePayout = getContractGasBalance();
    }

    private static void setTestFactors() {
        nrAccounts = 10; // max 404 for all tests to pass
        nrAccountsBigInt = BigInteger.valueOf(nrAccounts);
        presetTea = BigInteger.valueOf(100);
        presetTeas = getRandomTeasToPreset(nrAccounts, 1000, 10);
        teas = getUniformTeas(nrAccounts, BigInteger.valueOf(10000), BigInteger.TEN);
    }

    // endregion setup
    // region helper methods

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

    private void setTea(Hash160 scriptHash, BigInteger oldTea, BigInteger newTea) throws Throwable {
        Hash256 txHash = payoutContract
                .invokeFunction(setTea, hash160(scriptHash), integer(oldTea), integer(newTea))
                .signers(calledByEntry(owner))
                .sign()
                .send()
                .getSendRawTransaction()
                .getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
    }

    private void setTeas(Hash160[] scriptHash, BigInteger[] oldTea, BigInteger[] newTea) throws Throwable {
        Hash256 txHash = payoutContract
                .invokeFunction(setTeas, array(scriptHash), array(oldTea), array(newTea))
                .signers(calledByEntry(owner))
                .sign()
                .send()
                .getSendRawTransaction()
                .getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
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
    public void testChangeOwner() throws Throwable {
        Hash256 txHash = payoutContract.invokeFunction(changeOwner, publicKey(defaultPubKey.getEncoded(true)))
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
        txHash = payoutContract.invokeFunction(changeOwner, publicKey(ownerPubKey.getEncoded(true)))
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

    @Test
    public void testSetTeas() throws Throwable {
        int nrAccounts = 40;
        Hash160[] accounts = new Hash160[nrAccounts];
        BigInteger[] oldTeas = new BigInteger[nrAccounts];
        Arrays.fill(oldTeas, BigInteger.ZERO);
        BigInteger[] newTeas = new BigInteger[nrAccounts];
        for (int i = 0; i < nrAccounts; i++) {
            Hash160 randomAccount = Account.create().getScriptHash();
            accounts[i] = randomAccount;
            assertThat(getTea(randomAccount), is(BigInteger.ZERO));
            newTeas[i] = BigInteger.valueOf((long) (Math.random() * 100000L) + 1);
        }
        setTeas(accounts, oldTeas, newTeas);
        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(accounts[i]), is(newTeas[i]));
        }
    }

    @Test
    public void testSetTeas_withIncorrectOldTea() throws Throwable {
        int nrAccounts = 10;
        Hash160[] accounts = new Hash160[nrAccounts];
        BigInteger[] oldTeas = new BigInteger[nrAccounts];
        Arrays.fill(oldTeas, BigInteger.ZERO);
        oldTeas[3] = BigInteger.ONE; // Incorrect oldTea
        BigInteger[] newTeas = new BigInteger[nrAccounts];
        for (int i = 0; i < nrAccounts; i++) {
            Hash160 randomAccount = Account.create().getScriptHash();
            accounts[i] = randomAccount;
            assertThat(getTea(randomAccount), is(BigInteger.ZERO));
            newTeas[i] = BigInteger.valueOf((long) (Math.random() * 100000L) + 1);
        }
        setTeas(accounts, oldTeas, newTeas);
        for (int i = 0; i < nrAccounts; i++) {
            if (i == 3) {
                // Assert that tea was not updated for this account
                assertThat(getTea(accounts[i]), is(BigInteger.ZERO));
            } else {
                assertThat(newTeas[i], is(not(BigInteger.ZERO)));
                assertThat(getTea(accounts[i]), is(newTeas[i]));
            }
        }
    }

    @Test
    public void testSetTeas_withLowerNewTea() throws Throwable {
        int nrAccounts = 14;
        Hash160[] accounts = new Hash160[nrAccounts];
        BigInteger[] oldTeas = new BigInteger[nrAccounts];
        Arrays.fill(oldTeas, BigInteger.ZERO);
        BigInteger[] newTeas = new BigInteger[nrAccounts];
        for (int i = 0; i < nrAccounts; i++) {
            Hash160 randomAccount = Account.create().getScriptHash();
            accounts[i] = randomAccount;
            assertThat(getTea(randomAccount), is(BigInteger.ZERO));
            newTeas[i] = BigInteger.valueOf((long) (Math.random() * 100000L) + 10);
        }
        BigInteger[] preset = new BigInteger[nrAccounts];
        Arrays.fill(preset, BigInteger.valueOf(8));
        setTeas(accounts, oldTeas, preset);
        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(accounts[i]), is(BigInteger.valueOf(8)));
        }
        newTeas[6] = BigInteger.valueOf(5); // Lower than the set tea
        setTeas(accounts, preset, newTeas);
        for (int i = 0; i < nrAccounts; i++) {
            if (i == 6) {
                // Assert that tea was not updated for this account
                assertThat(getTea(accounts[i]), is(BigInteger.valueOf(8)));
            } else {
                assertThat(newTeas[i], is(not(BigInteger.ZERO)));
                assertThat(getTea(accounts[i]), is(newTeas[i]));
            }
        }
    }

    @Test
    public void testSetTeas_withIncorrectLengths() throws Throwable {
        int nrAccounts = 10;
        Hash160[] accounts = new Hash160[nrAccounts];
        BigInteger[] oldTeas = new BigInteger[nrAccounts - 1]; // Incorrect length
        Arrays.fill(oldTeas, BigInteger.ZERO);
        BigInteger[] newTeas = new BigInteger[nrAccounts];
        for (int i = 0; i < nrAccounts; i++) {
            Hash160 randomAccount = Account.create().getScriptHash();
            accounts[i] = randomAccount;
            assertThat(getTea(randomAccount), is(BigInteger.ZERO));
            newTeas[i] = BigInteger.valueOf((long) (Math.random() * 100000L) + 1);
        }
        boolean tested = false;
        try {
            setTeas(accounts, oldTeas, newTeas);
        } catch (TransactionConfigurationException e) {
            assertThat(e.getMessage(), containsString("Parameters must have same length."));
            tested = true;
        }
        assertTrue(tested);
    }

    @Test
    public void testSetTeas_withIncorrectLengths_newTea() throws Throwable {
        int nrAccounts = 10;
        Hash160[] accounts = new Hash160[nrAccounts];
        BigInteger[] oldTeas = new BigInteger[nrAccounts];
        Arrays.fill(oldTeas, BigInteger.ZERO);
        int newTeasLength = nrAccounts - 5;
        BigInteger[] newTeas = new BigInteger[newTeasLength]; // Incorrect length
        for (int i = 0; i < nrAccounts; i++) {
            Hash160 randomAccount = Account.create().getScriptHash();
            accounts[i] = randomAccount;
            assertThat(getTea(randomAccount), is(BigInteger.ZERO));
            if (i < newTeasLength) {
                newTeas[i] = BigInteger.valueOf((long) (Math.random() * 100000L) + 1);
            }
        }
        boolean tested = false;
        try {
            setTeas(accounts, oldTeas, newTeas);
        } catch (TransactionConfigurationException e) {
            assertThat(e.getMessage(), containsString("Parameters must have same length."));
            tested = true;
        }
        assertTrue(tested);
    }

    // endregion test basic contract methods
    // region test withdraw with signature

    @Test
    public void testWithdrawWithSignature() throws Throwable {
        Account dev = Account.create();
        fundAccounts(neow3j, devFundAmountFractions, dev);
        BigInteger balanceContractBefore = getContractGasBalance();
        // Dev earned 12 gas for the first time
        BigInteger teaDev = gasToken.toFractions(BigDecimal.valueOf(12));
        // Create a signature
        Sign.SignatureData sigData = createSignature(dev.getScriptHash(), teaDev, owner);
        // Dev invokes withdraw method with signatureData
        Transaction tx = payoutContract.invokeFunction(withdraw, hash160(dev), integer(teaDev), signature(sigData))
                .signers(none(dev)).sign();
        sendAndWaitUntilTransactionIsExecuted(tx, neow3j);

        assertThat(getContractGasBalance(), is(balanceContractBefore.subtract(teaDev)));
        BigInteger totalFee = getSystemFee(tx).add(getNetworkFee(tx));
        assertThat(getGasBalance(dev), is(devFundAmountFractions.add(teaDev).subtract(totalFee)));
        printFees("withdraw sig", tx);
    }

    @Test
    public void testWithdrawAgain() throws Throwable {
        Account dev = Account.create();
        fundAccounts(neow3j, devFundAmountFractions, dev);
        BigInteger balanceContractBefore = getContractGasBalance();
        BigInteger teaDev = gasToken.toFractions(BigDecimal.valueOf(10));
        Sign.SignatureData sigData = createSignature(dev.getScriptHash(), teaDev, owner);
        Transaction tx = payoutContract.invokeFunction(withdraw, hash160(dev), integer(teaDev), signature(sigData))
                .signers(none(dev)).sign();
        sendAndWaitUntilTransactionIsExecuted(tx, neow3j);

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
        fundAccounts(neow3j, devFundAmountFractions, dev);
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
        sendAndWaitUntilTransactionIsExecuted(tx, neow3j);

        assertThat(getContractGasBalance(), is(balanceContractBefore.subtract(teaDev)));
        BigInteger totalFee = getSystemFee(tx).add(getNetworkFee(tx));
        assertThat(getGasBalance(dev), is(devFundAmountFractions.add(teaDev).subtract(totalFee)));
        assertThat(getTea(dev), is(teaDev));
        printFees("withdraw witness", tx);
    }

    @Test
    public void testWithdrawWithWitnessAgain() throws Throwable {
        Account dev = Account.create();
        fundAccounts(neow3j, devFundAmountFractions, dev);
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
        sendAndWaitUntilTransactionIsExecuted(tx, neow3j);

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
    // region test batch payout with insufficient balance

    @Test
    public void test4_insufficientBalanceToPayoutAllAccounts_batchPayoutList() throws Throwable {
        BigInteger contractGasBalance = getContractGasBalance();
        Hash160[] devs = new Hash160[2];
        devs[0] = PayoutNeoIntegrationTest.devs[0];
        devs[1] = PayoutNeoIntegrationTest.devs[1];
        BigInteger[] teas = new BigInteger[2];
        teas[0] = contractGasBalance;
        teas[1] = BigInteger.ONE;
        ContractParameter accountsParam = array(devs);
        ContractParameter teasParam = array(teas);

        exceptionRule.expect(TransactionConfigurationException.class);
        exceptionRule.expectMessage("Transfer was not successful");
        payoutContract.invokeFunction(batchPayout, accountsParam, teasParam)
                .signers(calledByEntry(owner)).sign();
    }

    // endregion test batch payout with insufficient balance

}
