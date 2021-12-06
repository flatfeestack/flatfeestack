package io.flatfeestack;

import io.neow3j.devpack.ByteString;
import io.neow3j.devpack.ECPoint;
import io.neow3j.devpack.Hash160;
import io.neow3j.devpack.Map;
import io.neow3j.devpack.Storage;
import io.neow3j.devpack.StorageContext;
import io.neow3j.devpack.StorageMap;
import io.neow3j.devpack.annotations.ManifestExtra;
import io.neow3j.devpack.annotations.OnDeployment;
import io.neow3j.devpack.annotations.OnNEP17Payment;
import io.neow3j.devpack.annotations.Permission;
import io.neow3j.devpack.annotations.Safe;
import io.neow3j.devpack.constants.NativeContract;
import io.neow3j.devpack.contracts.CryptoLib;
import io.neow3j.devpack.contracts.GasToken;

import static io.neow3j.devpack.Helper.concat;
import static io.neow3j.devpack.Helper.toByteArray;
import static io.neow3j.devpack.Runtime.checkWitness;
import static io.neow3j.devpack.Runtime.getExecutingScriptHash;

/**
 * This contract was used to manually evaluate implementation specific details. The most optimized methods are
 * collected in the final {@link PayoutNeo} contract.
 */
@Permission(nativeContract = NativeContract.CryptoLib)
@Permission(nativeContract = NativeContract.GasToken)
@ManifestExtra(key = "Author", value = "Michael Bucher")
public class PayoutNeoForEvaluation {

    /**
     * The storage context
     */
    static final StorageContext ctx = Storage.getStorageContext();

    /**
     * The StorageMap to store contract relevant information.
     */
    static final StorageMap contractMap = ctx.createMap(new byte[]{0x01});

    /**
     * The key where the contract owner's public key is stored in the contractMap.
     * <p>
     * The method {@code withdraw(Hash160, int, String)} requires an ECPoint of the owner to be stored, since the
     * method {@code verifyWithECDsa} of the native contract {@code CryptoLib} requires the public key of the signer.
     * <p>
     * This restricts the owner from being a multi-sig account.
     */
    static final byte[] ownerKey = toByteArray((byte) 0x02);

    /**
     * StorageMap to store k-v pairs mapping addresses (key) to their total earned amount (tea).
     */
    static final StorageMap teaMap = ctx.createMap(new byte[]{0x10});

    /**
     * Upon deployment, the initial owner is set.
     *
     * @param data   The initial owner's public key.
     * @param update True, if the contract is being deployed, false if it is updated.
     */
    @OnDeployment
    public static void deploy(Object data, boolean update) {
        if (!update) {
            ByteString pubKey = (ByteString) data;
            // ECPoint instantiation checks valid public key length. Thus, makes sure that the data cannot be a Hash160.
            assert checkWitness(new ECPoint(pubKey)) : "Passed public key must match a witness.";
            contractMap.put(ownerKey, pubKey);
        }
    }

    /**
     * This method is called when the contract receives NEP-17 tokens.
     * <p>
     * It is required to receive NEP-17 tokens.
     *
     * @param from   The sender.
     * @param amount The amount transferred to this contract.
     * @param data   Arbitrary data. This field is required by standard and is not used here.
     */
    @OnNEP17Payment
    public static void onNep17Payment(Hash160 from, int amount, Object data) {
    }

    // region owner

    /**
     * Gets the {@code ECPoint} of the owner of this contract.
     *
     * @return the contract owner's {@code ECPoint}.
     */
    @Safe
    public static ECPoint getOwner() {
        return new ECPoint(contractMap.get(ownerKey));
    }

    /**
     * Changes the contract owner.
     *
     * @param newOwner The new contract owner.
     */
    public static void changeOwner(ECPoint newOwner) {
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization";
        assert checkWitness(newOwner) : "The new owner must witness this change.";
        contractMap.put(ownerKey, newOwner.toByteString());
    }

    // endregion owner
    // region tea

    /**
     * Gets the total earned amount (tea) of an account.
     *
     * @param account The account.
     * @return the total earned amount.
     */
    @Safe
    public static int getTea(Hash160 account) {
        return teaMap.get(account.toByteString()).toIntOrZero();
    }

    /**
     * This method supports multiple use cases. It may be used as a blacklist functionality, as a simple modifier or
     * in case a developer wants to change her address without withdrawing the earned funds to the old address (e.g.
     * in case of a loss of the private key).
     * <p>
     * The {@code oldTea} is compared with the stored {@code tea}, so that no immediate withdrawal takes place before
     * executing this.
     * <p>
     * In the case of an address change, the contract owner can set the {@code tea} to the highest {@code tea} of
     * that account for which a signature was provided, in order to invalidate that signature.
     *
     * @param account The account to set the tea for.
     * @param oldTea  The previous tea for that account.
     * @param newTea  The new tea for that account.
     */
    public static void setTea(Hash160 account, int oldTea, int newTea) {
        // Idea: If the developer is required to witness this, the method loses its blacklist functionality.
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization.";
        int storedTea = teaMap.get(account.toByteString()).toIntOrZero();
        assert oldTea == storedTea : "Stored tea is not equal to the provided oldTea.";
        assert newTea > storedTea : "The provided amount is lower than or equal to the stored tea.";
        teaMap.put(account.toByteString(), newTea);
    }

    // endregion tea
    // region withdrawal

    /**
     * Withdraws the earned amount with the option to delegate the payment of the emerging transaction fees.
     * <p>
     * This method uses a signature that is passed as parameter, so that an address different to
     * the beneficiary address may pay for the transaction fees that emerge from this transaction.
     * <p>
     * The signature is created by the owner of this contract and the signed message is
     * the concatenation of the account and the tea.
     * <p>
     * For the use of this method, the owner of the contract is expected to share just the
     * signature data with the beneficiary. The transaction can then be created by the
     * beneficiary using any signer with witness scope {@code none}, that then can sign the
     * transaction and hence pay for the transaction fees.
     * <p>
     * This method requires the contract owner to be a single-sig account, since its public key
     * is required for the verification of the signature.
     * <p>
     * The payment amount is equal to the difference of the provided tea and the currently stored value in the
     * {@code teaMap}. After calculating the payment amount, the value of the account in the {@code teaMap} is
     * updated with the new tea.
     *
     * @param account   The beneficiary account.
     * @param tea       The tea of this account.
     * @param signature The signature
     */
    public static void withdraw(Hash160 account, int tea, ByteString signature) {
        // Verify the signature
        assert CryptoLib.verifyWithECDsa(
                new ByteString(concat(account.toByteArray(), toByteArray(tea))), // the message
                new ECPoint(contractMap.get(ownerKey)), // the contract owner
                signature, // the signature
                (byte) 23 // the curve
        ) : "Signature invalid.";
        // Calculate the amount to withdraw
        int amountToWithdraw = tea - teaMap.get(account.toByteString()).toIntOrZero();
        assert amountToWithdraw > 0 : "These funds have already been withdrawn.";
        teaMap.put(account.toByteString(), tea);
        // Transfer the earned tokens to the account
        assert GasToken.transfer(getExecutingScriptHash(), account, amountToWithdraw, null) :
                "Transfer was not successful.";
    }

    /**
     * Withdraws the earned amount.
     * <p>
     * Requires to be witnessed by the contract owner.
     * <p>
     * The payment amount is equal to the difference of the provided tea and the currently stored value in the
     * {@code teaMap}. After calculating the payment amount, the value of the account in the {@code teaMap} is
     * updated with the new tea.
     *
     * @param account The beneficiary account.
     * @param tea     The tea of this account.
     */
    public static void withdraw(Hash160 account, int tea) {
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization";
        int amountToWithdraw = tea - teaMap.get(account.toByteString()).toIntOrZero();
        assert amountToWithdraw > 0 : "These funds have already been withdrawn.";
        teaMap.put(account.toByteString(), tea);
        assert GasToken.transfer(getExecutingScriptHash(), account, amountToWithdraw, null) : "Transfer was not " +
                "successful.";
    }

    // endregion withdrawal
    // region batch payout

    /**
     * Pays out the earned amount for multiple accounts.
     * <p>
     * Must be invoked by the contract owner.
     * <p>
     * The payment amount for each account is equal to the difference of the provided new tea in the parameter and the
     * currently stored value in the {@code teaMap}. After calculating the payment amount, the value of the account
     * in the {@code teaMap} is updated with the new tea.
     * <p>
     * A potential service fee for each included account can be deducted off-chain by the contract owner when
     * providing the first signature after each batch payout.
     *
     * @param accounts The accounts to pay out to.
     * @param teas     The corresponding teas.
     */
    public static void batchPayout(Hash160[] accounts, int[] teas) {
        // Note: int is always handled as BigInteger on NeoVM. -> It does not matter how high the number is.
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization";
        // Note: Instead of reading the length multiple times, storing its value in a local var is cheaper.
        int len = accounts.length;
        assert len == teas.length : "The parameters must have the same length.";
        // Idea: Return unsuccessful payouts -> This is not necessary, since the GasToken events can be used to track
        // the successful transfers and thus the unsuccessful payouts can be derived from the parameters and those
        // events.
        for (int i = 0; i < len; i++) {
            // Note: Initializing the account hash160 and the tea integer outside the loop does not affect
            // the Gas costs.
            Hash160 acc = accounts[i];
            int tea = teas[i];
            int payoutAmount = tea - teaMap.get(acc.toByteString()).toIntOrZero();
            if (payoutAmount <= 0) {
                // Throwing this even costs 0.04_388_202 Gas (Gas=10USD -> about 1 cent)
                // This case only happens if the dev herself already withdrew or the contract owner did not pass the
                // values correctly.
                continue;
            }
            teaMap.put(acc.toByteString(), tea);
            assert GasToken.transfer(getExecutingScriptHash(), acc, payoutAmount, null) : "Transfer was not " +
                    "successful.";
        }
    }

    /**
     * Pays out the earned amount for multiple accounts.
     * <p>
     * Must be invoked by the contract owner.
     * <p>
     * The payment amount for each account is equal to the difference of the provided new tea in the parameter and the
     * currently stored value in the {@code teaMap} minus the service fee. After calculating the payment amount, the
     * value of the account in the {@code teaMap} is updated with the new tea.
     *
     * @param accounts   The accounts to pay out to.
     * @param teas       The corresponding teas.
     * @param serviceFee The service fee that each developer pays to be included in this batch payout.
     */
    public static void batchPayoutWithServiceFee(Hash160[] accounts, int[] teas, int serviceFee) {
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization";
        int len = accounts.length;
        assert len == teas.length : "The parameters must have the same length.";
        for (int i = 0; i < len; i++) {
            Hash160 acc = accounts[i];
            int tea = teas[i];
            int payoutAmount = tea - teaMap.get(acc.toByteString()).toIntOrZero() - serviceFee;
            if (payoutAmount <= 0) {
                continue;
            }
            teaMap.put(acc.toByteString(), tea);
            assert GasToken.transfer(getExecutingScriptHash(), acc, payoutAmount, null) : "Transfer was not " +
                    "successful.";
        }
    }

    /**
     * Pays out the earned amount for multiple accounts.
     * <p>
     * Must be invoked by the contract owner.
     * <p>
     * The {@code teasToStore} are used to update the {@code teaMap}, while the {@code teasForWithdrawal} are used to
     * calculate the actual payment.
     * <p>
     * The payment amount for each account is equal to the difference of the provided new tea in the
     * {@code teaForWithdrawal} parameter and the currently stored value in the {@code teaMap}.
     *
     * @param accounts          The accounts to pay out to.
     * @param teasToStore       The teas that are used to store in the contract storage.
     * @param teasForWithdrawal The teas that are used for the calculation of the payout amount.
     */
    public static void batchPayoutWithTeas(Hash160[] accounts, int[] teasToStore, int[] teasForWithdrawal) {
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization.";
        int len = accounts.length;
        assert (len == teasToStore.length) &&
                (len == teasForWithdrawal.length) : "The parameters must have the same length.";
        for (int i = 0; i < accounts.length; i++) {
            Hash160 acc = accounts[i];
            int teaForWithdrawal = teasForWithdrawal[i];
            int payoutAmount = teaForWithdrawal - teaMap.get(acc.toByteString()).toIntOrZero();
            if (payoutAmount <= 0) {
                continue;
            }
            teaMap.put(acc.toByteString(), teasToStore[i]);
            assert GasToken.transfer(getExecutingScriptHash(), acc, payoutAmount, null) : "Transfer was not " +
                    "successful.";
        }
    }

    /**
     * Pays out the earned amount for multiple accounts.
     * <p>
     * Must be invoked by the contract owner.
     * <p>
     * The payment amount for each account is equal to the difference of the provided new tea in the parameter and the
     * currently stored value in the {@code teaMap}. After calculating the payment amount, the value of the
     * account in the {@code teaMap} is updated with the new tea.
     *
     * @param payoutMap The accounts and their corresponding teas.
     */
    public static void batchPayoutWithMap(Map<Hash160, Integer> payoutMap) {
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization.";
        for (Hash160 acc : payoutMap.keys()) {
            int tea = payoutMap.get(acc);
            int payoutAmount = tea - teaMap.get(acc.toByteString()).toIntOrZero();
            if (payoutAmount <= 0) {
                continue;
            }
            teaMap.put(acc.toByteString(), tea);
            assert GasToken.transfer(getExecutingScriptHash(), acc, payoutAmount, null) : "Transfer was not " +
                    "successful.";
        }
    }

    /**
     * Pays out the earned amount for multiple accounts.
     * <p>
     * Must be invoked by the contract owner.
     * <p>
     * The payment amount for each account is equal to the difference of the provided new tea in the parameter and the
     * currently stored value in the {@code teaMap} minus the service fee. After calculating the payment amount, the
     * value of the account in the {@code teaMap} is updated with the new tea.
     *
     * @param payoutMap  The accounts and their corresponding teas.
     * @param serviceFee A service fee that is deducted from the payment amount.
     */
    public static void batchPayoutWithMapAndServiceFee(Map<Hash160, Integer> payoutMap, int serviceFee) {
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization.";
        for (Hash160 acc : payoutMap.keys()) {
            int tea = payoutMap.get(acc);
            int payoutAmount = tea - teaMap.get(acc.toByteString()).toIntOrZero() - serviceFee;
            if (payoutAmount <= 0) {
                continue;
            }
            teaMap.put(acc.toByteString(), tea);
            assert GasToken.transfer(getExecutingScriptHash(), acc, payoutAmount, null) : "Transfer was not " +
                    "successful.";
        }
    }

    /**
     * Pays out the earned amount for multiple accounts.
     * <p>
     * Must be invoked by the contract owner.
     * <p>
     * The teas in the {@code storeMap} are used to update the {@code teaMap}, while the teas in the
     * {@code withdrawalMap} are used to calculate the actual payment.
     * <p>
     * The payment amount for each account is equal to the difference of the provided new tea in the
     * {@code withdrawalMap} parameter and the currently stored value in the {@code teaMap}.
     *
     * @param storeMap      The accounts and their corresponding tea that is used to store in the contract storage.
     * @param withdrawalMap The accounts and their corresponding tea that is used for the calculation of the payout
     *                      amount.
     */
    public static void batchPayoutWithDoubleMap(Map<Hash160, Integer> storeMap, Map<Hash160, Integer> withdrawalMap) {
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization.";
        for (Hash160 acc : storeMap.keys()) {
            int teaToStore = storeMap.get(acc);
            int teaForWithdrawal =
                    withdrawalMap.get(acc); // The VM will fault immediately if the key is not present.
            assert teaToStore > teaForWithdrawal : "Tea to store must be greater or equal than tea to withdraw.";
            int payoutAmount = teaForWithdrawal - teaMap.get(acc.toByteString()).toIntOrZero();
            if (payoutAmount <= 0) {
                continue;
            }
            teaMap.put(acc.toByteString(), teaToStore);
            assert GasToken.transfer(getExecutingScriptHash(), acc, payoutAmount, null) : "Transfer was not " +
                    "successful.";
        }
    }

    // endregion batch payout

}
