package io.flatfeestack;

import io.neow3j.devpack.ByteString;
import io.neow3j.devpack.ECPoint;
import io.neow3j.devpack.Hash160;
import io.neow3j.devpack.List;
import io.neow3j.devpack.Storage;
import io.neow3j.devpack.StorageContext;
import io.neow3j.devpack.StorageMap;
import io.neow3j.devpack.annotations.DisplayName;
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

@Permission(nativeContract = NativeContract.GasToken)
@Permission(nativeContract = NativeContract.CryptoLib)
public class PayoutNeo {

    /**
     * The storage context
     */
    static final StorageContext ctx = Storage.getStorageContext();

    /**
     * The prefix for the contractMap StorageMap
     */
    static final byte[] contractMapPrefix = toByteArray((byte) 0x01);
    /**
     * StorageMap to store contract relevant information
     */
    static final StorageMap contractMap = ctx.createMap(contractMapPrefix);
    /**
     * Key of the contract owner public key in the contractMap.
     * <p>
     * The method withdraw(Hash160, int, String) requires an ECPoint to be stored from the owner.
     * This restricts the owner from being a multi-sig account.
     */
    static final byte[] ownerKey = toByteArray((byte) 0x02);

    /**
     * The prefix for the total earned amount ({@code tea}) StorageMap
     */
    static final byte[] teaMapPrefix = toByteArray((byte) 0x10);
    /**
     * StorageMap to store k-v pairs mapping addresses (key) to their {@code Total Earned Amount}
     */
    static final StorageMap teaMap = ctx.createMap(teaMapPrefix);

    /**
     * Changes the contract owner.
     *
     * @param newOwner The new contract owner.
     */
    public static void changeOwner(ECPoint newOwner) {
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization";
        assert checkWitness(newOwner) : "New owner must witness this change.";
        contractMap.put(ownerKey, newOwner.toByteString());
    }

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
     * Withdraws the earned amount.
     * <p>
     * This solution approach may need a second contract owner address that does not hold any funds.
     * In order to guarantee, that the beneficiary account actually pays for the transaction.
     * Otherwise, the beneficiary could transfer any funds to another address, so that the
     *
     * @param account The beneficiary account.
     * @param tea     The {@code Total Earned Amount} of this account.
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
     * @return a list of all accounts that did not receive any payment.
     */
    public static List<Hash160> batchPayout(Hash160[] accounts, int[] teas, int serviceFee) {
        // Note: int is always handled as BigInteger on NeoVM. -> It does not matter how high the number is.
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization";
        assert accounts.length == teas.length : "The parameters must have the same length.";
        List<Hash160> unsuccessful = new List<>();
        Hash160 contractHash = getExecutingScriptHash();
        boolean transfer;
        Hash160 a;
        int tea;
        for (int i = 0; i < accounts.length; i++) {
            // TODO: 12.10.21 Evaluation -> Is it cheaper to store in local var or read every time used?
            //  list.length and entry from list
            a = accounts[i];
            tea = teas[i];
            int oldTea = teaMap.get(a.toByteString()).toIntOrZero();
            teaMap.put(a.toByteString(), tea + serviceFee);
            int payoutAmount = tea - oldTea - serviceFee;
            if (payoutAmount <= 0) {
                // TODO: 12.10.21 Evaluation -> Check this variation.
                //  AFAIK, this case should only occur, if the dev herself already withdrew. Otherwise, the contract
                //  owner has not calculated the payout correctly and should not have included this account in the
                //  batch payout in the first place.
                //  With the above mentioned, it is not clear who was mistaken.
                teaMap.put(a.toByteString(), oldTea + serviceFee);
                unsuccessful.add(a);
                continue;
            }
            transfer = GasToken.transfer(contractHash, a, payoutAmount, null);
            if (!transfer) {
                // TODO: 12.10.21 Evaluation -> Should the service fee be deducted if the transfer goes wrong?
                teaMap.put(a.toByteString(), oldTea + serviceFee);
                unsuccessful.add(a);
            }
        }
        return unsuccessful;
    }

    /**
     * This method provides an address blacklist functionality.
     * <p>
     * E.g., this may be used in the case a user wants to change her address. In that case, the contract owner
     * can set the {@code Tea} to the highest {@code Tea} of that account for which a signature was provided.
     * The new address can then be initialized with a {@code Tea} that is equal to the current {@code Tea} that is
     * stored off-chain minus the here provided {@code oldTea}.
     * <p>
     * The {@code oldTea} is checked, so that no immediate withdrawal takes place before executing this.
     *
     * @param account The account to set the {@code Total Earned Amount} for.
     * @param oldTea  The previous {@code Total Earned Amount} for that account.
     * @param newTea  The new {@code Total Earned Amount} for that account.
     */
    public static void setTotalEarnedAmount(Hash160 account, int oldTea, int newTea) {
        //assert checkWitness(account) : "No authorization.";
        // If the developer is required to witness this, the method looses its blacklist functionality.
        assert checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization.";
        int alreadyWithdrawn = teaMap.get(account.toByteString()).toIntOrZero();
        assert alreadyWithdrawn != oldTea : "Funds were withdrawn in the meantime.";
        assert newTea < alreadyWithdrawn : "The provided amount is lower than the already withdrawn amount.";
        teaMap.put(account.toByteString(), newTea);
    }

    @Safe
    public static int getTotalEarnedAmount(Hash160 account) {
        return teaMap.get(account.toByteString()).toIntOrZero();
    }

    @DisplayName("onContractFunding")
    private static Event2Args<Hash160, Integer> onContractFunding;

    /**
     * This method is called if the contract is being funded.
     *
     * @param from   The sender.
     * @param amount The amount transferred to this contract.
     * @param data   Arbitrary data.
     */
    @OnNEP17Payment
    public static void onNep17Payment(Hash160 from, int amount, Object data) {
        onContractFunding.fire(from, amount);
    }

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

    public static int length(int[] list) {
        return list.length;
    }

}
