package io.flatfeestack;

import io.neow3j.devpack.ByteString;
import io.neow3j.devpack.ECPoint;
import io.neow3j.devpack.Hash160;
import io.neow3j.devpack.Map;
import io.neow3j.devpack.Storage;
import io.neow3j.devpack.StorageContext;
import io.neow3j.devpack.StorageMap;
import io.neow3j.devpack.annotations.DisplayName;
import io.neow3j.devpack.annotations.ManifestExtra;
import io.neow3j.devpack.annotations.OnDeployment;
import io.neow3j.devpack.annotations.OnNEP17Payment;
import io.neow3j.devpack.annotations.Permission;
import io.neow3j.devpack.annotations.Safe;
import io.neow3j.devpack.constants.NativeContract;
import io.neow3j.devpack.contracts.CryptoLib;
import io.neow3j.devpack.contracts.GasToken;
import io.neow3j.devpack.events.Event2Args;

import static io.neow3j.devpack.Helper.concat;
import static io.neow3j.devpack.Helper.toByteArray;
import static io.neow3j.devpack.Runtime.checkWitness;
import static io.neow3j.devpack.Runtime.getExecutingScriptHash;

@Permission(nativeContract = NativeContract.CryptoLib)
@Permission(nativeContract = NativeContract.GasToken)
@ManifestExtra(key = "Author", value = "Michael Bucher")
public class PayoutNeo {

    /**
     * The storage context
     */
    static final StorageContext ctx = Storage.getStorageContext();
    /**
     * StorageMap to store contract relevant information
     */
    static final StorageMap contractMap = ctx.createMap(new byte[]{0x01});
    /**
     * Key of the contract owner public key in the contractMap.
     * <p>
     * The method {@code withdraw(Hash160, int, String)} requires an ECPoint of the owner to be stored, since the
     * method {@code verifyWithECDsa} of the native contract {@code CryptoLib} requires the public key.
     * <p>
     * This restricts the owner from being a multi-sig account.
     */
    static final byte[] ownerKey = toByteArray((byte) 0x02);

    /**
     * StorageMap to store k-v pairs mapping addresses (key) to their {@code Total Earned Amount}
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
            contractMap.put(ownerKey, ((ECPoint) data).toByteString());
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

    /**
     * Is fired if a payout was not successful.
     * <p>
     * The arguments relate to the account that should have been paid and the reason why it was not successful.
     */
    @DisplayName("onUnsuccessfulPayout")
    private static Event2Args<Hash160, String> onUnsuccessfulPayout;

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
    public static void setOwner(ECPoint newOwner) {
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization";
        assert checkWitness(newOwner) : "The new owner must witness this change.";
        contractMap.put(ownerKey, newOwner.toByteString());
    }

    // endregion owner
    // region tea

    /**
     * Gets the total earned amount ({@code tea}) of an account.
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
     * in case a developer may want to change her address without withdrawing the earned funds to the old address (e.g.
     * in case of a loss of the private key).
     * <p>
     * The {@code oldTea} is compared with the stored {@code tea}, so that no immediate withdrawal takes place before
     * executing this.
     * <p>
     * In the case of an address change, the contract owner can set the {@code Tea} to the highest {@code Tea} of
     * that account for which a signature was provided. The new address can then be initialized with a {@code Tea}
     * that is equal to the current {@code Tea} that is stored off-chain minus the here provided {@code oldTea}.
     *
     * @param account The account to set the {@code Total Earned Amount} for.
     * @param oldTea  The previous {@code Total Earned Amount} for that account.
     * @param newTea  The new {@code Total Earned Amount} for that account.
     */
    public static void setTea(Hash160 account, int oldTea, int newTea) {
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization.";
        int storedTea = teaMap.get(account.toByteString()).toIntOrZero();
        assert oldTea == storedTea : "Stored tea is not equal to the provided oldTea.";
        assert newTea > storedTea : "The provided amount is lower than or equal to the stored tea.";
        teaMap.put(account.toByteString(), newTea);
    }

    // endregion tea
    // region withdraw

    /**
     * Withdraws the earned amount with the option to delegate the payment of the emerging
     * transaction fees.
     * <p>
     * This method uses a signature that is passed as parameter, so that an address different to
     * the beneficiary address may pay for the transaction fees emerging of this withdrawal.
     * <p>
     * The signature is created by the owner of this contract and the signed message is
     * the concatenation of the account and the {@code Total Earned Amount}.
     * <p>
     * For the use of this method, the owner of the contract is expected to share just the
     * signature data with the beneficiary. The transaction can then be created by the
     * beneficiary using any signer with witness scope {@code none}, that then can sign the
     * transaction and hence pay for the transaction fees.
     * <p>
     * This method requires the contract owner to be a single-sig account, since its public key
     * is required for the verification of the signature.
     *
     * @param account   The beneficiary account.
     * @param tea       The {@code Total Earned Amount} of this account.
     * @param signature The signature
     */
    public static void withdraw(Hash160 account, int tea, ByteString signature) {
        assert CryptoLib.verifyWithECDsa(
                new ByteString(concat(account.toByteArray(), toByteArray(tea))), // the message
                new ECPoint(contractMap.get(ownerKey)), // the contract owner
                signature, // the signature
                (byte) 23 // the curve
        ) : "Signature invalid.";
        int amountToWithdraw = tea - teaMap.get(account.toByteString()).toIntOrZero();
        assert amountToWithdraw > 0 : "These funds have already been withdrawn.";
        assert GasToken.transfer(getExecutingScriptHash(), account, amountToWithdraw, null) :
                "Transfer was not successful.";
        teaMap.put(account.toByteString(), tea);
    }

    /**
     * Withdraws the earned amount.
     *
     * @param account The beneficiary account.
     * @param tea     The total earned amount of this account.
     */
    public static void withdraw(Hash160 account, int tea) {
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization";
        int amountToWithdraw = tea - teaMap.get(account.toByteString()).toIntOrZero();
        assert amountToWithdraw > 0 : "These funds have already been withdrawn.";
        assert GasToken.transfer(getExecutingScriptHash(), account, amountToWithdraw, null) :
                "Transfer was not successful.";
        teaMap.put(account.toByteString(), tea);
    }

    // endregion withdraw
    // region batched payouts

    public static void batchPayout(Hash160[] accounts, int[] teas) {
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization";
        int len = accounts.length;
        // Note: If teas had fewer items than accounts, the code would run into out of bounds anyways, but the other
        // way around that is not the case, thus this check is required.
        assert len == teas.length : "The parameters must have the same length.";
        for (int i = 0; i < len; i++) {
            Hash160 acc = accounts[i];
            int tea = teas[i];
            int payoutAmount = tea - teaMap.get(acc.toByteString()).toIntOrZero();
            if (payoutAmount <= 0) {
                onUnsuccessfulPayout.fire(acc, "The payout amount is lower or equal to 0.");
                continue;
            }
            if (!GasToken.transfer(getExecutingScriptHash(), acc, payoutAmount, null)) {
                // This can only be reached if contract funds are too low.
                onUnsuccessfulPayout.fire(acc, "The transfer was not successful.");
                continue;
            }
            teaMap.put(acc.toByteString(), tea);
        }
    }

    public static void batchPayout(Map<Hash160, Integer> payoutMap) {
        for (Hash160 acc : payoutMap.keys()) {
            int tea = payoutMap.get(acc);
            int payoutAmount = tea - teaMap.get(acc.toByteString()).toIntOrZero();
            if (payoutAmount <= 0) {
                onUnsuccessfulPayout.fire(acc, "The payout amount is lower or equal to 0.");
                continue;
            }
            if (!GasToken.transfer(getExecutingScriptHash(), acc, payoutAmount, null)) {
                onUnsuccessfulPayout.fire(acc, "The transfer was not successful.");
                continue;
            }
            teaMap.put(acc.toByteString(), tea);
        }
    }

    // endregion batched payouts

}
