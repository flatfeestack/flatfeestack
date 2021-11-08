package io.flatfeestack;

import java.math.BigInteger;

import static io.flatfeestack.EvaluationHelper.HUNDRED_GAS;
import static io.flatfeestack.EvaluationHelper.MAX_ACCOUNTS_BATCH_PAYOUT_MAP;
import static io.flatfeestack.EvaluationHelper.ONE_GAS;
import static io.flatfeestack.EvaluationHelper.TENTH_GAS;
import static io.flatfeestack.EvaluationHelper.TEN_GAS;
import static io.flatfeestack.EvaluationHelper.TSD_GAS;

public enum EvaluationTypeMap {

    ONE_ACC_TENTH_GAS(1, TENTH_GAS),
    ONE_ACC_ONE_GAS(1, ONE_GAS),
    ONE_ACC_TEN_GAS(1, TEN_GAS),
    ONE_ACC_32_GAS(1, new BigInteger("2147483647")),
    ONE_ACC_64_GAS(1, new BigInteger("2147483648")),
    ONE_ACC_HUNDRED_GAS(1, HUNDRED_GAS),
    ONE_ACC_TSD_GAS(1, TSD_GAS),
    TEN_ACC_TENTH_GAS(10, TENTH_GAS),
    TEN_ACC_ONE_GAS(10, ONE_GAS),
    TEN_ACC_TEN_GAS(10, TEN_GAS),
    TEN_ACC_32_GAS(10, new BigInteger("2147483647")),
    TEN_ACC_64_GAS(10, new BigInteger("2147483648")),
    TEN_ACC_HUNDRED_GAS(10, HUNDRED_GAS),
    TEN_ACC_TSD_GAS(10, TSD_GAS),
    HUNDRED_ACC_TENTH_GAS(100, TENTH_GAS),
    HUNDRED_ACC_ONE_GAS(100, ONE_GAS),
    HUNDRED_ACC_TEN_GAS(100, TEN_GAS),
    HUNDRED_ACC_32_GAS(100, new BigInteger("2147483647")),
    HUNDRED_ACC_64_GAS(100, new BigInteger("2147483648")),
    HUNDRED_ACC_HUNDRED_GAS(100, HUNDRED_GAS),
    HUNDRED_ACC_TSD_GAS(100, TSD_GAS),
    MAX_ACC_TENTH_GAS(MAX_ACCOUNTS_BATCH_PAYOUT_MAP, TENTH_GAS),
    MAX_ACC_ONE_GAS(MAX_ACCOUNTS_BATCH_PAYOUT_MAP, ONE_GAS),
    MAX_ACC_TEN_GAS(MAX_ACCOUNTS_BATCH_PAYOUT_MAP, TEN_GAS),
    MAX_ACC_32_GAS(MAX_ACCOUNTS_BATCH_PAYOUT_MAP, new BigInteger("2147483647")),
    MAX_ACC_64_GAS(MAX_ACCOUNTS_BATCH_PAYOUT_MAP, new BigInteger("2147483648")),
    MAX_ACC_HUNDRED_GAS(MAX_ACCOUNTS_BATCH_PAYOUT_MAP, HUNDRED_GAS),
    MAX_ACC_TSD_GAS(MAX_ACCOUNTS_BATCH_PAYOUT_MAP, TSD_GAS);

    BigInteger nrAccounts;
    BigInteger tea;

    EvaluationTypeMap(int nrAccounts, BigInteger tea) {
        this.nrAccounts = BigInteger.valueOf(nrAccounts);
        this.tea = tea;
    }

}
