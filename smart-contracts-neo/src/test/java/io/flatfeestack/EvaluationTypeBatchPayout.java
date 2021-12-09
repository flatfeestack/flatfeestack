package io.flatfeestack;

import java.math.BigInteger;

import static io.flatfeestack.EvaluationHelper.HUNDRED_GAS;
import static io.flatfeestack.EvaluationHelper.INT32_LIMIT_GAS;
import static io.flatfeestack.EvaluationHelper.INT64_LIMIT_GAS;
import static io.flatfeestack.EvaluationHelper.ONE_GAS;
import static io.flatfeestack.EvaluationHelper.TENTH_GAS;
import static io.flatfeestack.EvaluationHelper.TEN_GAS;

public enum EvaluationTypeBatchPayout {

    ONE_ACC_TENTH_GAS(1, TENTH_GAS),
    ONE_ACC_ONE_GAS(1, ONE_GAS),
    ONE_ACC_TEN_GAS(1, TEN_GAS),
    ONE_ACC_32_GAS(1, INT32_LIMIT_GAS),
    ONE_ACC_64_GAS(1, INT64_LIMIT_GAS),
    ONE_ACC_HUNDRED_GAS(1, HUNDRED_GAS),
    TEN_ACC_TENTH_GAS(10, TENTH_GAS),
    TEN_ACC_ONE_GAS(10, ONE_GAS),
    TEN_ACC_TEN_GAS(10, TEN_GAS),
    TEN_ACC_32_GAS(10, INT32_LIMIT_GAS),
    TEN_ACC_64_GAS(10, INT64_LIMIT_GAS),
    TEN_ACC_HUNDRED_GAS(10, HUNDRED_GAS),
    HUNDRED_ACC_TENTH_GAS(100, TENTH_GAS),
    HUNDRED_ACC_ONE_GAS(100, ONE_GAS),
    HUNDRED_ACC_TEN_GAS(100, TEN_GAS),
    HUNDRED_ACC_32_GAS(100, INT32_LIMIT_GAS),
    HUNDRED_ACC_64_GAS(100, INT64_LIMIT_GAS),
    HUNDRED_ACC_HUNDRED_GAS(100, HUNDRED_GAS),
    TSD_ACC_TENTH_GAS(1000, TENTH_GAS),
    TSD_ACC_ONE_GAS(1000, ONE_GAS),
    TSD_ACC_TEN_GAS(1000, TEN_GAS),
    TSD_ACC_32_GAS(1000, INT32_LIMIT_GAS),
    TSD_ACC_64_GAS(1000, INT64_LIMIT_GAS),
    TSD_ACC_HUNDRED_GAS(1000, HUNDRED_GAS),
    TSDTWELVE_ACC_TENTH_GAS(1012, TENTH_GAS),
    TSDTWELVE_ACC_ONE_GAS(1012, ONE_GAS),
    TSDTWELVE_ACC_TEN_GAS(1012, TEN_GAS),
    TSDTWELVE_ACC_32_GAS(1012, INT32_LIMIT_GAS),
    TSDTWELVE_ACC_64_GAS(1012, INT64_LIMIT_GAS),
    TSDTWELVE_ACC_HUNDRED_GAS(1012, HUNDRED_GAS);

    BigInteger nrAccounts;
    BigInteger tea;

    EvaluationTypeBatchPayout(int nrAccounts, BigInteger tea) {
        this.nrAccounts = BigInteger.valueOf(nrAccounts);
        this.tea = tea;
    }

}
