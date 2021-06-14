<script lang="ts">
  import { draw } from "svelte/transition";
  import { fly, fade } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import App from "../ts/sphere.ts";
  import { onMount } from "svelte";
  import { navigate } from "svelte-routing";

  let canvas: HTMLCanvasElement;

  function handleSubmit() {
    navigate("/signup");
  }

  onMount(async () => {
    const a = new App(canvas);
    a.SetDrawMode("LINES");
    a.SetZoom(-2);
    a.SetRotation("X", 0.00002);
    a.SetRotation("Y", -0.00002);
    a.SetQuality("3");
    a.Draw();
  });

  let svgVisible = true;
</script>

<style>
    .grey img {

        transform: scale(0.5);
    }

    .grey2 img {
        transform: scale(0.4);
    }

    .big {
        font-weight: 500;
        font-size: 10vw;
        text-align: center;
        line-height: 90%;
        text-shadow: 2px 2px 2px #666;
        padding: 1rem;
        padding-left: 3rem;
    }

    .medium {
        font-weight: 300;
        font-size: 6vw;
        text-align: center;
        line-height: 95%;
        text-shadow: 1px 1px 1px #666;
        padding: 1rem;
        padding-left: 3rem;
    }

    .medium2 {
        font-weight: 300;
        font-size: 4vw;
        text-align: center;
        line-height: 95%;
        text-shadow: 1px 1px 1px #666;
        padding: 1rem;
        padding-left: 3rem;
    }

    .small {
        padding: 1em;
        padding-top: 3em;
        padding-bottom: 3em;
        font-weight: 300;
        font-size: 130%;
        line-height: 110%;
    }

    .small2 {
        font-weight: 300;
        font-size: 130%;
        line-height: 110%;
    }

    .img {
        height: 10rem;
        position: relative;
        top: -157px;
        padding-left: 11em;
        align-items: center;
        overflow: hidden;
    }

    .center {
        align-items: center;
        text-align: center;
        justify-content: center;
    }

    .dark {
        background-color: #333;
        color: #fff;
        margin: 0;
        padding: 0
    }

    .parallax {
        background-image: url("/assets/images/code-dark2.jpg");
        color: #fff;
        background-attachment: fixed;
        background-position: center;
        background-repeat: no-repeat;
        background-size: cover;
    }

    .gcontainer {
        display: grid;
        grid-template-columns: 1fr 1fr;
        padding: 2em;
        margin: 1em;
        grid-gap: 1em;
    }

    @media (max-width: 36rem) {
        .gcontainer {
            display: grid;
            grid-template-columns: 1fr;
            padding: 2em;
            margin: 1em;
            grid-gap: 1em;
        }
    }

    .gcontainer3 {
        display: grid;
        grid-template-columns: 1fr 1fr 1fr;
        padding: 2em;
        margin: 1em;
        grid-gap: 1em;
    }

    @media (max-width: 36rem) {
        .gcontainer3 {
            display: grid;
            grid-template-columns: 1fr;
            padding: 2em;
            margin: 1em;
            grid-gap: 1em;
        }
    }

    .whiteborder {
        border-right: 1px solid #fff;
    }

    @media (max-width: 36rem) {
        .whiteborder {
            border-right: 0px;
        }
    }


    p > :global(a) {
        color: white;
    }

</style>
<div class="container-col2">
  <div class="container2">
    <div>
      <p class="big" in:fly="{{ y: 200, duration: 1500, easing: cubicOut }}">
        On the Shoulders of Giants
      </p>
      <p class="medium" in:fly="{{ delay: 500, y: 200, duration: 1500, easing: cubicOut }}">
        Without open source software, your software project would not exist.
      </p>
    </div>
    <div class="container-col2 hide-md">
      <canvas class="hide-md" bind:this={canvas} width="750rem" height="750rem" id="canvas" />
      <img class="img hide-md" src="assets/images/atlas.svg">
      <!--https://github.com/aholmes/webgl-sphere-->
    </div>
  </div>

  <div class="container-col3 parallax center" style="padding-top:5em;padding-bottom:5em;">

    <div class="center container3">

      <p class="medium2 w1-2 whiteborder">
        Flatfeestack is a flat-fee donations platform. Donations are
        distributed based on git metrics.
      </p>

      <p class="medium2 w1-2">
        Our mission: &#171;Open source contributors should be able to make a living with open
        source software&#187;
      </p>

    </div>
    <form on:submit|preventDefault="{handleSubmit}">
      <button class="button1 center" type="submit" style="font-size: 120%; width:10em;">Sign Up</button>
    </form>
  </div>

  <div class="container2">
    <div class="small w1-3 container-col2">
      <p>
        Some individuals already live their dream of developing open source
        software and manage to make a decent living. However, those are
        exceptions rather then the norm.
      </p>
      <img src="assets/images/coding.svg" style="height: 10rem;">
    </div>

    <div class="small w1-3 container-col2">
      <p>
        For a flat fee of $32 per quarter (or $125 per year), you can support
        as many open source project as you like. Your developer badge shows
        which projects you support.
      </p>
      <img src="assets/images/handshake.svg" style="height: 10rem;">
    </div>

    <div class="small w1-3 container-col2">
      <p>
        Earn money if your open source project is backed. Your project badge
        shows the sponsors. To receive funding, your open source project
        needs a public git repository.
      </p>
      <img src="assets/images/trophy.svg" style="height: 10rem;">
    </div>
  </div>

  <div class="container-col2 dark" style="padding-top:5em;padding-bottom:5em;">

    <div class="gcontainer">
      <div>

        <h2>Why Flatfeestack?</h2>

        <p class="small2">Donation services such as Open Collective, Patreon, or Github
          Sponsors support project or person: e.g., support project X with $50 per year,
          or support person Y with $50 per year. This is a good start, but the more projects or people you
          support, the more you pay. When coming across an interesting or promising
          project, we often end up not supporting, as we would have to
          support many other interesting projects. With flatfeestack, you pay a flat-fee, never more.
        </p>

      </div>
      <div>
        <h2>How does it work?</h2>
        <p class="small2">
          Each project receives your donation equally divided. Simply donate
          and don't worry about how many projects you want to support, it is
          always $32 per quarter (or $125 per year). It is a similar concept to Flattr for open
          source projects, where git metrics decide how much each
          contributor earns.
        </p>

      </div>
      <div>

        <h2>Can an organization donate as well?</h2>

        <p class="small2">
          Flatfeestack is compatible with your accounting department
          Sponsoring open source projects is difficult for a company. Try to
          convince your accounting department to sponsor
          &lt;insert-open-source-project&gt;. Flatfeestack is compatible
          with your accounting department. If the accounting department
          calculates $1500/per year per developer for its IT cost, with
          Flatfeestack it is $125 more: $1625. And the
          developer can support as many projects as you want. In addition the
          organization gets an up-to-date list of supported open source
          projects.
        </p>

      </div>
      <div>

        <h2>Why Cryptocurrencies?</h2>
        <p class="small2">
          While the donation payment is done with credit cards, the payout is with cryptocurrency.
          We chose Binance Coin, as the energy footprint is much lower compared to Ethereum and Bitcoin. With
          crypto currencies, the payout is without restrictions and anyone can receive funding. While
          <a href="https://github.blog/2021-01-05-advancing-developer-freedom-github-is-fully-available-in-iran/">
            Github is finally available in Iran</a>, it took them <a
          href="https://news.ycombinator.com/item?id=25648585">
          two years to get the license</a>. These restrictions do not apply for cryptocurrencies. For selecting crypto
          currencies, we have the following priorities: 1) energy-friendly, 2) popularity, 3) decentralization.
        </p>

      </div>
    </div>

  </div>


  <div class="container-col2" style="padding-top:2em;padding-bottom:2em;">
    <h2 class="p-2 m-2">
      Flatfeestack is supported by:
    </h2>
    <div class="gcontainer3 center">
      <a href="https://axlabs.com" class="grey">
        <img class="w-100" src="assets/images/axlabs.svg" alt="AxLabs" />
      </a>
      <a href="https://axelra.com" class="grey2">
        <img class="w-100" src="assets/images/axelra.svg" alt="Axelra" />
      </a>
      <a href="https://coinblesk.com" class="grey2">
        <img class="w-100" src="assets/images/coinblesk.svg" alt="Coinblesk" />
      </a>
    </div>
  </div>

</div>
