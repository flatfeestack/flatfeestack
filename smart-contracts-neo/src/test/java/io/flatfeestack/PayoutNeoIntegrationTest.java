package io.flatfeestack;

import io.neow3j.contract.GasToken;
import io.neow3j.contract.SmartContract;
import io.neow3j.crypto.Sign;
import io.neow3j.protocol.Neow3j;
import io.neow3j.protocol.core.response.InvocationResult;
import io.neow3j.protocol.core.response.NeoApplicationLog;
import io.neow3j.protocol.core.response.NeoSendRawTransaction;
import io.neow3j.protocol.core.stackitem.StackItem;
import io.neow3j.serialization.NeoSerializableInterface;
import io.neow3j.test.ContractTest;
import io.neow3j.test.ContractTestExtension;
import io.neow3j.test.DeployConfig;
import io.neow3j.test.DeployConfiguration;
import io.neow3j.transaction.Transaction;
import io.neow3j.transaction.Witness;
import io.neow3j.transaction.exceptions.TransactionConfigurationException;
import io.neow3j.types.ContractParameter;
import io.neow3j.types.Hash160;
import io.neow3j.types.Hash256;
import io.neow3j.types.NeoVMStateType;
import io.neow3j.wallet.Account;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.RegisterExtension;

import java.io.IOException;
import java.math.BigDecimal;
import java.math.BigInteger;
import java.util.Arrays;
import java.util.List;
import java.util.Random;

import static io.flatfeestack.TestHelper.committee;
import static io.flatfeestack.TestHelper.createSignature;
import static io.flatfeestack.TestHelper.defaultAccount;
import static io.flatfeestack.TestHelper.defaultPubKey;
import static io.flatfeestack.TestHelper.fundAccounts;
import static io.flatfeestack.TestHelper.fundContract;
import static io.flatfeestack.TestHelper.getHash160FromPublicKey;
import static io.flatfeestack.TestHelper.getNetworkFee;
import static io.flatfeestack.TestHelper.getRandomHashes;
import static io.flatfeestack.TestHelper.getRandomTeasToPreset;
import static io.flatfeestack.TestHelper.getSum;
import static io.flatfeestack.TestHelper.getSystemFee;
import static io.flatfeestack.TestHelper.getUniformTeas;
import static io.flatfeestack.TestHelper.handleFeeFactors;
import static io.flatfeestack.TestHelper.owner;
import static io.flatfeestack.TestHelper.ownerPubKey;
import static io.flatfeestack.TestHelper.printFees;
import static io.flatfeestack.TestHelper.sendAndWaitUntilTransactionIsExecuted;
import static io.neow3j.contract.Token.toFractions;
import static io.neow3j.transaction.AccountSigner.calledByEntry;
import static io.neow3j.transaction.AccountSigner.global;
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
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

@ContractTest(blockTime = 1, contracts = PayoutNeo.class)
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
    private static Random random;
    private static BigInteger[] ownerIds;
    private static BigInteger presetTea;
    private static BigInteger[] teas;
    private static BigInteger[] presetTeas;

    private static BigInteger contractGasBalanceBeforePayout;

    @RegisterExtension
    private static final ContractTestExtension ext = new ContractTestExtension();

    @BeforeAll
    public static void setUp() throws Throwable {
        neow3j = ext.getNeow3j();
        gasToken = new GasToken(neow3j);
        handleFeeFactors(neow3j, committee, defaultAccount);
        setTestFactors();
        System.out.println("\n##############setup#################");
        System.out.printf("Owner hash:    '%s'\n", owner.getScriptHash());
        System.out.printf("Owner address: '%s'\n", owner.getAddress());
        payoutContract = ext.getDeployedContract(PayoutNeo.class);
        System.out.printf("Payout contract hash: '%s'", payoutContract.getScriptHash());
        fundAccounts(neow3j, gasToken.toFractions(BigDecimal.valueOf(10_000)), defaultAccount);
        fundContract(neow3j, payoutContract, contractFundAmount);
        System.out.println("##############setup#################\n");
    }

    @DeployConfig(PayoutNeo.class)
    public static DeployConfiguration configure() throws Throwable {
        fundAccounts(ext.getNeow3j(), new GasToken(ext.getNeow3j()).toFractions(BigDecimal.valueOf(10_000)), owner);
        DeployConfiguration config = new DeployConfiguration();
        ContractParameter ownerPubKeyParam = publicKey(ownerPubKey);
        config.setDeployParam(ownerPubKeyParam);
        config.setSigner(global(owner));
        return config;
    }

    @BeforeEach
    public void setUpTest() throws IOException {
        devs = getRandomHashes(nrAccounts);
        ownerIds = new BigInteger[nrAccounts];
        random = new Random();
        for (int i = 0; i < nrAccounts; i++) {
            ownerIds[i] = getRandomOwnerId();
        }
        contractGasBalanceBeforePayout = getContractGasBalance();
    }

    private BigInteger getRandomOwnerId() {
        return BigInteger.valueOf(random.nextInt(Integer.MAX_VALUE));
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

    private BigInteger getTea(BigInteger ownerId) throws IOException {
        return payoutContract.callFunctionReturningInt(getTea, integer(ownerId));
    }

    private void setTea(BigInteger ownerId, BigInteger oldTea, BigInteger newTea) throws Throwable {
        Hash256 txHash = payoutContract
                .invokeFunction(setTea, integer(ownerId), integer(oldTea), integer(newTea))
                .signers(calledByEntry(owner))
                .sign()
                .send()
                .getSendRawTransaction()
                .getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
    }

    private void setTeas(BigInteger[] ownerIds, BigInteger[] oldTea, BigInteger[] newTea)
            throws Throwable {

        Hash256 txHash = payoutContract
                .invokeFunction(setTeas, array(ownerIds), array(oldTea), array(newTea))
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
                .sign()
                .send();
        assertTrue(tx.hasError());
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
        BigInteger randomOwnerId = getRandomOwnerId();
        assertThat(getTea(randomOwnerId), is(BigInteger.ZERO));
        setTea(randomOwnerId, BigInteger.ZERO, BigInteger.valueOf(1234890));
        assertThat(getTea(randomOwnerId), is(BigInteger.valueOf(1234890)));
        setTea(randomOwnerId, BigInteger.valueOf(1234890), BigInteger.valueOf(1234891));
        assertThat(getTea(randomOwnerId), is(BigInteger.valueOf(1234891)));
    }

    @Test
    public void testSetTeaInvalidOldTea() throws Throwable {
        BigInteger randomOwnerId = getRandomOwnerId();
        assertThat(getTea(randomOwnerId), is(BigInteger.ZERO));
        boolean failed = false;
        try {
            setTea(randomOwnerId, BigInteger.valueOf(10), BigInteger.valueOf(11));
        } catch (TransactionConfigurationException e) {
            failed = true;
            assertThat(e.getMessage(), containsString("Provided current amount must be equal and new amount must be " +
                    "greater than the stored value"));
        }
        assertTrue(failed);
    }

    @Test
    public void testSetTeaInvalidNewTea() throws Throwable {
        BigInteger randomOwnerId = getRandomOwnerId();
        assertThat(getTea(randomOwnerId), is(BigInteger.ZERO));
        boolean failed = false;
        try {
            setTea(randomOwnerId, BigInteger.ZERO, BigInteger.valueOf(-1));
        } catch (TransactionConfigurationException e) {
            failed = true;
            assertThat(e.getMessage(), containsString("Provided current amount must be equal and new amount must be " +
                    "greater than the stored value"));
        }
        assertTrue(failed);
    }

    @Test
    public void testSetTeas() throws Throwable {
        int nrOwners = 40;
        BigInteger[] ownerIds = new BigInteger[nrOwners];
        BigInteger[] oldTeas = new BigInteger[nrOwners];
        Arrays.fill(oldTeas, BigInteger.ZERO);
        BigInteger[] newTeas = new BigInteger[nrOwners];
        for (int i = 0; i < nrOwners; i++) {
            BigInteger randomOwnerId = getRandomOwnerId();
            ownerIds[i] = randomOwnerId;
            assertThat(getTea(randomOwnerId), is(BigInteger.ZERO));
            newTeas[i] = BigInteger.valueOf((long) (Math.random() * 100000L) + 1);
        }
        setTeas(ownerIds, oldTeas, newTeas);
        for (int i = 0; i < nrOwners; i++) {
            assertThat(getTea(ownerIds[i]), is(newTeas[i]));
        }
    }

    @Test
    public void testSetTeas_withIncorrectOldTea() throws Throwable {
        int nrOwners = 10;
        BigInteger[] ownerIds = new BigInteger[nrOwners];
        BigInteger[] oldTeas = new BigInteger[nrOwners];
        Arrays.fill(oldTeas, BigInteger.ZERO);
        oldTeas[3] = BigInteger.ONE; // Incorrect oldTea
        BigInteger[] newTeas = new BigInteger[nrOwners];
        for (int i = 0; i < nrOwners; i++) {
            BigInteger randomOwnerId = getRandomOwnerId();
            ownerIds[i] = randomOwnerId;
            assertThat(getTea(randomOwnerId), is(BigInteger.ZERO));
            newTeas[i] = BigInteger.valueOf((long) (Math.random() * 100000L) + 1);
        }
        setTeas(ownerIds, oldTeas, newTeas);
        for (int i = 0; i < nrOwners; i++) {
            if (i == 3) {
                // Assert that tea was not updated for this account
                assertThat(getTea(ownerIds[i]), is(BigInteger.ZERO));
            } else {
                assertThat(newTeas[i], is(not(BigInteger.ZERO)));
                assertThat(getTea(ownerIds[i]), is(newTeas[i]));
            }
        }
    }

    @Test
    public void testSetTeas_withLowerNewTea() throws Throwable {
        int nrOwners = 14;
        BigInteger[] ownerIds = new BigInteger[nrOwners];
        BigInteger[] oldTeas = new BigInteger[nrOwners];
        Arrays.fill(oldTeas, BigInteger.ZERO);
        BigInteger[] newTeas = new BigInteger[nrOwners];
        for (int i = 0; i < nrOwners; i++) {
            BigInteger randomOwnerId = getRandomOwnerId();
            ownerIds[i] = randomOwnerId;
            assertThat(getTea(randomOwnerId), is(BigInteger.ZERO));
            newTeas[i] = BigInteger.valueOf((long) (Math.random() * 100000L) + 10);
        }
        BigInteger[] preset = new BigInteger[nrOwners];
        Arrays.fill(preset, BigInteger.valueOf(8));
        setTeas(ownerIds, oldTeas, preset);
        for (int i = 0; i < nrOwners; i++) {
            assertThat(getTea(ownerIds[i]), is(BigInteger.valueOf(8)));
        }
        newTeas[6] = BigInteger.valueOf(5); // Lower than the set tea
        setTeas(ownerIds, preset, newTeas);
        for (int i = 0; i < nrOwners; i++) {
            if (i == 6) {
                // Assert that tea was not updated for this account
                assertThat(getTea(ownerIds[i]), is(BigInteger.valueOf(8)));
            } else {
                assertThat(newTeas[i], is(not(BigInteger.ZERO)));
                assertThat(getTea(ownerIds[i]), is(newTeas[i]));
            }
        }
    }

    @Test
    public void testSetTeas_withIncorrectLengths() throws Throwable {
        int nrOwners = 10;
        BigInteger[] ownerIds = new BigInteger[nrOwners];
        BigInteger[] oldTeas = new BigInteger[nrOwners - 1]; // Incorrect length
        Arrays.fill(oldTeas, BigInteger.ZERO);
        BigInteger[] newTeas = new BigInteger[nrOwners];
        for (int i = 0; i < nrOwners; i++) {
            BigInteger randomOwnerId = getRandomOwnerId();
            ownerIds[i] = randomOwnerId;
            assertThat(getTea(randomOwnerId), is(BigInteger.ZERO));
            newTeas[i] = BigInteger.valueOf((long) (Math.random() * 100000L) + 1);
        }
        boolean tested = false;
        try {
            setTeas(ownerIds, oldTeas, newTeas);
        } catch (TransactionConfigurationException e) {
            assertThat(e.getMessage(), containsString("Parameters must have same length"));
            tested = true;
        }
        assertTrue(tested);
    }

    @Test
    public void testSetTeas_withIncorrectLengths_newTea() throws Throwable {
        int nrOwners = 10;
        BigInteger[] ownerIds = new BigInteger[nrOwners];
        BigInteger[] oldTeas = new BigInteger[nrOwners];
        Arrays.fill(oldTeas, BigInteger.ZERO);
        int newTeasLength = nrOwners - 5;
        BigInteger[] newTeas = new BigInteger[newTeasLength]; // Incorrect length
        for (int i = 0; i < nrOwners; i++) {
            BigInteger randomOwnerId = getRandomOwnerId();
            ownerIds[i] = randomOwnerId;
            assertThat(getTea(randomOwnerId), is(BigInteger.ZERO));
            if (i < newTeasLength) {
                newTeas[i] = BigInteger.valueOf((long) (Math.random() * 100000L) + 1);
            }
        }
        boolean tested = false;
        try {
            setTeas(ownerIds, oldTeas, newTeas);
        } catch (TransactionConfigurationException e) {
            assertThat(e.getMessage(), containsString("Parameters must have same length"));
            tested = true;
        }
        assertTrue(tested);
    }

    // endregion test basic contract methods
    // region test withdraw with signature

    @Test
    public void testWithdrawWithSignature() throws Throwable {
        Account dev = Account.create();
        BigInteger ownerId = getRandomOwnerId();
        fundAccounts(neow3j, devFundAmountFractions, dev);
        BigInteger balanceContractBefore = getContractGasBalance();
        // Dev earned 12 gas for the first time
        BigInteger teaDev = gasToken.toFractions(BigDecimal.valueOf(12));
        // Create a signature
        Sign.SignatureData sigData = createSignature(ownerId, teaDev, owner);
        // Dev invokes withdraw method with signatureData
        Transaction tx =
                payoutContract.invokeFunction(withdraw, hash160(dev), integer(ownerId), integer(teaDev),
                                signature(sigData))
                        .signers(none(dev))
                        .sign();
        sendAndWaitUntilTransactionIsExecuted(tx, neow3j);

        assertThat(getContractGasBalance(), is(balanceContractBefore.subtract(teaDev)));
        BigInteger totalFee = getSystemFee(tx).add(getNetworkFee(tx));
        assertThat(getGasBalance(dev), is(devFundAmountFractions.add(teaDev).subtract(totalFee)));
        printFees("withdraw sig", tx);
    }

    @Test
    public void testWithdrawWithSignatureAgain() throws Throwable {
        BigInteger ownerId = BigInteger.valueOf(random.nextInt(Integer.MAX_VALUE));
        Account dev = Account.create();
        fundAccounts(neow3j, devFundAmountFractions, dev);
        BigInteger balanceContractBefore = getContractGasBalance();
        BigInteger teaDev = gasToken.toFractions(BigDecimal.valueOf(10));
        Sign.SignatureData sigData = createSignature(ownerId, teaDev, owner);
        Transaction tx =
                payoutContract.invokeFunction(withdraw, hash160(dev), integer(ownerId), integer(teaDev),
                                signature(sigData))
                        .signers(none(dev)).sign();
        sendAndWaitUntilTransactionIsExecuted(tx, neow3j);

        assertThat(getContractGasBalance(), is(balanceContractBefore.subtract(teaDev)));
        BigInteger totalFee = getSystemFee(tx).add(getNetworkFee(tx));
        assertThat(getGasBalance(dev), is(devFundAmountFractions.add(teaDev).subtract(totalFee)));
        assertThat(getTea(ownerId), is(teaDev));

        // Test that signature was invalidated and cannot be used again
        boolean tested = false;
        try {
            payoutContract.invokeFunction(withdraw, hash160(dev), integer(ownerId), integer(teaDev), signature(sigData))
                    .signers(none(dev)).sign();
        } catch (TransactionConfigurationException e) {
            assertThat(e.getMessage(), containsString("These funds have already been withdrawn"));
            tested = true;
        }
        assertTrue(tested);
    }

    // endregion withdraw with signature
    // region withdraw with witness

    @Test
    public void testWithdrawWithWitness() throws Throwable {
        BigInteger randomOwnerId = getRandomOwnerId();
        Account dev = Account.create();
        fundAccounts(neow3j, devFundAmountFractions, dev);
        BigInteger balanceContractBefore = getContractGasBalance();
        BigInteger teaDev = gasToken.toFractions(BigDecimal.valueOf(22));

        Transaction unsignedTransaction =
                payoutContract.invokeFunction(withdraw, hash160(dev), integer(randomOwnerId), integer(1))
                        .signers(calledByEntry(owner))
                        .getUnsignedTransaction();
        byte[] unsignedTxBytes = unsignedTransaction.toArray();
        Transaction tx = NeoSerializableInterface.from(unsignedTxBytes, Transaction.class);
        tx.setNeow3j(neow3j);

        Transaction txWithoutWitnesses =
                payoutContract.invokeFunction(withdraw, hash160(dev), integer(randomOwnerId), integer(teaDev))
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
        assertThat(getTea(randomOwnerId), is(teaDev));
        printFees("withdraw witness", tx);
    }

    @Test
    public void testWithdrawWithWitnessAgain() throws Throwable {
        BigInteger randomOwnerId = getRandomOwnerId();
        Account dev = Account.create();
        fundAccounts(neow3j, devFundAmountFractions, dev);
        BigInteger balanceContractBefore = getContractGasBalance();
        BigInteger teaDev = gasToken.toFractions(BigDecimal.valueOf(22));
        Transaction txToBePreSigned = payoutContract.invokeFunction(withdraw, hash160(dev), integer(randomOwnerId),
                        integer(teaDev))
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
        assertThat(getTea(randomOwnerId), is(teaDev));

        // Create transaction and witness from bytes again and try to send it again.
        tx = NeoSerializableInterface.from(preSignedTxBytes, Transaction.class);
        tx.setNeow3j(neow3j);
        devWitness = Witness.create(tx.getHashData(), dev.getECKeyPair());
        tx.addWitness(devWitness);
        ownerWitnessFromBytes = NeoSerializableInterface.from(witnessBytes, Witness.class);
        tx.addWitness(ownerWitnessFromBytes);

        NeoSendRawTransaction rawTx = tx.send();
        assertTrue(rawTx.hasError());
        assertThat(rawTx.getError().getMessage(), containsString("AlreadyExists"));
    }

    // endregion withdraw with witness
    // region test batch payout methods

    @Test
    public void test1_batchPayout() throws Throwable {
        ContractParameter accountsParam = array(devs);
        ContractParameter ownerIdsParam = array(ownerIds);
        ContractParameter teasParam = array(teas);
        Transaction tx = payoutContract.invokeFunction(batchPayout, ownerIdsParam, accountsParam, teasParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(ownerIds[i]), is(teas[i]));
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
        for (BigInteger ownerId : ownerIds) {
            setTea(ownerId, BigInteger.ZERO, presetTea);
        }
        ContractParameter ownerIdsParam = array(ownerIds);
        ContractParameter accountsParam = array(devs);
        ContractParameter teasParam = array(teas);
        Transaction tx = payoutContract.invokeFunction(batchPayout, ownerIdsParam, accountsParam, teasParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(ownerIds[i]), is(teas[i]));
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
            setTea(ownerIds[i], BigInteger.ZERO, presetTeas[i]);
        }
        ContractParameter ownerIdsParam = array(ownerIds);
        ContractParameter accountsParam = array(devs);
        ContractParameter teasParam = array(teas);
        Transaction tx = payoutContract.invokeFunction(batchPayout, ownerIdsParam, accountsParam, teasParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        for (int i = 0; i < nrAccounts; i++) {
            assertThat(getTea(ownerIds[i]), is(teas[i]));
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
        BigInteger[] ownerIds = new BigInteger[2];
        ownerIds[0] = PayoutNeoIntegrationTest.ownerIds[0];
        ownerIds[1] = PayoutNeoIntegrationTest.ownerIds[1];
        BigInteger[] teas = new BigInteger[2];
        teas[0] = contractGasBalance;
        teas[1] = BigInteger.ONE;
        ContractParameter ownerIdsParam = array(ownerIds);
        ContractParameter accountsParam = array(devs);
        ContractParameter teasParam = array(teas);

        TransactionConfigurationException thrown = assertThrows(TransactionConfigurationException.class,
                () -> payoutContract.invokeFunction(batchPayout, ownerIdsParam, accountsParam, teasParam)
                        .signers(calledByEntry(owner))
                        .sign());
        assertThat(thrown.getMessage(), containsString("Transfer was not successful"));
    }

    // endregion test batch payout with insufficient balance

}
