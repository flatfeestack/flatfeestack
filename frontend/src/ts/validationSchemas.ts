import yup from "../utils/yup";

export const commentSchema = yup.object().shape({
  content: yup.string().min(1).max(500).required(),
});
