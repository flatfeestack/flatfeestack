package io.flatfeestack;

import io.neow3j.devpack.ByteString;
import io.neow3j.devpack.ECPoint;
import io.neow3j.devpack.Hash160;
import io.neow3j.devpack.Helper;
import io.neow3j.devpack.Storage;
import io.neow3j.devpack.StorageContext;
import io.neow3j.devpack.StorageMap;
import io.neow3j.devpack.annotations.DisplayName;
import io.neow3j.devpack.annotations.ManifestExtra;
import io.neow3j.devpack.annotations.OnDeployment;
import io.neow3j.devpack.annotations.OnNEP17Payment;
import io.neow3j.devpack.annotations.Permission;
import io.neow3j.devpack.annotations.Safe;
import io.neow3j.devpack.constants.NamedCurve;
import io.neow3j.devpack.constants.NativeContract;
import io.neow3j.devpack.contracts.CryptoLib;
import io.neow3j.devpack.contracts.GasToken;
import io.neow3j.devpack.events.Event3Args;

import static io.neow3j.devpack.Helper.concat;
import static io.neow3j.devpack.Helper.toByteArray;
import static io.neow3j.devpack.Runtime.checkWitness;
import static io.neow3j.devpack.Runtime.getExecutingScriptHash;

/**
 * This contract was used for the evaluation.
 */
@Permission(nativeContract = NativeContract.CryptoLib)
@Permission(nativeContract = NativeContract.GasToken)
@ManifestExtra(key = "Author", value = "Michael Bucher")
public class PayoutNeo {

    /**
     * The storage context
     */
    static final StorageContext ctx = Storage.getStorageContext();

    /**
     * The StorageMap to store contract relevant information.
     */
    static final StorageMap contractMap = new StorageMap(ctx, new byte[]{0x01});

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
     * The StorageMap to store k-v pairs mapping addresses as key to their total earned amount (tea) as value.
     */
    static final StorageMap teaMap = new StorageMap(ctx, new byte[]{0x10});

    /**
     * Upon deployment, the initial owner is set.
     *
     * @param data   The initial owner's public key.
     * @param update True, if the contract is being deployed, false if it is updated.
     */
    @OnDeployment
    public static void deploy(Object data, boolean update) throws Exception {
        if (!update) {
            if (!ECPoint.isValid(data) || !checkWitness((ECPoint) data))
                throw new Exception("No authorization");
        }
        contractMap.put(ownerKey, (ByteString) data);
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

    /**
     * Is fired if the tea of an account is changed by the contract owner without a payment.
     * <p>
     * The arguments relate to the account, the tea that was stored before and the new tea.
     */
    @DisplayName("onTeaUpdateWithoutPayment")
    private static Event3Args<Integer, Integer, Integer> onTeaUpdateWithoutPayment;

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
     * @param newOwner the new contract owner.
     */
    public static void changeOwner(ECPoint newOwner) throws Exception {
        if (!checkWitness(new ECPoint(contractMap.get(ownerKey)))) {
            throw new Exception("No authorization");
        }
        if (!checkWitness(newOwner)) {
            throw new Exception("No authorization");
        }
        contractMap.put(ownerKey, newOwner.toByteString());
    }

    // endregion owner
    // region tea

    /**
     * Gets the total earned amount (tea) of an account.
     *
     * @param ownerId ghe owner id.
     * @return the total earned amount.
     */
    @Safe
    public static int getTea(int ownerId) {
        return teaMap.get(ownerId).toIntOrZero();
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
     * @param ownerId the owner id to set the tea for.
     * @param oldTea  the previous tea for that account.
     * @param newTea  the new tea for that account.
     */
    public static void setTea(int ownerId, int oldTea, int newTea) throws Exception {
        if (!checkWitness(new ECPoint(contractMap.get(ownerKey)))) {
            throw new Exception("No authorization");
        }
        int storedTea = teaMap.get(ownerId).toIntOrZero();
        if (oldTea != storedTea || newTea <= storedTea) {
            throw new Exception("Provided current amount must be equal and new amount must be greater than the stored" +
                    " value");
        }
        teaMap.put(ownerId, newTea);
        onTeaUpdateWithoutPayment.fire(ownerId, oldTea, newTea);
    }

    /**
     * This method supports multiple use cases. It may be used as a blacklist functionality, as a simple modifier or
     * in case developers want to change their addresses without withdrawing the earned funds to the current
     * address (e.g. in case of a loss of the private key).
     * <p>
     * For each account, {@code oldTea} is compared with the stored {@code tea}, so that no immediate withdrawal
     * takes place before executing this.
     * <p>
     * In case of an address change, the contract owner can set the {@code tea} to the highest {@code tea} of that
     * account for which a signature was provided, in order to invalidate that signature.
     *
     * @param ownerIds the owner ids to set the tea for.
     * @param oldTeas  the previously stored tea for the accounts.
     * @param newTeas  the new tea for the accounts.
     */
    public static void setTeas(int[] ownerIds, int[] oldTeas, int[] newTeas) throws Exception {
        if (!checkWitness(new ECPoint(contractMap.get(ownerKey)))) {
            throw new Exception("No authorization");
        }
        int len = oldTeas.length;
        if (len != newTeas.length) {
            throw new Exception("Parameters must have same length");
        }
        for (int i = 0; i < len; i++) {
            int ownerId = ownerIds[i];
            int storedTea = teaMap.get(ownerId).toIntOrZero();
            int oldTea = oldTeas[i];
            if (oldTea != storedTea) {
                continue;
            }
            int newTea = newTeas[i];
            if (newTea <= storedTea) {
                continue;
            }
            teaMap.put(ownerId, newTea);
            onTeaUpdateWithoutPayment.fire(ownerId, oldTea, newTea);
        }
    }

    // endregion tea
    // region withdraw

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
     * @param account   the beneficiary account.
     * @param tea       the tea of this account.
     * @param signature the signature
     */
    public static void withdraw(Hash160 account, int ownerId, int tea, ByteString signature) throws Exception {
        if (new CryptoLib().verifyWithECDsa(
                new ByteString(concat(Helper.toByteArray(ownerId), toByteArray(tea))), // the message
                new ECPoint(contractMap.get(ownerKey)), // the contract owner
                signature, // the signature
                NamedCurve.Secp256r1 // the curve
        )) {
            throw new Exception("Provided signature was not valid for the given parameters");
        }
        int amountToWithdraw = tea - teaMap.get(ownerId).toIntOrZero();
        if (amountToWithdraw <= 0) {
            throw new Exception("These funds have already been withdrawn");
        }
        teaMap.put(ownerId, tea);
        if (!new GasToken().transfer(getExecutingScriptHash(), account, amountToWithdraw, null)) {
            throw new Exception("Transfer was not successful");
        }
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
     * @param account the beneficiary account.
     * @param tea     the tea of this account.
     */
    public static void withdraw(Hash160 account, int ownerId, int tea) throws Exception {
        if (!checkWitness(new ECPoint(contractMap.get(ownerKey)))) {
            throw new Exception("No authorization.");
        }
        int amountToWithdraw = tea - teaMap.get(ownerId).toIntOrZero();
        if (amountToWithdraw <= 0) {
            throw new Exception("Amount to withdraw must be positive");
        }
        teaMap.put(ownerId, tea);
        if (!new GasToken().transfer(getExecutingScriptHash(), account, amountToWithdraw, null)) {
            throw new Exception("Transfer was not successful");
        }
    }

    // endregion withdraw
    // region batched payouts

    /**
     * Pays out the earned amount for multiple accounts.
     * <p>
     * Must be invoked by the contract owner.
     * <p>
     * The payment amount for each account is equal to the difference of the provided new tea in the parameter and the
     * currently stored value in the {@code teaMap}. After calculating the payment amount, the value of the
     * account in the {@code teaMap} is updated with the new tea.
     * <p>
     * A potential service fee for each included account can be deducted off-chain by the contract owner when
     * providing the first signature after each batch payout.
     *
     * @param ownerIds the owner ids.
     * @param accounts the accounts to pay out to.
     * @param teas     the corresponding teas.
     */
    public static void batchPayout(int[] ownerIds, Hash160[] accounts, int[] teas) throws Exception {
        if (!checkWitness(new ECPoint(contractMap.get(ownerKey)))) {
            throw new Exception("No authorization.");
        }
        int len = ownerIds.length;
        // Note: If teas had fewer items than accounts, the code would run into out of bounds anyway, but the other
        // way around that is not the case, thus this check is required.
        if (len != accounts.length && len != teas.length) {
            throw new Exception("Parameters must have same length");
        }
        for (int i = 0; i < len; i++) {
            Hash160 acc = accounts[i];
            int ownerId = ownerIds[i];
            int tea = teas[i];
            int payoutAmount = tea - teaMap.get(ownerId).toIntOrZero();
            if (payoutAmount <= 0) {
                continue;
            }
            teaMap.put(ownerId, tea);
            if (!new GasToken().transfer(getExecutingScriptHash(), acc, payoutAmount, null)) {
                throw new Exception("Transfer was not successful");
            }
        }
    }

    // endregion batched payouts

}
