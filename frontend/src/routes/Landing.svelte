<style>
    .img {
        max-width: 400px;
        object-fit: contain;
        width: 100%;
        height: auto;
        margin: auto;
    }

    .hero-bg {
        background-image: linear-gradient(
                229deg,
                rgba(6, 98, 59, 0.04) 0%,
                rgba(100, 157, 102, 0.04) 93%
        );
        margin-top: calc(-1 * var(--navbar-height));
        padding-top: var(--navbar-height);
    }

    .spacing {
        margin-top: calc(2.5 * var(--spacing-x5));
    }

    .padding {
        padding-right: calc(2 * var(--spacing-x5));
    }

    .red {
        color: #c61010;
    }


</style>

<script lang="ts">
  import Spacer from "../components/UI/Spacer.svelte";
  import { getContext } from 'svelte';
  import Popup from "../components/UI/Popup.svelte";

  const { open } = getContext('simple-modal');

  const showWIP = () => {
    open(Popup, { message: "It's a modal!" });
  };
</script>

<section class="hero-bg">
  <div class="container">
    <div class="flex spacing">
      <div class="flex-1 padding">
        <div class="lead">Show your <span class="red">♥</span>️ for Open Source Software</div>
        <Spacer x5 />
        <div class="lead2">
          Flatfeestack is a flat-fee open source software donations platform for companies and individuals. Any
          developer can support awesome open source projects. The only requirement: the awesome open source
          project needs a public git repository.
        </div>
      </div>
      <div class="flex-1 justify-center">
        <img
          class="img"
          src="assets/images/Flatfeestack_LowPolyTree.png"
          alt="Tree"
        />
      </div>
    </div>
    <Spacer x5 />
    <Spacer x5 />
  </div>
</section>

<section>
  <div class="container">
    <div class="flex spacing">
      <div class="section-title">&#171;Software Engineers should be able to make a living
        with open source software&#187;
      </div>
    </div>
    <Spacer x5 />
    <div class="subtitle">
      Some individuals already live their dream of developing open source software and manage to make a decent living.
      However, those are exceptions rather then the norm. We want to change this!
    </div>
    <div class="flex spacing">
      <div class="container">
        <div class="section-title justify-center">Become a Sponsor</div>
        <div class="subtitle justify-center">Support your favorite stack and projects.</div>
        <Spacer x5 />
        <div class="justify-center">
          <button class="btn-base btn-primary" on:click={showWIP}>Support Open Source</button>
        </div>
        <Spacer x5 />
        <div class="justify-center">
          For a flat fee of $10 per month, you can support as many open source project
          as you like. Your developer badge shows which projects you are
          passionate about.
        </div>
      </div>
      <div class="vl"></div>
      <div class="container">
        <div class="section-title justify-center">Monetize your Project</div>
        <div class="subtitle justify-center">Put a value on your open source contribution.</div>
        <Spacer x5 />
        <div class="justify-center">
          <button class="btn-base btn-secondary" on:click={showWIP}>Receive Support</button>
        </div>
        <Spacer x5 />
        <div class="justify-center">
          Earn money if your open source project is backed. Your developer badge
          shows a multiplicator, earned money divided by $10. Compare your
          score with other developers.
        </div>
      </div>
    </div>
    <Spacer x5 />
  </div>
  <Spacer x5 />
</section>
<Spacer x5 />
<Spacer x5 />

<section class="section-padding hero-bg">
  <div class="container">
    <div class="section-title">Why Flatfeestack?</div>
    <Spacer x5 />
    <div class="flex">
      <div class="flex-1 padding vertical">
        <div class="container">
          <div class="subtitle">Agony of choice</div>
          <Spacer x3 />
          <div>
            Donation services such as Open Collective, Patreon, or Github Sponsors support
            project or person based contributions: e.g., support project X with $50
            per year, or support person Y with $50 per year. The problem is that the
            more projects or people you support, the more you pay. When coming across an interesting project,
            we often end up not supporting, as we would have to support at least 50 other interesting projects.
          </div>
        </div>
      </div>
      <div class="flex-1 justify-center"><img class="img" src="assets/images/flat1.svg" alt="Support" /></div>
    </div>
    <div class="flex spacing">
      <div class="flex-1 justify-center"><img class="img" src="assets/images/flat2.svg" alt="XLS" /></div>
      <div class="flex-1 padding vertical">
        <div class="container">
          <div class="subtitle">Flatfeestack is compatible with your accounting department</div>
          <Spacer x3 />
          <div>
            Sponsoring open source projects is difficult for a company.
            Try to convince your accounting department to sponsor &lt;insert-open-source-project&gt;. Flatfeestack is
            compatible with your
            accounting department. If the accounting department calculates $1500/per year per developer for its IT cost,
            with Flatfeestack it is $120 (12 month x $10) more: $1620. And the developer can support as many projects as you want.
          </div>
        </div>
      </div>
    </div>
    <div class="flex spacing">
      <div class="flex-1 padding vertical">
        <div class="container">
          <div class="subtitle">Simplified flat-fee approach</div>
          <Spacer x3 />
          <div>
            Each project receives your donation equally divided. Simply donate and
            don't worry about how many projects you want to support, it is always $10 per month. It is a similar concept to Flattr
            for open source projects, where git commits decide how much each contributor earns.
          </div>
        </div>
      </div>
      <div class="flex-1 justify-center"><img class="img" src="assets/images/flat3.svg" alt="Support" /></div>
    </div>
    <Spacer x5 />
    <Spacer x5 />
  </div>
</section>

<Spacer x5 />
<section>
  <div class="container">
    <div class="subtitle2">Flatfeestack is supported by:</div>
    <div class="flex">
      <a href="https://axlabs.io" class="flex-1 justify-center grey"><img class="img" src="assets/images/axlabs.png" alt="AxLabs" /></a>
      <a href="https://axelra.com" class="flex-1 justify-center grey2"><img class="img" src="assets/images/axelra.svg" alt="Axelra" /></a>
      <a href="https://coinblesk.com" class="flex-1 justify-center grey2"><img class="img" src="assets/images/coinblesk.svg" alt="Coinblesk" /></a>
    </div>
  </div>
</section>