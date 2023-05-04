import type { ValidationError } from "yup";
import type yup from "./yup";

export async function getAllFormErrors(
  formValues: any,
  schema: yup.ObjectSchema<any>
): Promise<ValidationError[]> {
  try {
    await schema.validate(formValues, { abortEarly: false });
    return [];
  } catch (error) {
    return error.inner;
  }
}
