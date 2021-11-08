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
import java.util.HashMap;

import static io.flatfeestack.EvaluationHelper.EXEC_FEE_FACTION;
import static io.flatfeestack.EvaluationHelper.FEE_PER_BYTE;
import static io.flatfeestack.EvaluationHelper.MAX_ACCOUNTS_BATCH_PAYOUT_LIST;
import static io.flatfeestack.EvaluationHelper.MAX_ACCOUNTS_BATCH_PAYOUT_MAP;
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
import static io.neow3j.contract.Token.toFractions;
import static io.neow3j.transaction.AccountSigner.calledByEntry;
import static io.neow3j.transaction.AccountSigner.none;
import static io.neow3j.types.ContractParameter.array;
import static io.neow3j.types.ContractParameter.hash160;
import static io.neow3j.types.ContractParameter.integer;
import static io.neow3j.types.ContractParameter.map;
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
    private static final boolean batchPayout_map = false;
    private static final boolean batchPayout_map_oneToMaxAccs_32 = true;
    private static final boolean batchPayout_map_oneToMaxAccs_64 = true;

    private static final BigDecimal contractFundAmount = BigDecimal.valueOf(51_000_000);
    private static final BigInteger devFundAmountFractions = toFractions(BigDecimal.valueOf(100), GasToken.DECIMALS);

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

    private static void runWithdrawWithSignature() throws Throwable {
        // Requires the contract to be funded with 51_000_500 Gas [(10_000/2)*(1_000.1) Gas]
        FileWriter w = getResultFileWriter("withdrawWithSignature");
        writeNetworkFactors(w, neow3j);
        w.write("tea,systemfee,networkfee\n");
        for (EvaluationTypeWithdraw type : EvaluationTypeWithdraw.values()) {
            BigInteger tea = type.tea;
            Account dev = Account.create();
            Sign.SignatureData sig = createSignature(dev.getScriptHash(), tea, owner);
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
            w.write(format("%s,%s,%s\n", gasToken.toDecimals(tea),
                    gasToken.toDecimals(BigInteger.valueOf(tx.getSystemFee())),
                    gasToken.toDecimals(BigInteger.valueOf(tx.getNetworkFee()))));
        }
        w.close();
    }

    private static void runWithdrawWithWitness() throws Throwable {
        // Requires the contract to be funded with 50_050_000 Gas [(10_000/2)*(10_001) Gas].
        FileWriter w = getResultFileWriter("withdrawWithWitness");
        writeNetworkFactors(w, neow3j);
        w.write("tea,systemfee,networkfee\n");
        for (EvaluationTypeWithdraw type : EvaluationTypeWithdraw.values()) {
            BigInteger tea = type.tea;
            Account dev = Account.create();
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
            w.write(format("%s,%s,%s\n", gasToken.toDecimals(tea),
                    gasToken.toDecimals(BigInteger.valueOf(tx.getSystemFee())),
                    gasToken.toDecimals(BigInteger.valueOf(tx.getNetworkFee()))));
        }
        w.close();
    }

    // region withdraw sig

    @Test
    public void evaluateWithdrawSignature() throws Throwable {
        assertCorrectNetworkFactors();
        if (evaluateWithdrawSignature) {
            runWithdrawWithSignature();
        }
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

    // endregion withdraw witness
    // region batch list - 0.1 Gas to 1000 Gas

    @Test
    public void testEvaluateBatchPayout_list() throws Throwable {
        assertCorrectNetworkFactors();
        if (batchPayout_list) {
            runBatchPayoutEvaluation_list();
        }
    }

    private static void runBatchPayoutEvaluation_list() throws Throwable {
        FileWriter w = getResultFileWriter("batchPayout_list");
        writeNetworkFactors(w, neow3j);
        w.write("#accounts,tea,systemfee,networkfee\n");
        for (EvaluationTypeList type : EvaluationTypeList.values()) {
            System.out.println(type.name());
            Transaction tx = executeSingleBatchPayout_list(type);
            NeoApplicationLog.Execution log = null;
            try {
                log = tx.getApplicationLog().getExecutions().get(0);
                assertThat(log.getState(), is(NeoVMStateType.HALT));
                assertThat(log.getNotifications(), hasSize(type.nrAccounts.intValue()));
            } catch (Exception e) {
                System.out.println(e.getMessage());
            }
            for (NeoApplicationLog.Execution.Notification notification : log.getNotifications()) {
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

    private static Transaction executeSingleBatchPayout_list(EvaluationTypeList type) throws Throwable {
        BigInteger tea = type.tea;
        BigInteger nrAccounts = type.nrAccounts;
        Hash160[] randomHashes = getRandomHashes(nrAccounts.intValue());
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

    // endregion batch list - 0.1 Gas to 1000 Gas
    // region batch list - 1 to 1012 accounts - pushint32/pushint64

    @Test
    public void evaluateBatchPayout_list_oneToMax_pushint32() throws Throwable {
        assertCorrectNetworkFactors();
        if (batchPayout_list_oneToMaxAccs_32) {
            FileWriter w = getResultFileWriter("batchPayout_list_oneToMax_pushint32");
            runBatchPayoutEvaluation_list_oneToMax(w, gasToken.toFractions(BigDecimal.ONE));
        }
    }

    @Test
    public void evaluateBatchPayout_list_oneToMax_pushint64() throws Throwable {
        if (batchPayout_list_oneToMaxAccs_64) {
            assertCorrectNetworkFactors();
            FileWriter w = getResultFileWriter("batchPayout_list_oneToMax_pushint64");
            runBatchPayoutEvaluation_list_oneToMax(w, gasToken.toFractions(new BigDecimal("25")));
        }
    }

    private static void runBatchPayoutEvaluation_list_oneToMax(FileWriter w, BigInteger tea) throws Throwable {
        writeNetworkFactors(w, neow3j);
        w.write(format("tea=%s\n", gasToken.toDecimals(tea)));
        w.write("#accounts,systemfee,networkfee,totalfee,sysfeeperacc,netfeeperacc,totalfeeperacc\n");
        for (int i = 0; i <= MAX_ACCOUNTS_BATCH_PAYOUT_LIST; i++) {
            Transaction tx = executeSingleBatchPayout_list(i, tea);
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

    private static Transaction executeSingleBatchPayout_list(int nrAccounts, BigInteger tea) throws Throwable {
        Hash160[] randomHashes = getRandomHashes(nrAccounts);
        if (nrAccounts % 10 == 0) {
            System.out.println(nrAccounts + " accounts");
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

    // endregion batch list - 1 to 1012 accounts - pushint32/pushint64
    // region batch map - 0.1 Gas to 1000 Gas

    @Test
    public void testEvaluateBatchPayout_map() throws Throwable {
        assertCorrectNetworkFactors();
        if (batchPayout_map) {
            runBatchPayoutEvaluation_map();
        }
    }

    private static void runBatchPayoutEvaluation_map() throws Throwable {
        FileWriter w = getResultFileWriter("batchPayout_map_v2");
        writeNetworkFactors(w, neow3j);
        w.write("#accounts,tea,systemfee,networkfee\n");
        for (EvaluationTypeMap type : EvaluationTypeMap.values()) {
            Transaction tx = executeSingleBatchPayout_map(type);
            waitUntilTransactionIsExecuted(tx.getTxId(), neow3j);
            NeoApplicationLog log = tx.getApplicationLog();
            NeoApplicationLog.Execution exec = log.getExecutions().get(0);
            assertThat(exec.getState(), is(NeoVMStateType.HALT));
            assertThat(exec.getNotifications(), hasSize(type.nrAccounts.intValue()));
            for (NeoApplicationLog.Execution.Notification not : exec.getNotifications()) {
                assertThat(not.getEventName(), is(TRANSFER_EVENT_NAME));
            }
            w.write(format("%s,%s,%s,%s\n", type.nrAccounts,
                    gasToken.toDecimals(type.tea),
                    gasToken.toDecimals(BigInteger.valueOf(tx.getSystemFee())),
                    gasToken.toDecimals(BigInteger.valueOf(tx.getNetworkFee()))));
        }
        w.close();
    }

    private static Transaction executeSingleBatchPayout_map(EvaluationTypeMap type) throws Throwable {
        BigInteger tea = type.tea;
        BigInteger nrAccounts = type.nrAccounts;
        Hash160[] randomHashes = getRandomHashes(nrAccounts.intValue());
        HashMap<Hash160, BigInteger> hashMap = new HashMap<>();
        for (Hash160 randomHash : randomHashes) {
            hashMap.put(randomHash, tea);
        }
        ContractParameter mapParam = map(hashMap);
        Transaction tx = payoutContract.invokeFunction(batchPayout, mapParam)
                .signers(calledByEntry(owner)).sign();
        NeoSendRawTransaction rawTx = tx.send();
        Hash256 txHash = rawTx.getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
        return tx;
    }

    // endregion batch map - 0.1 Gas to 1000 Gas
    // region batch map - 1 to 674 accounts - pushint32/pushint64

    // region batch list - 1 to 1012 accounts - pushint32/pushint64

    @Test
    public void evaluateBatchPayout_map_oneToMax_pushint32() throws Throwable {
        assertCorrectNetworkFactors();
        if (batchPayout_map_oneToMaxAccs_32) {
            FileWriter w = getResultFileWriter("batchPayout_map_oneToMax_pushint32");
            runBatchPayoutEvaluation_map_oneToMax(w, gasToken.toFractions(BigDecimal.ONE));
        }
    }

    @Test
    public void evaluateBatchPayout_map_oneToMax_pushint64() throws Throwable {
        if (batchPayout_map_oneToMaxAccs_64) {
            assertCorrectNetworkFactors();
            FileWriter w = getResultFileWriter("batchPayout_map_oneToMax_pushint64");
            runBatchPayoutEvaluation_map_oneToMax(w, gasToken.toFractions(new BigDecimal("25")));
        }
    }

    private static void runBatchPayoutEvaluation_map_oneToMax(FileWriter w, BigInteger tea) throws Throwable {
        writeNetworkFactors(w, neow3j);
        w.write(format("tea=%s\n", gasToken.toDecimals(tea)));
        w.write("#accounts,systemfee,networkfee,totalfee,sysfeeperacc,netfeeperacc,totalfeeperacc\n");
        for (int i = 0; i <= MAX_ACCOUNTS_BATCH_PAYOUT_MAP; i++) {
            Transaction tx = executeSingleBatchPayout_map(i, tea);
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

    private static Transaction executeSingleBatchPayout_map(int nrAccounts, BigInteger tea) throws Throwable {
        Hash160[] randomHashes = getRandomHashes(nrAccounts);
        if (nrAccounts % 10 == 0) {
            System.out.println(nrAccounts + " accounts");
        }
        HashMap<Hash160, BigInteger> hashMap = new HashMap<>();
        for (Hash160 randomHash : randomHashes) {
            hashMap.put(randomHash, tea);
        }
        ContractParameter mapParam = map(hashMap);
        Transaction tx = payoutContract.invokeFunction(batchPayout, mapParam)
                .signers(calledByEntry(owner)).sign();
        Hash256 txHash = tx.send().getSendRawTransaction().getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
        return tx;
    }

    // endregion batch map - 1 to 674 accounts - pushint32/pushint64

}
