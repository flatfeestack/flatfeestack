package io.flatfeestack;

import io.neow3j.compiler.CompilationUnit;
import io.neow3j.compiler.Compiler;
import io.neow3j.contract.ContractManagement;
import io.neow3j.contract.GasToken;
import io.neow3j.crypto.ECKeyPair;
import io.neow3j.protocol.Neow3j;
import io.neow3j.protocol.core.response.NeoApplicationLog;
import io.neow3j.protocol.core.response.NeoSendRawTransaction;
import io.neow3j.protocol.http.HttpService;
import io.neow3j.transaction.AccountSigner;
import io.neow3j.transaction.TransactionBuilder;
import io.neow3j.types.Hash160;
import io.neow3j.types.Hash256;
import io.neow3j.types.NeoVMStateType;
import io.neow3j.utils.Await;
import io.neow3j.wallet.Account;

import static io.neow3j.contract.SmartContract.calcContractHash;
import static io.neow3j.transaction.AccountSigner.none;
import static io.neow3j.types.ContractParameter.publicKey;

public class NeoExpressDeployment {

    // The owner of the smart contract.
    static final Account owner = Account.fromWIF("KzrHihgvHGpF9urkSbrbRcgrxSuVhpDWkSfWvSg97pJ5YgbdHKCQ");
    static final ECKeyPair.ECPublicKey ownerPubKey = owner.getECKeyPair().getPublicKey();

    // The node to connect to.
    private static final String NODE = "http://localhost:50012";

    public static void main(String[] args) throws Throwable {
        Neow3j neow3j = Neow3j.build(new HttpService(NODE));

        if (new GasToken(neow3j).getBalanceOf(owner).intValue() == 0) {
            throw new RuntimeException("Alice has no GAS. If you're running a neo express instance run `neoxp " +
                    "transfer 100 GAS genesis alice` in a terminal in the root directory of this project.");
        }
        AccountSigner signer = AccountSigner.none(owner);

        deployPayoutSmartContract(signer, neow3j);
    }

    private static void deployPayoutSmartContract(AccountSigner signer, Neow3j neow3j) throws Throwable {
        // Compile the contract you want to deploy
        CompilationUnit res = new Compiler().compile(PayoutNeo.class.getCanonicalName());

        // Calculate the contract hash based on the deployer account and the contract's Nef file and its manifest.
        Hash160 hash = calcContractHash(owner.getScriptHash(), res.getNefFile().getCheckSumAsInteger(),
                res.getManifest().getName());

        // Call the deploy method with the compiled Nef file and Manifest
        // Additionally, the deploy method requires the owner's public key as data.
        // The owner should sign the transaction and its signature should only be valid in the contract that will be
        // deployed.
        TransactionBuilder builder = new ContractManagement(neow3j)
                .deploy(res.getNefFile(), res.getManifest(), publicKey(ownerPubKey))
                .signers(none(owner).setAllowedContracts(hash));

        // Sign, send and get
        NeoSendRawTransaction response = builder.sign().send();

        if (response.hasError()) {
            throw new RuntimeException("Failed to deploy contract.");
        }

        // Get the returned tx hash and wait until the tx is persisted in a block on-chain.
        Hash256 txHash = response.getResult().getHash();
        System.out.println("Deployment Transaction Hash: " + txHash.toString());
        Await.waitUntilTransactionIsExecuted(txHash, neow3j);

        // Read the application log of the tx and verify that the tx was successful.
        NeoApplicationLog log = neow3j.getApplicationLog(txHash).send().getApplicationLog();
        if (log.getExecutions().get(0).getState().equals(NeoVMStateType.FAULT)) {
            throw new Exception(
                    "Failed to deploy contract. NeoVM error message: " + log.getExecutions().get(0).getException());
        }

        System.out.println("Deployment successful. Contract Hash: " + hash);
    }
}
