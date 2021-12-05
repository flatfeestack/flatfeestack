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
import java.util.List;

import static io.flatfeestack.EvaluationHelper.committee;
import static io.flatfeestack.EvaluationHelper.compileContract;
import static io.flatfeestack.EvaluationHelper.createMapParam;
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
import static io.neow3j.types.ContractParameter.*;
import static io.neow3j.utils.Await.waitUntilTransactionIsExecuted;
import static io.neow3j.utils.Numeric.toHexString;
import static io.neow3j.utils.Numeric.toHexStringNoPrefix;
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

    private static final BigDecimal contractFundAmount = BigDecimal.valueOf(7000);
    private static final BigInteger devFundAmountFractions = toFractions(BigDecimal.valueOf(100), GasToken.DECIMALS);

    // Methods
    private static final String getOwner = "getOwner";
    private static final String changeOwner = "changeOwner";
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
        compileContract(PayoutNeoForEvaluation.class.getCanonicalName());
        System.out.println("\n##############setup#################");
        System.out.printf("Owner hash:    '%s'\n", owner.getScriptHash());
        System.out.printf("Owner address: '%s'\n", owner.getAddress());
        fundAccounts(neow3j, gasToken.toFractions(BigDecimal.valueOf(10_000)), defaultAccount, owner);
        payoutContract = deployPayoutNeoContract(neow3j, false);
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
        serviceFee = BigInteger.TEN;
        serviceFeeParam = integer(serviceFee);
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

    // endregion helper methods

    @Test
    public void testSig() throws IOException {
        Account dev = Account.fromWIF("L1D7rqnLDXpJJmHa9o6GyMcoKPr1sSVZzHXFa8fk8GuFJ7dQ7c3L");
        System.out.println("dev: " + dev.getScriptHash());
        BigInteger tea = gasToken.toFractions(new BigDecimal("2"));
        System.out.printf("tea: '%s'\n", toHexString(tea.toByteArray()));
        Sign.SignatureData sig = createSignature(dev.getScriptHash(), tea, owner);
        System.out.println("v: " + sig.getV());
        System.out.println("r: " + toHexStringNoPrefix(sig.getR()));
        System.out.println("s: " + toHexStringNoPrefix(sig.getS()));
        System.out.println(toHexString(sig.getConcatenated()));
    }

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
        System.out.println("v" + sigData.getV());
        System.out.println("r" + sigData.getR());
        System.out.println("s" + sigData.getS());
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

    @Test
    public void test4_insufficientBalanceToPayoutAllAccounts_batchPayoutList() throws Throwable {
        BigInteger contractGasBalance = getContractGasBalance();
        Hash160[] devs = new Hash160[2];
        devs[0] = PayoutNeoEvalIntTest.devs[0];
        devs[1] = PayoutNeoEvalIntTest.devs[1];
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

    @Test
    public void test4_insufficientBalanceToPayoutAllAccounts_batchPayoutListWithServiceFee() throws Throwable {
        BigInteger contractGasBalance = getContractGasBalance();
        Hash160[] devs = new Hash160[2];
        devs[0] = PayoutNeoEvalIntTest.devs[0];
        devs[1] = PayoutNeoEvalIntTest.devs[1];
        BigInteger[] teas = new BigInteger[2];
        teas[0] = contractGasBalance.add(serviceFee);
        teas[1] = BigInteger.ONE.add(serviceFee);
        ContractParameter accountsParam = array(devs);
        ContractParameter teasParam = array(teas);

        exceptionRule.expect(TransactionConfigurationException.class);
        exceptionRule.expectMessage("Transfer was not successful");
        payoutContract.invokeFunction(batchPayoutWithServiceFee, accountsParam, teasParam, integer(serviceFee))
                .signers(calledByEntry(owner)).sign();
    }

    // endregion batch payout with service fee
    // region batch payout with two tea lists

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

    @Test
    public void test4_insufficientBalanceToPayoutAllAccounts_batchPayoutListList() throws Throwable {
        BigInteger contractGasBalance = getContractGasBalance();
        Hash160[] devs = new Hash160[2];
        devs[0] = PayoutNeoEvalIntTest.devs[0];
        devs[1] = PayoutNeoEvalIntTest.devs[1];
        BigInteger[] teasForWithdrawal = new BigInteger[2];
        teasForWithdrawal[0] = contractGasBalance;
        teasForWithdrawal[1] = BigInteger.ONE;
        BigInteger[] teasToStore = new BigInteger[2];
        teasToStore[0] = teasForWithdrawal[0].add(serviceFee);
        teasToStore[1] = teasForWithdrawal[1].add(serviceFee);
        ContractParameter accountsParam = array(devs);
        ContractParameter teasToStoreParam = array(teasToStore);
        ContractParameter teasForWithdrawalParam = array(teasForWithdrawal);

        exceptionRule.expect(TransactionConfigurationException.class);
        exceptionRule.expectMessage("Transfer was not successful");
        payoutContract.invokeFunction(batchPayoutWithTeas, accountsParam, teasToStoreParam, teasForWithdrawalParam)
                .signers(calledByEntry(owner)).sign();
    }

    // endregion batch payout with two tea lists
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

    @Test
    public void test4_insufficientBalanceToPayoutAllAccounts_batchPayoutMap() throws Throwable {
        BigInteger contractGasBalance = getContractGasBalance();
        Hash160[] devs = new Hash160[2];
        devs[0] = PayoutNeoEvalIntTest.devs[0];
        devs[1] = PayoutNeoEvalIntTest.devs[1];
        BigInteger[] teas = new BigInteger[2];
        teas[0] = contractGasBalance;
        teas[1] = BigInteger.ONE;

        ContractParameter payoutMapParam = createMapParam(devs, teas);

        exceptionRule.expect(TransactionConfigurationException.class);
        exceptionRule.expectMessage("Transfer was not successful.");
        payoutContract.invokeFunction(batchPayoutWithMap, payoutMapParam)
                .signers(calledByEntry(owner)).sign();
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

    @Test
    public void test4_insufficientBalanceToPayoutAllAccounts_batchPayoutMapServiceFee() throws Throwable {
        BigInteger contractGasBalance = getContractGasBalance();
        Hash160[] devs = new Hash160[2];
        devs[0] = PayoutNeoEvalIntTest.devs[0];
        devs[1] = PayoutNeoEvalIntTest.devs[1];
        BigInteger[] teas = new BigInteger[2];
        teas[0] = contractGasBalance.add(serviceFee);
        teas[1] = BigInteger.ONE.add(serviceFee);

        ContractParameter payoutMapParam = createMapParam(devs, teas);

        exceptionRule.expect(TransactionConfigurationException.class);
        exceptionRule.expectMessage("Transfer was not successful.");
        payoutContract.invokeFunction(batchPayoutWithMapAndServiceFee, payoutMapParam, integer(serviceFee))
                .signers(calledByEntry(owner)).sign();
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

    @Test
    public void test4_insufficientBalanceToPayoutAllAccounts_batchPayoutMapMap() throws Throwable {
        BigInteger contractGasBalance = getContractGasBalance();
        Hash160[] devs = new Hash160[2];
        devs[0] = PayoutNeoEvalIntTest.devs[0];
        devs[1] = PayoutNeoEvalIntTest.devs[1];
        BigInteger[] teasForWithdrawal = new BigInteger[2];
        teasForWithdrawal[0] = contractGasBalance;
        teasForWithdrawal[1] = BigInteger.ONE;
        BigInteger[] teasToStore = new BigInteger[2];
        teasToStore[0] = teasForWithdrawal[0].add(BigInteger.TEN);
        teasToStore[1] = teasForWithdrawal[1].add(BigInteger.TEN);

        ContractParameter mapForWithdrawParam = createMapParam(devs, teasForWithdrawal);
        ContractParameter mapToStoreParam = createMapParam(devs, teasToStore);

        exceptionRule.expect(TransactionConfigurationException.class);
        exceptionRule.expectMessage("Transfer was not successful.");
        payoutContract.invokeFunction(batchPayoutWithDoubleMap, mapToStoreParam, mapForWithdrawParam)
                .signers(calledByEntry(owner)).sign();
    }

    // endregion batch payout with two maps

}
