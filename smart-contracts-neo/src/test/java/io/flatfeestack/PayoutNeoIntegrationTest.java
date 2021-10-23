package io.flatfeestack;

import io.neow3j.compiler.CompilationUnit;
import io.neow3j.compiler.Compiler;
import io.neow3j.contract.ContractManagement;
import io.neow3j.contract.GasToken;
import io.neow3j.contract.NefFile;
import io.neow3j.contract.SmartContract;
import io.neow3j.crypto.ECKeyPair;
import io.neow3j.crypto.Sign;
import io.neow3j.protocol.Neow3j;
import io.neow3j.protocol.core.response.*;
import io.neow3j.protocol.core.stackitem.StackItem;
import io.neow3j.protocol.http.HttpService;
import io.neow3j.serialization.NeoSerializableInterface;
import io.neow3j.transaction.Transaction;
import io.neow3j.transaction.Witness;
import io.neow3j.types.*;
import io.neow3j.wallet.Account;
import io.neow3j.wallet.Wallet;
import org.junit.BeforeClass;
import org.junit.ClassRule;
import org.junit.Test;

import java.io.File;
import java.io.IOException;
import java.math.BigDecimal;
import java.math.BigInteger;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;

import static io.neow3j.contract.ContractUtils.writeContractManifestFile;
import static io.neow3j.contract.ContractUtils.writeNefFile;
import static io.neow3j.contract.SmartContract.calcContractHash;
import static io.neow3j.crypto.Sign.signMessage;
import static io.neow3j.protocol.ObjectMapperFactory.getObjectMapper;
import static io.neow3j.transaction.AccountSigner.calledByEntry;
import static io.neow3j.transaction.AccountSigner.none;
import static io.neow3j.types.ContractParameter.*;
import static io.neow3j.utils.ArrayUtils.*;
import static io.neow3j.utils.Await.waitUntilTransactionIsExecuted;
import static io.neow3j.wallet.Account.createMultiSigAccount;
import static java.util.Arrays.asList;
import static java.util.Collections.singletonList;
import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.Matchers.hasSize;
import static org.hamcrest.core.Is.is;

public class PayoutNeoIntegrationTest {

    private static Neow3j neow3j;
    private static GasToken gasToken;
    private static SmartContract payoutContract;

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
    private static final Account dev1 =
            Account.fromWIF("L1RgqMJEBjdXcuYCMYB6m7viQ9zjkNPjZPAKhhBoXxEsygNXENBb");
    private static final Account dev1Dupl =
            Account.fromWIF("L1RgqMJEBjdXcuYCMYB6m7viQ9zjkNPjZPAKhhBoXxEsygNXENBb");
    private static final Account dev2 =
            Account.fromWIF("Kzkwmjq4aygAHPYwCAhFYwrviar3E5JyiPuNYVcg2Ks88iLm4TmV");
    private static final Account dev2Dupl =
            Account.fromWIF("Kzkwmjq4aygAHPYwCAhFYwrviar3E5JyiPuNYVcg2Ks88iLm4TmV");
    private static final Account dev3 =
            Account.fromWIF("KzTJz7cKJM4dZDeFJroPPK2buag3nA1gWpJtLvoxuEcQUyC4hbzp");

    private static final Wallet committeeWallet = Wallet.withAccounts(committee, defaultAccount);
    private static final Wallet ownerWallet = Wallet.withAccounts(owner, dev1Dupl, dev2Dupl);

    private static final Wallet dev1Wallet = Wallet.withAccounts(dev1);
    private static final Wallet dev2Wallet = Wallet.withAccounts(dev2);

    // Methods
    private static final String batchPayout = "batchPayout";
    private static final String changeOwner = "changeOwner";
    private static final String getOwner = "getOwner";
    private static final String withdraw = "withdraw";
    private static final String getTea = "getTea";

    @ClassRule
    public static NeoTestContainer neoTestContainer = new NeoTestContainer();

    @BeforeClass
    public static void setUp() throws Throwable {
        neow3j = Neow3j.build(new HttpService(neoTestContainer.getNodeUrl()));
        compileContract(PayoutNeoForEvaluation.class.getCanonicalName());
        gasToken = new GasToken(neow3j);
        System.out.println("Owner hash:   " + owner.getScriptHash());
        System.out.println("Owner address:" + owner.getAddress());
        System.out.println("Dev1 hash:    " + dev1.getScriptHash());
        System.out.println("Dev1 address: " + dev1.getAddress());
        fundAccounts(gasToken.toFractions(BigDecimal.valueOf(10_000)), defaultAccount, owner);
        fundAccounts(gasToken.toFractions(BigDecimal.valueOf(10)), dev1, dev2);
        payoutContract = deployPayoutNeoContract();
        System.out.println("Payout contract hash: " + payoutContract.getScriptHash());
        fundPreSignContract();
    }

    private Sign.SignatureData createSignature(Hash160 account, BigInteger tea, Account signer) {
        byte[] accountArray = account.toLittleEndianArray();
        byte[] teaArray = reverseArray(tea.toByteArray());
        byte[] message = concatenate(accountArray, teaArray);
        return signMessage(message, signer.getECKeyPair());
    }

    // Helper methods

    private static void compileContract(String canonicalName) throws IOException {
        CompilationUnit res = new Compiler().compile(canonicalName);

        // Write contract (compiled, NEF) to the disk
        Path buildNeow3jPath = Paths.get("build", "neow3j");
        buildNeow3jPath.toFile().mkdirs();
        writeNefFile(res.getNefFile(), res.getManifest().getName(), buildNeow3jPath);

        // Write manifest to the disk
        writeContractManifestFile(res.getManifest(), buildNeow3jPath);
    }

    private static void fundPreSignContract() throws Throwable {
        BigInteger balanceOf = gasToken.getBalanceOf(payoutContract.getScriptHash());
        BigInteger fundAmount = gasToken.toFractions(BigDecimal.valueOf(7000));
        Hash256 txHash = gasToken
                .transfer(owner, payoutContract.getScriptHash(), fundAmount)
                .sign()
                .send()
                .getSendRawTransaction()
                .getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);
        NeoApplicationLog log = neow3j.getApplicationLog(txHash).send().getApplicationLog();
        System.out.println(log.getExecutions().get(0).getNotifications().get(0).getEventName());
    }

    private static void fundAccounts(BigInteger gasFractions, Account... accounts) throws Throwable {
        BigInteger minAmount = gasToken.toFractions(new BigDecimal("500"));
        BigInteger committeeBalance = gasToken.getBalanceOf(committee);
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

    private BigInteger getContractBalance() throws IOException {
        return getBalance(payoutContract.getScriptHash());
    }

    private BigInteger getBalance(Hash160 account) throws IOException {
        return gasToken.getBalanceOf(account);
    }

//    private Hash256 withdrawWithWitness(Hash160 dev, BigInteger teaFractions) {
//    }

    @Test
    public void testVerifySig() throws IOException {
        BigInteger balanceContract = getContractBalance();
        // Dev1 earned 12 gas for the first time
        BigInteger teaDev1 = gasToken.toFractions(BigDecimal.valueOf(12));
        byte[] message = concatenate(dev1.getScriptHash().toLittleEndianArray(),
                reverseArray(teaDev1.toByteArray()));

        Sign.SignatureData signatureData = signMessage(message, owner.getECKeyPair());
        InvocationResult invocationResult = payoutContract.callInvokeFunction("verifySig",
                        asList(hash160(dev1), integer(teaDev1), signature(signatureData)), calledByEntry(owner)) //
                // returns false
//                        asList(byteArray(message), signature(signatureData)), calledByEntry(owner)) // returns true
                .getInvocationResult();
        System.out.println("#################");
        System.out.println(invocationResult.getStack().get(0).getBoolean());
        System.out.println("#################");
    }

    @Test
    public void testFundContract() throws Throwable {
        BigInteger contractBalance = gasToken.getBalanceOf(payoutContract.getScriptHash());
        BigInteger fundAmount = gasToken.toFractions(BigDecimal.valueOf(1500L));
        Hash256 txHash = gasToken.transfer(owner, payoutContract.getScriptHash(), fundAmount)
                .signers(calledByEntry(owner))
                .sign()
                .send()
                .getSendRawTransaction()
                .getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        BigInteger balanceAfterTransfer = gasToken.getBalanceOf(payoutContract.getScriptHash());
        assertThat(balanceAfterTransfer, is(contractBalance.add(fundAmount)));
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
    public void testWithdrawWithSignature() throws Throwable {
        BigInteger balanceContractBefore = getContractBalance();
        BigInteger balanceDev1Before = getBalance(dev1.getScriptHash());
        BigInteger teaStored = payoutContract.callFuncReturningInt(getTea, hash160(dev1));
        // Dev1 earned 12 gas for the first time
        BigInteger teaDev1 = gasToken.toFractions(BigDecimal.valueOf(12));
        BigInteger expectedPayoutAmount = teaDev1.subtract(teaStored);

        // Create a signature
        Sign.SignatureData signatureData = createSignature(dev1.getScriptHash(), teaDev1, owner);

        // Dev1 invokes withdraw method with signatureData
        Transaction tx = payoutContract.invokeFunction(
                        withdraw,
                        hash160(dev1), integer(teaDev1), signature(signatureData))
                .signers(none(dev1))
                .sign();
        Hash256 txHash = tx.send()
                .getSendRawTransaction()
                .getHash();
        waitUntilTransactionIsExecuted(txHash, neow3j);

        BigInteger balanceContractAfter = getContractBalance();
        assertThat(balanceContractAfter, is(balanceContractBefore.subtract(expectedPayoutAmount)));

        BigInteger balanceDev1After = getBalance(dev1.getScriptHash());
        BigInteger networkFee = BigInteger.valueOf(tx.getNetworkFee());
        BigInteger systemFee = new BigInteger(
                neow3j.getApplicationLog(txHash).send().getApplicationLog()
                        .getExecutions().get(0).getGasConsumed());
        BigInteger totalFee = systemFee.add(networkFee);
        System.out.println("'withdraw' system fee (gasconsumed): " + systemFee);
        System.out.println("'withdraw' network fee:              " + networkFee);
        System.out.println("'withdraw' total fee:                " + systemFee.add(networkFee));
        assertThat(balanceDev1After, is(balanceDev1Before.add(expectedPayoutAmount).subtract(totalFee)));
    }

    @Test
    public void testWithdrawWithWitness() throws Throwable {
        // check if second contract owner is needed that does not hold any funds
        // to prevent paying for the transaction.
        // Todo: Test, if the second signer automatically pays for the fees, if the first one can not cover them.

        BigInteger initialBalanceDev1 = gasToken.getBalanceOf(dev1);
        BigInteger initialBalanceOwner = gasToken.getBalanceOf(owner);
        BigInteger teaDev1 = gasToken.toFractions(BigDecimal.valueOf(2.2));

        // Create the pre-signed transaction
        Transaction txToBePreSignedByOwner = payoutContract.invokeFunction(withdraw, hash160(dev1), integer(teaDev1))
                .signers(none(dev1), calledByEntry(owner).setAllowedContracts(payoutContract.getScriptHash()))
                .getUnsignedTransaction();

        Witness ownerWitness =
                Witness.create(txToBePreSignedByOwner.getHashData(), owner.getECKeyPair());
        byte[] witnessBytes = ownerWitness.toArray();
        byte[] preSignedTxBytes = txToBePreSignedByOwner.toArray();

        Transaction tx = NeoSerializableInterface.from(preSignedTxBytes, Transaction.class);
        tx.setNeow3j(neow3j);
        Witness dev1Witness = Witness.create(tx.getHashData(), dev1.getECKeyPair());
        tx.addWitness(dev1Witness);
        Witness ownerWitnessFromBytes = NeoSerializableInterface.from(witnessBytes, Witness.class);
        tx.addWitness(ownerWitnessFromBytes);

        NeoSendRawTransaction result = tx.send();
        Hash256 txHash = result.getSendRawTransaction().getHash();

        waitUntilTransactionIsExecuted(txHash, neow3j);

        BigInteger networkFee = BigInteger.valueOf(tx.getNetworkFee());
        BigInteger systemFee = BigInteger.valueOf(tx.getSystemFee());
        BigInteger totalFee = networkFee.add(systemFee);

        BigInteger dev1FinalBalance = gasToken.getBalanceOf(dev1);
        assertThat(dev1FinalBalance, is(initialBalanceDev1.add(teaDev1).subtract(totalFee)));

        StackItem teaOnContract = payoutContract.callInvokeFunction(getTea, asList(hash160(dev1)))
                .getInvocationResult().getStack().get(0);
        assertThat(teaOnContract.getType(), is(StackItemType.INTEGER));
        assertThat(teaOnContract.getInteger(), is(teaDev1));
    }

    @Test
    public void testBatchPayout() {
    }

}
