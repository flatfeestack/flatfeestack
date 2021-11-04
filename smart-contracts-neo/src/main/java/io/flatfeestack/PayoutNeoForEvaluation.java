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
import io.neow3j.devpack.events.Event1Arg;

import static io.neow3j.devpack.Helper.concat;
import static io.neow3j.devpack.Helper.toByteArray;
import static io.neow3j.devpack.Runtime.checkWitness;
import static io.neow3j.devpack.Runtime.getExecutingScriptHash;

@Permission(nativeContract = NativeContract.CryptoLib)
@Permission(nativeContract = NativeContract.GasToken)
@ManifestExtra(key = "Author", value = "Michael Bucher")
public class PayoutNeoForEvaluation {

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

    // region tested

    /**
     * Upon deployment, the initial owner is set.
     *
     * @param data   The initial owner's public key.
     * @param update True, if the contract is being deployed, false if it is updated.
     */
    @OnDeployment
    public static void deploy(Object data, boolean update) {
        if (!update) {
            ECPoint initialOwner = (ECPoint) data;
            contractMap.put(ownerKey, initialOwner.toByteString());
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

    @DisplayName("onUnsuccessfulPayout")
    private static Event1Arg<Hash160> onUnsuccessfulPayout;

    // region owner

    /**
     * Get the {@code ECPoint} of the owner of this contract.
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
     * The {@code oldTea} is checked, so that no immediate withdrawal takes place before executing this.
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
        // The message that was signed by the owner and resulted in the provided signature.
        ByteString message = new ByteString(concat(account.toByteArray(), toByteArray(tea)));
        // Verify the signature
        boolean verified = CryptoLib.verifyWithECDsa(message, getOwner(), signature, (byte) 23);
        assert verified : "Signature invalid.";
        // Get the stored tea
        int storedTea = teaMap.get(account.toByteString()).toIntOrZero();
        // Calculate the amount to withdraw
        int amountToWithdraw = tea - storedTea;
        assert amountToWithdraw > 0 : "These funds have already been withdrawn.";
        // Update the withdrawalMap with the new tea
        teaMap.put(account.toByteString(), tea);
        // Transfer the earned tokens to the account
        boolean transfer = GasToken.transfer(getExecutingScriptHash(), account, amountToWithdraw, null);
        assert transfer : "Transfer was not successful.";
    }

    /**
     * Withdraws the earned amount.
     *
     * @param account The beneficiary account.
     * @param tea     The total earned amount of this account.
     */
    public static void withdraw(Hash160 account, int tea) {
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization";
        int storedTea = teaMap.get(account.toByteString()).toIntOrZero();
        int amountToWithdraw = tea - storedTea;
        assert amountToWithdraw > 0 : "These funds have already been withdrawn.";
        teaMap.put(account.toByteString(), tea);
        boolean transfer = GasToken.transfer(getExecutingScriptHash(), account, amountToWithdraw, null);
        assert transfer : "Transfer was not successful.";
    }

    // endregion withdrawal
    // region batch payout

    // TODO: 20.10.21 Evaluation -> find cheapest batchPayout method and use that to evaluate scenarios.

    // service fee is deducted off-chain - as soon as it is deducted, the account is added to upcoming batchPayout list.
    //
    // add case with service fee -> if user wants to be included without further expected increase of her tea.
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
            // Note: Initializing the account hash160, the tea and oldTea integer outside the loop does not affect
            // the Gas costs.
            Hash160 acc = accounts[i];
            int tea = teas[i];
            int oldTea = teaMap.get(acc.toByteString()).toIntOrZero();
            int payoutAmount = tea - oldTea;
            if (payoutAmount <= 0) {
                // Throwing this even costs 0.04_388_202 Gas (Gas=10USD -> about 1 cent)
                // This case only happens if the dev herself already withdrew or the contract owner did not pass the
                // values correctly.
                onUnsuccessfulPayout.fire(acc);
                continue;
            }
            boolean transfer = GasToken.transfer(getExecutingScriptHash(), acc, payoutAmount, null);
            if (!transfer) {
                // This can only be reached if contract funds are too low.
                onUnsuccessfulPayout.fire(acc);
                continue;
            }
            // The contract cannot be drained, since only the contract owner can call this method.
            teaMap.put(acc.toByteString(), tea);
        }
    }

    // endregion tested

    /**
     * Withdraws the earned amount for multiple accounts.
     * <p>
     * Must be invoked by the contract owner.
     * <p>
     * The service fee is deducted off-chain by the contract owner, when providing the first signature after each
     * batch payout.
     * <p>
     * The pre-signatures that are provided for accounts that are included in this batch payout should include the
     * amount being the {@code tea} minus the {@code serviceFee}.
     *
     * @param accounts   The accounts to pay out to.
     * @param teas       The corresponding {@code Total Earned Amount}s.
     * @param serviceFee The service fee that each developer pays to be included in this batch payout.
     */
    public static void batchPayoutWithServiceFee(Hash160[] accounts, int[] teas, int serviceFee) {
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization";
        int len = accounts.length;
        assert len == teas.length : "The parameters must have the same length.";
        for (int i = 0; i < len; i++) {
            Hash160 acc = accounts[i];
            int tea = teas[i];
            int oldTea = teaMap.get(acc.toByteString()).toIntOrZero();
            int payoutAmount = tea - oldTea - serviceFee;
            if (payoutAmount <= 0) {
                // TODO: 12.10.21 Evaluation -> Check this variation.
                //  AFAIK, this case should only occur, if the dev herself already withdrew. Otherwise, the contract
                //  owner has not calculated the payout correctly and should not have included this account in the
                //  batch payout in the first place.
                //  With the above mentioned, it is not clear who was mistaken.
//                teaMap.put(acc.toByteString(), oldTea + serviceFee);
                onUnsuccessfulPayout.fire(acc);
                continue;
            }
            boolean transfer = GasToken.transfer(getExecutingScriptHash(), acc, payoutAmount, null);
            if (transfer) {
                teaMap.put(acc.toByteString(), tea);
            } else {
                // This can only be reached, if contract funds are too low.
                teaMap.put(acc.toByteString(), oldTea);
                onUnsuccessfulPayout.fire(acc);
            }
        }
    }

    /**
     * Same as method above, but allows for individual service fee.
     *
     * @param accounts          The accounts to pay out to.
     * @param teas              The total earned amounts that are used to store in the contract storage.
     * @param teasForWithdrawal The total earned amounts that are used for the calculation of the payout amount.
     * @return a list of all accounts that did not receive any payment.
     */
    public static void batchPayoutWithTeas(Hash160[] accounts, int[] teas, int[] teasForWithdrawal) {
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization.";
        assert (accounts.length == teas.length) && (accounts.length == teasForWithdrawal.length) : "The parameters " +
                "must have the same length.";
        boolean transfer;
        Hash160 a;
        int teaForWithdrawal;
        int payoutAmount;
        for (int i = 0; i < accounts.length; i++) {
            a = accounts[i];
            teaForWithdrawal = teasForWithdrawal[i];
            int oldTea = teaMap.get(a.toByteString()).toIntOrZero();
            payoutAmount = teaForWithdrawal - oldTea;
            if (payoutAmount <= 0) {
                // TODO: 20.10.21 Evaluation -> Should the dev be charged a fee if this is reached?
                onUnsuccessfulPayout.fire(a);
                continue;
            }
            teaMap.put(a.toByteString(), teas[i]);
            transfer = GasToken.transfer(getExecutingScriptHash(), a, payoutAmount, null);
            if (!transfer) {
                teaMap.put(a.toByteString(), oldTea);
                onUnsuccessfulPayout.fire(a);
            }
        }
    }

    public static void batchPayoutWithMap(Map<Hash160, Integer> payoutMap) {
        int payoutAmount;
        boolean transfer;
        for (Hash160 a : payoutMap.keys()) {
            int tea = payoutMap.get(a);
            int oldTea = teaMap.get(a.toByteString()).toIntOrZero();
            payoutAmount = tea - oldTea;
            if (payoutAmount <= 0) {
                onUnsuccessfulPayout.fire(a);
                continue;
            }
            teaMap.put(a.toByteString(), tea);
            transfer = GasToken.transfer(getExecutingScriptHash(), a, payoutAmount, null);
            if (!transfer) {
                // This can only be reached, if contract funds are too low.
                teaMap.put(a.toByteString(), oldTea);
                onUnsuccessfulPayout.fire(a);
            }
        }
    }

    public static void batchPayoutWithMapAndServiceFee(Map<Hash160, Integer> payoutMap, int serviceFee) {
        boolean transfer;
        for (Hash160 a : payoutMap.keys()) {
            int tea = payoutMap.get(a);
            int oldTea = teaMap.get(a.toByteString()).toIntOrZero();
            int payoutAmount = tea - oldTea - serviceFee;
            if (payoutAmount <= 0) {
                onUnsuccessfulPayout.fire(a);
                continue;
            }
            teaMap.put(a.toByteString(), tea);
            transfer = GasToken.transfer(getExecutingScriptHash(), a, payoutAmount, null);
            if (!transfer) {
                // This can only be reached, if contract funds are too low.
                teaMap.put(a.toByteString(), oldTea);
                onUnsuccessfulPayout.fire(a);
            }
        }

    }

    public static void batchPayoutWithDoubleMap(Map<Hash160, Integer> payoutMap,
            Map<Hash160, Integer> teasForWithdrawal) {
    }

    // endregion batch payout

    // region helper methods for evaluation

    // Helper methods for development process

    // 1. get the script hash's byte array in little endian
    // 2. get the integer's byte array
    // 3. reverse the integer's byte array
    // 4. concatenate the little endian script hash's byte array with the reversed byte array of the integer amount
    // 5. Sign this concatenation
    public static boolean verifySig(Hash160 account, int tea, ByteString signature) {
        ByteString message = new ByteString(concat(account.toByteArray(), toByteArray(tea)));
        return CryptoLib.verifyWithECDsa(message, getOwner(), signature, (byte) 0x17);
    }

    public static byte[] concatAccInt(Hash160 a, int i) {
        return concat(a.toByteArray(), toByteArray(i));
    }

    // endregion helper methods for evaluation

}
