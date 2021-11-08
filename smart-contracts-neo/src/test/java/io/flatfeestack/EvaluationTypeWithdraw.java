package io.flatfeestack;

import java.math.BigInteger;

import static io.flatfeestack.EvaluationHelper.HUNDRED_GAS;
import static io.flatfeestack.EvaluationHelper.ONE_GAS;
import static io.flatfeestack.EvaluationHelper.TENTH_GAS;
import static io.flatfeestack.EvaluationHelper.TEN_GAS;
import static io.flatfeestack.EvaluationHelper.TSD_GAS;

public enum EvaluationTypeWithdraw {

    TENTH_GAS_VAL(TENTH_GAS),
    ONE_GAS_VAL(ONE_GAS),
    TEN_GAS_VAL(TEN_GAS),
    INT32_GAS_VAL(new BigInteger("2147483647")),
    INT64_GAS_VAL(new BigInteger("2147483648")),
    HUNDRED_GAS_VAL(HUNDRED_GAS),
    TSD_GAS_VAL(TSD_GAS);

    BigInteger tea;

    EvaluationTypeWithdraw(BigInteger tea) {
        this.tea = tea;
    }

}
