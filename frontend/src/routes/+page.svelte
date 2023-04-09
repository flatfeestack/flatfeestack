<script lang="ts">
  import ffsLogoWhite from "$lib/images/ffs-logo-white.svg";

  import holdingWorldBw from "$lib/images/holding-world-bw.svg";
  import holdingWorld from "$lib/images/holding-world.svg";

  import houseBw from "$lib/images/house-bw.svg";
  import house from "$lib/images/house.svg";

  import runningBw from "$lib/images/running-bw.svg";
  import runnning from "$lib/images/running.svg";

  import metricsBw from "$lib/images/metrics-bw.svg";
  import metrics from "$lib/images/metrics.svg";

  import seatBw from "$lib/images/seat-bw.svg";
  import seat from "$lib/images/seat.svg";

  import ccBw from "$lib/images/cc-bw.svg";
  import cc from "$lib/images/cc.svg";

  import axlabs from "$lib/images/axlabs.svg";
  import axelra from "$lib/images/axelra.svg";
  import coinblesk from "$lib/images/coinblesk.svg";
  import csg from "$lib/images/csg.svg";
  import dsl from "$lib/images/dsl.svg";
</script>

<style>
  section {
    width: 100%;
  }

  .logoFFSWhiteSmall {
    width: 15rem;
    padding-top: 6rem;
    padding-bottom: 2rem;
  }

  .logoFFSWhiteLarge {
    width: 30rem;
    padding-top: 6rem;
    padding-bottom: 2rem;
  }

  .icon {
    width: 8rem;
    padding-top: 1rem;
  }

  .iconSupport {
    width: 10rem;
    padding: 2rem;
    filter: grayscale(100%) opacity(50%);
    align-content: center;
    transition: filter 0.5s;
  }

  .iconSupport:hover {
    filter: none;
  }

  .parallax {
    text-align: center;
    justify-content: center;
    background-image: url($lib/images/code-dark2.jpg);
    color: #fff;
    background-attachment: fixed;
    background-position: center;
    background-repeat: no-repeat;
    background-size: cover;
    box-shadow: 0 0 5px 5px var(--secondary-200);
    margin-bottom: 1em;
  }

  .flex-col {
    display: flex;
    flex-direction: column;
  }

  .flex-row {
    display: flex;
    flex-direction: row;
    justify-content: center;
  }

  .wrap {
    flex-wrap: wrap;
    align-items: center;
  }

  /** replaced by section curerntly 
  main {
    width: 100%;
  }
*/

  h1 {
    padding-bottom: 4rem;
  }
  h2 {
    padding-top: 2rem;
  }

  .card {
    color: var(--secondary-700);
    border-radius: 5px;
    box-shadow: 1px 1px 2px 1px var(--secondary-200);
    background-color: var(--secondary-100);
    padding: 2rem 2rem 4rem;
    margin: 2em;
    justify-content: center;
    gap: 1rem;
    text-align: center;
    align-items: center;
  }

  .cardLeft {
    text-align: left;
  }

  .card img.bottom {
    position: absolute;
    animation: fadeIn 0.5s;
    opacity: 1;
  }

  .card img.top {
    position: relative;
    animation: fadeOut 0.5s;
    opacity: 0;
  }

  .card:hover img.bottom {
    position: absolute;
    animation: fadeOut 0.5s;
    opacity: 0;
  }

  .card:hover img.top {
    position: relative;
    animation: fadeIn 0.5s;
    opacity: 1;
  }

  @keyframes fadeIn {
    0% {
      opacity: 0;
    }
    100% {
      opacity: 1;
    }
  }

  @keyframes fadeOut {
    0% {
      opacity: 1;
    }
    100% {
      opacity: 0;
    }
  }

  /* Small only */
  @media screen and (max-width: 39.9375em) {
    .flex-row {
      flex-direction: column;
    }

    .rightBorder {
      border-right: 0;
    }
  }

  /* Medium and up */
  @media screen and (min-width: 40em) {
    .flex-row {
      flex-direction: row;
    }

    .rightBorder {
      border-right: 2px solid #fff;
    }
  }
</style>

<svelte:head>
  <title>FlatFeeStack</title>
  <meta name="description" content="FlatFeeStak Home" />
</svelte:head>

<section>
  <div class="flex-col">
    <div class="parallax">
      <img
        class="hide-mda logoFFSWhiteSmall"
        src={ffsLogoWhite}
        alt="FlatFeeStack Logo"
      />
      <img
        class="hide-sx logoFFSWhiteLarge"
        src={ffsLogoWhite}
        alt="FlatFeeStack Logo"
      />
      <h1>
        A flat-fee sponsoring platform for <br />
        open source projects based on git metrics
      </h1>
    </div>
  </div>

  <div class="flex-row py-300">
    <div class="card">
      <div>
        <img
          class="icon bottom"
          src={holdingWorldBw}
          alt="Hands holding world icon bw"
        />
        <img
          class="icon top"
          src={holdingWorld}
          alt="Hands holding world icon color"
        />
      </div>
      <h2>On the Shoulders of Giants</h2>
      <div class="max-w20">
        Without open source software, your software project would not exist
      </div>
    </div>
    <div class="card">
      <div>
        <img class="icon bottom" src={houseBw} alt="House icon bw" />
        <img class="icon top" src={house} alt="House icon bw" />
      </div>
      <h2>Our Mission</h2>
      <div class="max-w20">
        &#171;Open source contributors should be able to make a living with open
        source software&#187;
      </div>
    </div>
  </div>

  <div class="flex-col">
    <div class="flex-row py-300 parallax">
      <div class="p-300 max-w20 rightBorder">
        <h3>
          For a flat fee of $125/year, sponsor any open source projects. Show
          your supported projects badge.
        </h3>
        <a href="/user/signup">
          <button class="button1">Sign Up</button>
        </a>
      </div>
      <div class="p-300 max-w20">
        <h3>
          Earn money if your open source project is sponsored. All you need is a
          public git repository.
        </h3>
        <a href="/user/signup">
          <button class="button1">Sign Up</button>
        </a>
      </div>
    </div>
  </div>

  <div class="flex-row">
    <div class="flex-row card max-w50 cardLeft">
      <div>
        <img class="icon bottom" src={runningBw} alt="person running icon bw" />
        <img class="icon top" src={runnning} alt="person running icon color" />
      </div>
      <div>
        <h2>Why Sponsoring?</h2>
        <div class="max-w40">
          Sponsoring a project gives incentives to its maintainers to add
          features, fix bugs, and security patches. Sponsoring puts social
          pressure and expectations onto a project. Sponsoring provides
          incentives to keep the project active and healthy.
        </div>
      </div>
    </div>
  </div>

  <div class="flex-row">
    <div class="flex-row card max-w50 cardLeft">
      <div>
        <h2>Why FlatFeeStack?</h2>
        <div class="max-w40">
          Donation services such as Open Collective, Patreon, or Github Sponsors
          are great to sponsor a dedicated person or a whole projects.
          FlatFeeStack distributes the funds according to git metrics. For
          $125/year, you can sponsor as many project you like with one click.
        </div>
      </div>
      <div>
        <img class="icon bottom" src={metricsBw} alt="metric icon bw" />
        <img class="icon top" src={metrics} alt="metric icon bw" />
      </div>
    </div>
  </div>

  <div class="flex-row">
    <div class="flex-row card max-w50 cardLeft">
      <div>
        <img class="icon bottom" src={seatBw} alt="seat icon bw" />
        <img class="icon top" src={seat} alt="metric icon color" />
      </div>
      <div>
        <h2>Which Projects Should your Company Sponsor?</h2>
        <div class="max-w40">
          All of them! Moreover, delegate the task of sponsoring projects to
          your engineers. They know best what to sponsor. FlatFeeStack supports
          seats, which makes your accounting happy, due to the predictable
          sponsoring amount. Your sponsor badge will show all the projects your
          company is passionate about.
        </div>
      </div>
    </div>
  </div>

  <div class="flex-row">
    <div class="flex-row card max-w50 cardLeft">
      <div>
        <h2>Why Crypto Currencies?</h2>
        <div class="max-w40">
          Payin with credit card, payout with crypto currencies. With crypto
          currencies, the payout is without restrictions and anyone can receive
          funding. It took Github <a
            href="https://news.ycombinator.com/item?id=25648849">two years</a
          >
          to get the license for Iran. These restrictions do not apply to crypto
          currencies. We chose crypto currency based on: (1) energy-friendly, (2)
          popularity, (3) decentralization.
        </div>
      </div>
      <div>
        <img class="icon bottom" src={ccBw} alt="crypto currencies icon bw" />
        <img class="icon top" src={cc} alt="crypto currencies icon bw" />
      </div>
    </div>
  </div>

  <div class="flex-col">
    <h2 class="text-center">Supported by</h2>
    <div class="flex-row wrap">
      <a href="https://axlabs.com">
        <img class="iconSupport" src={axlabs} alt="AxLabs" />
      </a>
      <a href="https://axelra.com">
        <img class="iconSupport" src={axelra} alt="Axelra" />
      </a>
      <a href="https://coinblesk.com">
        <img class="iconSupport" src={coinblesk} alt="Coinblesk" />
      </a>
      <a href="https://www.csg.uzh.ch">
        <img class="iconSupport" src={csg} alt="csg" />
      </a>
      <a href="https://dsl.rj.ost.ch">
        <img class="iconSupport" src={dsl} alt="dsl" />
      </a>
    </div>
  </div>
</section>
