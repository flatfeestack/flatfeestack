package io.flatfeestack;

import java.math.BigInteger;

import static io.flatfeestack.EvaluationHelper.HUNDRED_GAS;
import static io.flatfeestack.EvaluationHelper.INT32_LIMIT_GAS;
import static io.flatfeestack.EvaluationHelper.INT64_LIMIT_GAS;
import static io.flatfeestack.EvaluationHelper.ONE_GAS;

public enum EvaluationTypeWithdraw {

    INT32(INT32_LIMIT_GAS, null),
    INT64(INT64_LIMIT_GAS, null),
    INT32_PRESET_INT32(INT32_LIMIT_GAS, ONE_GAS),
    INT64_PRESET_INT32(INT64_LIMIT_GAS, ONE_GAS),
    INT64_PRESET_INT64(HUNDRED_GAS, INT64_LIMIT_GAS);

    BigInteger tea;
    BigInteger presetTea;

    public String getTeaType() {
        if (tea.compareTo(INT32_LIMIT_GAS) > 0) {
            return "int64";
        }
        return "int32";
    }

    public String getPresetTeaType() {
        if (presetTea == null) {
            return "-";
        }
        if (presetTea.compareTo(INT32_LIMIT_GAS) > 0) {
            return "int64";
        }
        return "int32";
    }

    public boolean hasPresetTea() {
        return presetTea != null;
    }

    EvaluationTypeWithdraw(BigInteger tea, BigInteger presetTea) {
        this.tea = tea;
        this.presetTea = presetTea;
    }

}
