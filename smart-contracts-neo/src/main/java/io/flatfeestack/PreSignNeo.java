package io.flatfeestack;

import io.neow3j.devpack.ByteString;
import io.neow3j.devpack.ECPoint;
import io.neow3j.devpack.Hash160;
import io.neow3j.devpack.List;
import io.neow3j.devpack.Runtime;
import io.neow3j.devpack.Storage;
import io.neow3j.devpack.StorageContext;
import io.neow3j.devpack.StorageMap;
import io.neow3j.devpack.annotations.DisplayName;
import io.neow3j.devpack.annotations.OnDeployment;
import io.neow3j.devpack.annotations.OnNEP17Payment;
import io.neow3j.devpack.annotations.Permission;
import io.neow3j.devpack.annotations.Safe;
import io.neow3j.devpack.contracts.CryptoLib;
import io.neow3j.devpack.contracts.GasToken;
import io.neow3j.devpack.events.Event2Args;

import static io.neow3j.devpack.Helper.concat;
import static io.neow3j.devpack.Helper.toByteArray;

@Permission(contract = "0xd2a4cff31913016155e38e474a2c06d08be276cf") // GasToken
@Permission(contract = "0x726cb6e0cd8628a1350a611384688911ab75f51b") // CryptoLib
public class PreSignNeo {

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
     * Key to the curve used in the signature verification
     */
    static final byte[] curveKey = toByteArray((byte) 0x03);

    /**
     * The prefix for the withdrawalMap StorageMap
     */
    static final byte[] withdrawalMapPrefix = toByteArray((byte) 0x10);
    /**
     * StorageMap to store k-v pairs mapping addresses (key) to their {@code Total Earned Amount}
     */
    static final StorageMap withdrawalMap = ctx.createMap(withdrawalMapPrefix);

    /**
     * Changes the contract owner.
     *
     * @param newOwner The new contract owner.
     * @throws Exception if the current owner did not witness this transaction.
     */
    public static void changeOwner(ECPoint newOwner) throws Exception {
        if (!Runtime.checkWitness(new ECPoint(contractMap.get(ownerKey)))) {
            throw new Exception("No authorization.");
        }
        assert Runtime.checkWitness(new ECPoint(contractMap.get(ownerKey))) : "No authorization";
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
     * @throws Exception if the owner of the contract did not witness the transaction or no funds
     *                   were earned.
     */
    public static void withdraw(Hash160 account, int tea) throws Exception {
        if (!Runtime.checkWitness(new ECPoint(contractMap.get(ownerKey)))) {
            throw new Exception("No authorization.");
        }
        ByteString balance = withdrawalMap.get(account.toByteString());
        boolean transfer;
        if (balance == null) {
            withdrawalMap.put(account.toByteString(), tea);
            transfer = GasToken.transfer(Runtime.getExecutingScriptHash(), account,
                    tea, null);
        } else {
            int alreadyWithdrawn = balance.toInt();
            int amountToWithdraw = tea - alreadyWithdrawn;
            if (amountToWithdraw <= 0) {
                throw new Exception("These funds have already been withdrawn.");
            }
            transfer = GasToken.transfer(Runtime.getExecutingScriptHash(), account,
                    amountToWithdraw, null);
        }
        if (!transfer) {
            throw new Exception("Transfer was not successful.");
        }
        onWithdrawal.fire(account, tea);
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
     * @throws Exception if the provided signature is invalid, the funds have already been
     *                   withdrawn, or the transfer was not successful.
     */
    public static void withdraw(Hash160 account, int tea, String signature) throws Exception {

        // The message that was signed by the owner and resulted in the provided signature.
        byte[] messageByteArray = concat(account.toByteArray(), toByteArray(tea));
        ByteString message = new ByteString(messageByteArray);

        // The curve that was used for the signature.
        int curve = contractMap.getInteger(curveKey);

        boolean verified = CryptoLib.verifyWithECDsa(message, getOwner(), signature, (byte) curve);
        if (!verified) {
            throw new Exception("Signature invalid.");
        }

        ByteString withdrawn = withdrawalMap.get(account.toByteString());
        boolean transfer;
        Hash160 executingScriptHash = Runtime.getExecutingScriptHash(); // This contract's script hash
        if (withdrawn == null) {
            withdrawalMap.put(account.toByteString(), tea);
            transfer = GasToken.transfer(executingScriptHash, account, tea, null);

        } else {
            int withdrawnInt = withdrawn.toInt();
            int amountToWithdraw = tea - withdrawnInt;
            if (amountToWithdraw <= 0) {
                throw new Exception("These funds have already been withdrawn.");
            }
            transfer = GasToken.transfer(executingScriptHash, account, amountToWithdraw, null);
        }
        if (!transfer) {
            throw new Exception("Transfer was not successful.");
        }
        onWithdrawal.fire(account, tea);
    }

    // return list of all unsuccessful transfers - check if transfer 0 returns true
    public static void batchWithdraw(Hash160[] accounts, int[] teas, int[] totalAmountForPayout) throws Exception {
        if (!Runtime.checkWitness(new ECPoint(contractMap.get(ownerKey)))) {
            throw new Exception("No authorization.");
        }
        int nrAccounts = accounts.length;
        if (nrAccounts != teas.length && nrAccounts != totalAmountForPayout.length) {
            throw new Exception("The parameters must have the same length.");
        }
        // withdrawal loop
        // return list of hash160 or map... whatever is cheaper and serves the case.
    }

    /**
     * Withdraws the earned amount for multiple accounts.
     * <p>
     * Must be invoked by the contract owner.
     *
     * @param accounts The accounts to pay out to.
     * @param teas     The corresponding {@code Total Earned Amount}s.
     * @return a list of all accounts that did not receive any payment.
     * @throws Exception if the transaction was not witnessed by the contract owner or the
     *                   parameters are not of equal length.
     */
    public static List<Hash160> batchWithdraw(Hash160[] accounts, int[] teas)
    // or batchWithdraw(accounts, teas, tea_withDeductedFee)
    // ask claude if int is always 256 or if it is converted to byte[] and the size of this is used.
            throws Exception {

        if (!Runtime.checkWitness(new ECPoint(contractMap.get(ownerKey)))) {
            throw new Exception("No authorization.");
        }
        int nrAccounts = accounts.length;
        if (nrAccounts != teas.length) {
            throw new Exception("The parameters must have the same length.");
        }
        List<Hash160> unsuccessful = null;
        Hash160 contractHash = Runtime.getExecutingScriptHash();
        boolean transfer;
        Hash160 a;
        int tea;
        for (int i = 0; i < nrAccounts; i++) {
            // is it cheaper to store in local var or read every time used?
            a = accounts[i];
            tea = teas[i];
            ByteString withdrawn = withdrawalMap.get(a.toByteString());
            int toWithdraw;
            if (withdrawn == null) {
                toWithdraw = tea;
            } else {
                int withdrawnInt = withdrawn.toInt();
                if (withdrawnInt >= tea) {
                    unsuccessful.add(a);
                    continue;
                } else {
                    toWithdraw = tea - withdrawn.toInt();
                }
            }
            transfer = GasToken.transfer(contractHash, a, toWithdraw, null);
            if (transfer) {
                withdrawalMap.put(a.toByteString(), toWithdraw);
            } else {
                unsuccessful.add(a);
            }
        }
        return unsuccessful;
    }

    // Not sure, whether this event provides any useful function
    @DisplayName("onWithdrawal")
    private static Event2Args<Hash160, Integer> onWithdrawal;

    /**
     * Gets the curve that is used to create the signature for withdrawals.
     *
     * @return the curve.
     */
    @Safe
    public static int getCurve() {
        return contractMap.getInteger(curveKey);
    }

    /**
     * Changes the curve that is used to create the signature for withdrawals.
     *
     * @param newCurve The new curve.
     * @throws Exception if the curve is not allowed.
     */
    public static void changeCurve(int newCurve) throws Exception {
        int curve = contractMap.getInteger(curveKey);
        if (newCurve == curve) {
            return;
        }
        // Secp256k1 = 22
        // Secp256r1 = 23 (default)
        if (newCurve != 22 && newCurve != 23) {
            throw new Exception("Curve not supported.");
        }
        contractMap.put(curveKey, newCurve);
    }

    /**
     * This method provides a blacklist functionality. The old {@code Total Earned Amount} is checked, so
     * that no immediate withdrawal takes place before executing this.
     *
     * @param account The account to set the {@code Total Earned Amount} for.
     * @param oldTea  The previous {@code Total Earned Amount} for that account.
     * @param newTea  The new {@code Total Earned Amount} for that account.
     * @throws Exception if the caller is not allowed to invoke this method.
     */
    public static void setTotalEarnedAmount(Hash160 account, int oldTea, int newTea) throws Exception {

        ECPoint owner = new ECPoint(contractMap.get(ownerKey));
        if (!Runtime.checkWitness(owner)) {
            throw new Exception("No authorization.");
        }

        ByteString alreadyWithdrawn = withdrawalMap.get(account.toByteString());
        int alreadyWithdrawnInt;
        if (alreadyWithdrawn == null) {
            alreadyWithdrawnInt = 0;
        } else {
            alreadyWithdrawnInt = alreadyWithdrawn.toInt();
            if (alreadyWithdrawnInt != oldTea) {
                throw new Exception("Funds were withdrawn in the meantime.");
            }
        }
        if (newTea < alreadyWithdrawnInt) {
            throw new Exception("The provided amount is lower than the already withdrawn amount.");
        }
        withdrawalMap.put(account.toByteString(), newTea);
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
     * Upon deployment, the initial owner is set and the curve is set to the secp256r1 curve.
     *
     * @param data   The initial owner.
     * @param update True, if the contract is being deployed, false if it is updated.
     */
    @OnDeployment
    public static void deploy(Object data, boolean update) {
        if (!update) {
            ECPoint initialOwner = (ECPoint) data;
            contractMap.put(ownerKey, initialOwner.toByteString());
            contractMap.put(curveKey, 23);
        }
    }

}
