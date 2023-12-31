import { isAddress } from "ethers";
import * as yup from "yup";
import type { AnyObject, Maybe } from "yup/lib/types";

yup.addMethod<yup.StringSchema>(yup.string, "isEthereumAddress", () =>
  yup
    .string()
    .test(
      "isEthereumAddress",
      "Must be a valid Ethereum address!",
      (value: string) => isAddress(value)
    )
);

declare module "yup" {
  interface StringSchema<
    TType extends Maybe<string> = string | undefined,
    TContext extends AnyObject = AnyObject,
    TOut extends TType = TType
  > extends yup.BaseSchema<TType, TContext, TOut> {
    isEthereumAddress(): StringSchema<TType, TContext>;
  }
}

export default yup;
