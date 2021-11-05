package io.flatfeestack;

import io.neow3j.types.Hash160;
import io.neow3j.wallet.Account;

import java.math.BigInteger;

public class EvaluationHelper {

    static Hash160[] getRandomHashes(int arrLength) {
        Hash160[] arr = new Hash160[arrLength];
        for (int i = 0; i < arrLength; i++) {
            arr[i] = Account.create().getScriptHash();
        }
        return arr;
    }

    static BigInteger[] getUniformTeas(int arrLength, BigInteger start, BigInteger step) {
        BigInteger[] arr = new BigInteger[arrLength];
        BigInteger tea = start;
        for (int i = 0; i < arrLength; i++) {
            arr[i] = tea;
            tea = tea.add(step);
        }
        return arr;
    }

    static BigInteger[] getRandomTeasToPreset(int nrAccounts, long min, long multiplier) {
        BigInteger[] arr = new BigInteger[nrAccounts];
        for (int i = 0; i < nrAccounts; i++) {
            BigInteger rand = BigInteger.valueOf((long) (Math.random() * multiplier) + min);
            arr[i] = rand;
        }
        return arr;
    }

    static BigInteger getSum(BigInteger[] arr) {
        BigInteger totalAmount = BigInteger.ZERO;
        for (BigInteger val : arr) {
            totalAmount = totalAmount.add(val);
        }
        return totalAmount;
    }

}
