<script>
  export let email = "";
  import { getContext } from "svelte";
  const { close } = getContext("simple-modal");
  const sendEmail = async () => {
    console.log(email)
    const axios = require("axios");
    await axios({
      url: "https://flatfeestack.com/email.php",
      method: "POST",
      data: {
        email: email
      }
    }).then(function(response) {
      //console.log(response);
      close();
    })
    .catch(function(error) {
      //console.log(error);
      close();
    });
  };
</script>

<div class="container">
  <div class="section-title">We are not ready yet!</div>
  <div>&nbsp;</div>
  <div class="subtitle">But we will launch soon!</div>
  <div>&nbsp;</div>
  <div>If your are interested to know more, give us your email</div>
  <div>&nbsp;</div>
  <div class="flex"><input class="flex-2" bind:value={email} placeholder="E-Mail">&nbsp;
    <button on:click={sendEmail} class="btn-base btn-primary flex-1">Submit</button>
  </div>
</div>