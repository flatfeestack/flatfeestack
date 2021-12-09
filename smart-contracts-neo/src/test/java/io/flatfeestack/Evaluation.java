package io.flatfeestack;

import io.neow3j.contract.GasToken;
import io.neow3j.contract.PolicyContract;
import io.neow3j.contract.SmartContract;
import io.neow3j.crypto.Sign;
import io.neow3j.protocol.Neow3j;
import io.neow3j.protocol.core.response.NeoApplicationLog;
import io.neow3j.protocol.core.response.NeoSendRawTransaction;
import io.neow3j.protocol.http.HttpService;
import io.neow3j.serialization.NeoSerializableInterface;
import io.neow3j.transaction.Transaction;
import io.neow3j.transaction.Witness;
import io.neow3j.types.ContractParameter;
import io.neow3j.types.Hash160;
import io.neow3j.types.Hash256;
import io.neow3j.types.NeoVMStateType;
import io.neow3j.wallet.Account;
import org.junit.BeforeClass;
import org.junit.ClassRule;
import org.junit.Test;

import java.io.FileWriter;
import java.io.IOException;
import java.math.BigDecimal;
import java.math.BigInteger;
import java.util.Arrays;

import static io.flatfeestack.EvaluationHelper.EXEC_FEE_FACTION;
import static io.flatfeestack.EvaluationHelper.FEE_PER_BYTE;
import static io.flatfeestack.EvaluationHelper.MAX_ACCOUNTS_BATCH_PAYOUT_LIST;
import static io.flatfeestack.EvaluationHelper.STORAGE_PRICE;
import static io.flatfeestack.EvaluationHelper.committee;
import static io.flatfeestack.EvaluationHelper.compileContract;
import static io.flatfeestack.EvaluationHelper.createSignature;
import static io.flatfeestack.EvaluationHelper.defaultAccount;
import static io.flatfeestack.EvaluationHelper.deployPayoutNeoContract;
import static io.flatfeestack.EvaluationHelper.feePayAccount;
import static io.flatfeestack.EvaluationHelper.fundAccounts;
import static io.flatfeestack.EvaluationHelper.fundContract;
import static io.flatfeestack.EvaluationHelper.getRandomHashes;
import static io.flatfeestack.EvaluationHelper.getResultFileWriter;
import static io.flatfeestack.EvaluationHelper.handleFeeFactors;
import static io.flatfeestack.EvaluationHelper.owner;
import static io.flatfeestack.EvaluationHelper.writeFees;
import static io.neow3j.transaction.AccountSigner.calledByEntry;
import static io.neow3j.transaction.AccountSigner.none;
import static io.neow3j.types.ContractParameter.array;
import static io.neow3j.types.ContractParameter.hash160;
import static io.neow3j.types.ContractParameter.integer;
import static io.neow3j.types.ContractParameter.signature;
import static io.neow3j.utils.Await.waitUntilTransactionIsExecuted;
import static java.lang.String.format;
import static org.hamcrest.CoreMatchers.is;
import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.Matchers.hasSize;

public class Evaluation {

    private static Neow3j neow3j;
    private static GasToken gasToken;
    private static SmartContract payoutContract;

    private static final boolean evaluateWithdrawSignature = false;
    private static final boolean evaluateWithdrawWitness = false;

    private static final boolean batchPayout_list = false;
    private static final boolean batchPayout_list_oneToMaxAccs_32 = false;
    private static final boolean batchPayout_list_oneToMaxAccs_64 = false;
    private static final boolean preset_batchPayout_list = false;
    private static final boolean preset_batchPayout_list_oneToMaxAccs_32 = false;
    private static final boolean preset_batchPayout_list_oneToMaxAccs_64 = false;

    private static final BigDecimal presetTea = new BigDecimal("0.01");

    private static final BigDecimal contractFundAmount = BigDecimal.valueOf(51_000_000);

    private static final String TRANSFER_EVENT_NAME = "Transfer";

    // Methods
    private static final String withdraw = "withdraw";
    private static final String batchPayout = "batchPayout";

    @ClassRule
    public static NeoTestContainer neoTestContainer = new NeoTestContainer();

    @BeforeClass
    public static void setUp() throws Throwable {
        neow3j = Neow3j.build(new HttpService(neoTestContainer.getNodeUrl()));
        gasToken = new GasToken(neow3j);
        compileContract(PayoutNeo.class.getCanonicalName());
        System.out.println("\n##############setup#################");
        System.out.printf("Owner hash:    '%s'\n", owner.getScriptHash());
        System.out.printf("Owner address: '%s'\n", owner.getAddress());
        fundAccounts(neow3j, gasToken.toFractions(BigDecimal.valueOf(10_000)), defaultAccount, owner, feePayAccount);
        payoutContract = deployPayoutNeoContract(neow3j, true);
        System.out.printf("Payout contract hash: '%s'\n", payoutContract.getScriptHash());
        fundContract(neow3j, payoutContract, contractFundAmount);
        handleFeeFactors(neow3j, committee, defaultAccount);
        System.out.println("##############setup#################\n");
    }

    private static void writeNetworkFactors(FileWriter w, Neow3j neow3j) throws IOException {
        PolicyContract policyContract = new PolicyContract(neow3j);
        w.write(format("feeperbyte=%s\n", policyContract.getFeePerByte()));
        w.write(format("storageprice=%s\n", policyContract.getStoragePrice()));
        w.write(format("executionfeefactor=%s\n", policyContract.getExecFeeFactor()));
    }

    private void assertCorrectNetworkFactors() throws IOException {
        assertThat(new PolicyContract(neow3j).getFeePerByte(), is(FEE_PER_BYTE));
        assertThat(new PolicyContract(neow3j).getStoragePrice(), is(STORAGE_PRICE));
        assertThat(new PolicyContract(neow3j).getExecFeeFactor(), is(EXEC_FEE_FACTION));
    }

    private static void setTea(Hash160 dev) throws Throwable {
        setTea(dev, gasToken.toFractions(presetTea));
    }

    private static void setTea(Hash160 dev, BigInteger tea) throws Throwable {
        Transaction tx = payoutContract.invokeFunction("setTea", hash160(dev), integer(BigInteger.ZERO),
                        integer(tea))
                .signers(none(feePayAccount), calledByEntry(owner))
                .sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
        assertTeaEquals(dev, tea);
    }

    private static void setTeas(Hash160[] devs) throws Throwable {
        BigInteger teaToSet = gasToken.toFractions(presetTea);
        if (devs.length == 1) {
            setTea(devs[0]);
            assertTeaEquals(devs[0], teaToSet);
            return;
        }
        Hash160[] devs1 = Arrays.copyOfRange(devs, 0, devs.length / 2);
        BigInteger[] oldTeas1 = new BigInteger[devs1.length];
        Arrays.fill(oldTeas1, BigInteger.ZERO);
        BigInteger[] newTeas1 = new BigInteger[devs1.length];
        Arrays.fill(newTeas1, teaToSet);
        Hash160[] devs2 = Arrays.copyOfRange(devs, devs.length / 2, devs.length);
        BigInteger[] oldTeas2 = new BigInteger[devs2.length];
        Arrays.fill(oldTeas2, BigInteger.ZERO);
        BigInteger[] newTeas2 = new BigInteger[devs2.length];
        Arrays.fill(newTeas2, teaToSet);
        Transaction tx = payoutContract.invokeFunction("setTeas", array(devs1), array(oldTeas1), array(newTeas1))
                .signers(none(feePayAccount), calledByEntry(owner))
                .sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
        tx = payoutContract.invokeFunction("setTeas", array(devs2), array(oldTeas2), array(newTeas2))
                .signers(none(feePayAccount), calledByEntry(owner))
                .sign();
        txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
        for (Hash160 dev : devs) {
            assertTeaEquals(dev, teaToSet);
        }
    }

    private static void assertTeaEquals(Hash160 dev, BigInteger tea) throws IOException {
        assertThat(payoutContract.callFuncReturningInt("getTea", hash160(dev)), is(tea));
    }

    // region withdraw sig

    @Test
    public void evaluateWithdrawSignature() throws Throwable {
        assertCorrectNetworkFactors();
        if (evaluateWithdrawSignature) {
            runWithdrawWithSignature();
        }
    }

    private static void runWithdrawWithSignature() throws Throwable {
        FileWriter w = getResultFileWriter("withdraw-signature");
        w.write(format("tea=%s\n", gasToken.toDecimals(gasToken.toFractions(presetTea))));
        writeNetworkFactors(w, neow3j);
        w.write("teatype,presettype,systemfee,networkfee,totalfee\n");
        for (EvaluationTypeWithdraw type : EvaluationTypeWithdraw.values()) {
            BigInteger tea = type.tea;
            BigInteger presetTea = type.presetTea;
            Hash160 dev = Account.create().getScriptHash();
            if (type.hasPresetTea()) {
                setTea(dev, presetTea);
            }
            Sign.SignatureData sig = createSignature(dev, tea, owner);
            Transaction tx = payoutContract.invokeFunction(withdraw, hash160(dev), integer(tea), signature(sig))
                    .signers(none(feePayAccount))
                    .sign();
            tx.send();
            waitUntilTransactionIsExecuted(tx.getTxId(), neow3j);

            // Make sure, the transaction was successful and really included the transfer
            NeoApplicationLog.Execution log = tx.getApplicationLog().getExecutions().get(0);
            assertThat(log.getState(), is(NeoVMStateType.HALT));
            assertThat(log.getNotifications().get(0).getContract(), is(GasToken.SCRIPT_HASH));
            assertThat(log.getNotifications().get(0).getEventName(), is(TRANSFER_EVENT_NAME));

            BigDecimal systemFee = gasToken.toDecimals(BigInteger.valueOf(tx.getSystemFee()));
            BigDecimal networkFee = gasToken.toDecimals(BigInteger.valueOf(tx.getNetworkFee()));
            BigDecimal totalFee = systemFee.add(networkFee);
            w.write(format("%s,%s,%s,%s,%s\n", type.getTeaType(), type.getPresetTeaType(), systemFee, networkFee,
                    totalFee));
        }
        w.close();
    }

    // endregion withdraw sig
    // region withdraw witness

    @Test
    public void evaluateWithdrawWitness() throws Throwable {
        assertCorrectNetworkFactors();
        if (evaluateWithdrawWitness) {
            runWithdrawWithWitness();
        }
    }

    private static void runWithdrawWithWitness() throws Throwable {
        FileWriter w = getResultFileWriter("withdraw-witness");
        writeNetworkFactors(w, neow3j);
        w.write("teatype,presettype,systemfee,networkfee,totalfee\n");
        for (EvaluationTypeWithdraw type : EvaluationTypeWithdraw.values()) {
            BigInteger tea = type.tea;
            BigInteger presetTea = type.presetTea;
            Hash160 dev = Account.create().getScriptHash();
            if (type.hasPresetTea()) {
                setTea(dev, presetTea);
            }
            Transaction txWithoutWitnesses = payoutContract.invokeFunction(withdraw, hash160(dev), integer(tea))
                    .signers(none(feePayAccount),
                            calledByEntry(owner).setAllowedContracts(payoutContract.getScriptHash()))
                    .getUnsignedTransaction();

            Witness ownerWitness = Witness.create(txWithoutWitnesses.getHashData(), owner.getECKeyPair());
            byte[] witnessBytes = ownerWitness.toArray();
            byte[] txWithoutWitnessesBytes = txWithoutWitnesses.toArray();

            // The following steps are done by the dev after receiving the transaction and witness bytes.
            Transaction tx = NeoSerializableInterface.from(txWithoutWitnessesBytes, Transaction.class);
            tx.setNeow3j(neow3j);
            Witness feePayWitness = Witness.create(tx.getHashData(), feePayAccount.getECKeyPair());
            tx.addWitness(feePayWitness);
            Witness ownerWitnessFromBytes = NeoSerializableInterface.from(witnessBytes, Witness.class);
            tx.addWitness(ownerWitnessFromBytes);
            tx.send();
            waitUntilTransactionIsExecuted(tx.getTxId(), neow3j);

            // Make sure, the transaction was successful and really included the transfer
            NeoApplicationLog.Execution log = tx.getApplicationLog().getExecutions().get(0);
            assertThat(log.getState(), is(NeoVMStateType.HALT));
            assertThat(log.getNotifications().get(0).getContract(), is(GasToken.SCRIPT_HASH));
            assertThat(log.getNotifications().get(0).getEventName(), is(TRANSFER_EVENT_NAME));
            BigDecimal systemFee = gasToken.toDecimals(BigInteger.valueOf(tx.getSystemFee()));
            BigDecimal networkFee = gasToken.toDecimals(BigInteger.valueOf(tx.getNetworkFee()));
            BigDecimal totalFee = systemFee.add(networkFee);
            w.write(format("%s,%s,%s,%s,%s\n", type.getTeaType(), type.getPresetTeaType(), systemFee, networkFee,
                    totalFee));
        }
        w.close();
    }

    // endregion withdraw witness
    // region batchPayout - EvaluationTypeList

    @Test
    public void evaluateBatchPayout_list() throws Throwable {
        assertCorrectNetworkFactors();
        if (batchPayout_list) {
            runBatchPayoutEvaluation_list(false);
        }
    }

    @Test
    public void preset_evaluateBatchPayout_list() throws Throwable {
        assertCorrectNetworkFactors();
        if (preset_batchPayout_list) {
            runBatchPayoutEvaluation_list(true);
        }
    }

    private static void runBatchPayoutEvaluation_list(boolean preset) throws Throwable {
        FileWriter w;
        if (preset) {
            w = getResultFileWriter("preset_batchPayout_list");
            w.write(format("preset_tea=%s\n", gasToken.toDecimals(gasToken.toFractions(presetTea))));
        } else {
            w = getResultFileWriter("batchPayout_list");
        }
        writeNetworkFactors(w, neow3j);
        w.write("#accounts,tea,systemfee,networkfee\n");
        for (EvaluationTypeBatchPayout type : EvaluationTypeBatchPayout.values()) {
            Transaction tx = executeSingleBatchPayout_list(type, preset);
            NeoApplicationLog.Execution exec = tx.getApplicationLog().getExecutions().get(0);
            assertThat(exec.getState(), is(NeoVMStateType.HALT));
            assertThat(exec.getNotifications(), hasSize(type.nrAccounts.intValue()));
            for (NeoApplicationLog.Execution.Notification notification : exec.getNotifications()) {
                assertThat(notification.getContract(), is(GasToken.SCRIPT_HASH));
                assertThat(notification.getEventName(), is(TRANSFER_EVENT_NAME));
            }
            w.write(format("%s,%s,%s,%s\n", type.nrAccounts,
                    gasToken.toDecimals(type.tea),
                    gasToken.toDecimals(BigInteger.valueOf(tx.getSystemFee())),
                    gasToken.toDecimals(BigInteger.valueOf(tx.getNetworkFee()))));
        }
        w.close();
    }

    private static Transaction executeSingleBatchPayout_list(EvaluationTypeBatchPayout type, boolean preset) throws Throwable {
        BigInteger tea = type.tea;
        BigInteger nrAccounts = type.nrAccounts;
        Hash160[] randomHashes = getRandomHashes(nrAccounts.intValue());
        if (preset) {
            setTeas(randomHashes);
        }
        ContractParameter accountsParam = array(randomHashes);
        BigInteger[] teas = new BigInteger[nrAccounts.intValue()];
        Arrays.fill(teas, tea);
        ContractParameter teasParam = array(teas);
        Transaction tx = payoutContract.invokeFunction(batchPayout, accountsParam, teasParam)
                .signers(calledByEntry(owner)).sign();
        NeoSendRawTransaction rawTx = tx.send();
        Hash256 txHash = rawTx.getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
        return tx;
    }

    // endregion batch list - EvaluationTypeList
    // region batchPayout - 1 to 1012 accounts - pushint32/pushint64

    @Test
    public void evaluateBatchPayout_list_oneToMax_pushint32() throws Throwable {
        assertCorrectNetworkFactors();
        if (batchPayout_list_oneToMaxAccs_32) {
            FileWriter w = getResultFileWriter("batchPayout_list_oneToMax_pushint32");
            runBatchPayoutEvaluation_list_oneToMax(w, gasToken.toFractions(BigDecimal.ONE), false);
        }
    }

    @Test
    public void preset_evaluateBatchPayout_list_oneToMax_pushint32() throws Throwable {
        assertCorrectNetworkFactors();
        if (preset_batchPayout_list_oneToMaxAccs_32) {
            FileWriter w = getResultFileWriter("preset_batchPayout_list_oneToMax_pushint32");
//            w.write(format("preset_tea=%s\n", gasToken.toDecimals(gasToken.toFractions(presetTea))));
            runBatchPayoutEvaluation_list_oneToMax(w, gasToken.toFractions(BigDecimal.ONE), true);
        }
    }

    @Test
    public void evaluateBatchPayout_list_oneToMax_pushint64() throws Throwable {
        if (batchPayout_list_oneToMaxAccs_64) {
            assertCorrectNetworkFactors();
            FileWriter w = getResultFileWriter("batchPayout_list_oneToMax_pushint64");
            runBatchPayoutEvaluation_list_oneToMax(w, gasToken.toFractions(new BigDecimal("25")), false);
        }
    }

    @Test
    public void preset_evaluateBatchPayout_list_oneToMax_pushint64() throws Throwable {
        if (preset_batchPayout_list_oneToMaxAccs_64) {
            assertCorrectNetworkFactors();
            FileWriter w = getResultFileWriter("preset_batchPayout_list_oneToMax_pushint64");
            w.write(format("preset_tea=%s\n", gasToken.toDecimals(gasToken.toFractions(presetTea))));
            runBatchPayoutEvaluation_list_oneToMax(w, gasToken.toFractions(new BigDecimal("25")), false);
        }
    }

    private static void runBatchPayoutEvaluation_list_oneToMax(FileWriter w, BigInteger tea, boolean preset)
            throws Throwable {
        writeNetworkFactors(w, neow3j);
        w.write(format("tea=%s\n", gasToken.toDecimals(tea)));
        w.write("#accounts,systemfee,networkfee,totalfee,sysfeeperacc,netfeeperacc,totalfeeperacc\n");
        for (int i = 1; i <= MAX_ACCOUNTS_BATCH_PAYOUT_LIST; i++) {
            Transaction tx = executeSingleBatchPayout_list(i, tea, preset);
            waitUntilTransactionIsExecuted(tx.getTxId(), neow3j);
            NeoApplicationLog.Execution log = tx.getApplicationLog().getExecutions().get(0);
            assertThat(log.getState(), is(NeoVMStateType.HALT));
            assertThat(log.getNotifications(), hasSize(i));
            for (NeoApplicationLog.Execution.Notification notification : log.getNotifications()) {
                assertThat(notification.getContract(), is(GasToken.SCRIPT_HASH));
                assertThat(notification.getEventName(), is(TRANSFER_EVENT_NAME));
            }
            writeFees(w, gasToken, i, tx.getSystemFee(), tx.getNetworkFee());
        }
        w.close();
    }

    private static Transaction executeSingleBatchPayout_list(int nrAccounts, BigInteger tea, boolean preset) throws Throwable {
        Hash160[] randomHashes = getRandomHashes(nrAccounts);
        if (nrAccounts % 10 == 0) {
            System.out.println(nrAccounts + " accounts");
        }
        if (preset) {
            setTeas(randomHashes);
        }
        ContractParameter accsParam = array(randomHashes);
        BigInteger[] teas = new BigInteger[nrAccounts];
        Arrays.fill(teas, tea);
        ContractParameter teasParam = array(teas);
        Transaction tx = payoutContract.invokeFunction(batchPayout, accsParam, teasParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
        return tx;
    }

    // endregion batchPayout - 1 to 1012 accounts - pushint32/pushint64

}
