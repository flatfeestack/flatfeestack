import yup from "../utils/yup";

export const commentSchema = yup.object().shape({
  content: yup.string().min(1).max(500).required(),
});

export const postSchema = yup.object().shape({
  content: yup.string().min(1).max(1000).required(),
  title: yup.string().min(1).max(100).required(),
});
