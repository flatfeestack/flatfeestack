package io.flatfeestack;

import io.neow3j.devpack.ByteString;
import io.neow3j.devpack.Hash160;
import io.neow3j.devpack.Helper;
import io.neow3j.devpack.Runtime;
import io.neow3j.devpack.Storage;
import io.neow3j.devpack.StorageContext;
import io.neow3j.devpack.StorageMap;
import io.neow3j.devpack.annotations.DisplayName;
import io.neow3j.devpack.annotations.OnDeployment;
import io.neow3j.devpack.annotations.OnNEP17Payment;
import io.neow3j.devpack.annotations.Permission;
import io.neow3j.devpack.annotations.Safe;
import io.neow3j.devpack.contracts.GasToken;
import io.neow3j.devpack.events.Event2Args;

import static io.neow3j.devpack.StringLiteralHelper.addressToScriptHash;

@Permission(contract = "0xd2a4cff31913016155e38e474a2c06d08be276cf")
public class PreSignNeo {

    static final StorageContext ctx = Storage.getStorageContext();
    static final byte[] contractMapPrefix = Helper.toByteArray((byte) 0x01);
    static final byte[] ownerKey = Helper.toByteArray((byte) 0xff);
    static final StorageMap contractMap = ctx.createMap(contractMapPrefix);

    static final byte[] balanceMapPrefix = Helper.toByteArray((byte) 0x10);
    static final StorageMap balanceMap = ctx.createMap(balanceMapPrefix);

    public static void setOwner(Hash160 newOwner) throws Exception {
        if (!Runtime.checkWitness(new Hash160(contractMap.get(ownerKey)))) {
            throw new Exception("No authorization.");
        }
        contractMap.put(ownerKey, newOwner.toByteString());
    }

    @Safe
    public static Hash160 getOwner() {
        return new Hash160(contractMap.get(ownerKey));
    }

//    public static void register(Hash160[] accounts) {
//        for (Hash160 account : accounts) {
//            // Crucial: Never delete an account that was once put on the balanceMap!
//            if (balanceMap.get(account.toByteString()) == null) {
//                balanceMap.put(account.toByteString(), 0);
//            }
//        }
//    }

    public static int withdraw(Hash160 account, int totalAmountOverall) throws Exception {
        if (!Runtime.checkWitness(new Hash160(contractMap.get(ownerKey)))) {
            throw new Exception("No authorization.");
        }
        ByteString balance = balanceMap.get(account.toByteString());
        if (balance == null) {
            balanceMap.put(account.toByteString(), totalAmountOverall);
            GasToken.transfer(Runtime.getExecutingScriptHash(), account, totalAmountOverall, null);
            return totalAmountOverall;
        }

        int alreadyWithdrawn = balance.toInteger();
        int amountToWithdraw = totalAmountOverall - alreadyWithdrawn;
        if (amountToWithdraw <= 0) {
            throw new Exception("These funds have already been withdrawn.");
        }
        GasToken.transfer(Runtime.getExecutingScriptHash(), account, amountToWithdraw,
                null);
        return totalAmountOverall;
    }

    @OnDeployment
    public static void deploy(Object data, boolean update) {
        if (!update) {
            contractMap.put(ownerKey, addressToScriptHash("NV1Q1dTdvzPbThPbSFz7zudTmsmgnCwX6c"));
        }
    }

    @DisplayName("onContractFunding")
    private static Event2Args<Hash160, Integer> onContractFunding;

    @OnNEP17Payment
    public static void onNep17Payment(Hash160 from, int amount, Object data) {
        onContractFunding.fire(from, amount);
    }

}
