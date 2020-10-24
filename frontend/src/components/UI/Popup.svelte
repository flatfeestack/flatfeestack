<script>
export let email = "";
import { getContext } from "svelte";
const { close } = getContext("simple-modal");
const sendEmail = async () => {
  console.log(email);
  const axios = require("axios");
  await axios({
    url: "https://flatfeestack.com/email.php",
    method: "POST",
    data: {
      email: email,
    },
  })
    .then(function (response) {
      //console.log(response);
      close();
    })
    .catch(function (error) {
      //console.log(error);
      close();
    });
};
</script>

<div class="container">
  <h2 class="text-lg">We are not ready yet!</h2>
  <div>&nbsp;</div>
  <div class="subtitle">But we will launch soon!</div>
  <div>&nbsp;</div>
  <div>If your are interested to know more, give us your email</div>
  <div>&nbsp;</div>
  <div class="flex">
    <input
      class="w-2/3 shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
      bind:value="{email}"
      placeholder="E-Mail"
    />&nbsp;
    <button
      on:click="{sendEmail}"
      class="bg-primary-500 p-2 w-1/3 text-white"
    >Submit</button>
  </div>
</div>
