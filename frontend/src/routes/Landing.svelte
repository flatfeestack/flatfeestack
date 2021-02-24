<style>
.section-title {
  @apply text-3xl;
}
@screen md {
  .section-title {
    @apply text-4xl;
  }
  .subtitle {
    @apply font-normal;
  }
}

.subtitle {
  @apply text-primary-400 text-lg font-bold;
}
.grey img {
  filter: grayscale(100%);
  transform: scale(0.5);
}
.grey2 img {
  filter: grayscale(100%);
  transform: scale(0.25);
}
</style>

<script lang="ts">
import Spacer from "../components/UI/Spacer.svelte";
import Navigation from "../components/Navigation.svelte";
import { draw } from "svelte/transition";


let svgVisible = true;
</script>

<div class="bg-gradient-to-tr from-light-100 to-primary-100">
  <Navigation />
  <div class="container mx-auto px-3 lg:px-0 pt-20">
    <div class="flex flex-wrap pb-10">
      <div
        class="w-100 md:w-1/2 py-10 flex flex-col items-center md:items-start"
      >
        <h1 class="text-4xl md:text-6xl font-bold text-center md:text-left">
          Show your
        </h1>
        {#if svgVisible}
          <svg
            height="100px"
            width="100px"
            x="0px"
            y="0px"
            viewBox="0 0 91 91"
          ><g>
              <path
                in:draw="{{ duration: 1500 }}"
                fill="none"
                stroke="#ff0000"
                stroke-width="7px"
                stroke-linecap="round"
                d="M52.7,71c-3.1,3.9,2.5,8.3,5.4,4.2c6.6-9.1,12.6-18.8,17.1-29.1c3.8-8.7,7-19.8,2.7-28.9c-4.7-10-15.6-3-20.5,2.7   c-4.6,5.4-7.7,11.8-10.3,18.4c-2.4-8-5.2-15.8-9.6-23C33.2,8,24.1,0.4,15.8,7.6c-8.4,7.3-4.6,19.3-1.2,28.1   c3.7,9.6,8.6,18.8,14.8,27c7,9.1,15.2,17.1,24.3,24"
              ></path>
            </g></svg>
        {/if}
        <h1 class="text-4xl md:text-6xl font-bold text-center md:text-left">
          for Open Source Software
        </h1>
        <div
          class="text-lg md:text-2xl  text-primary-500 text-center md:text-left"
        >
          Flatfeestack is a flat-fee open source software donations platform for
          companies and individuals. Any developer can support awesome open
          source projects. The only requirement: the awesome open source project
          needs a public git repository.
        </div>
      </div>
      <div class="justify-center align-center flex flex-1 md:w-1/2">
        <img
          class="object-contain w-64 h-64 m-auto"
          src="assets/images/Flatfeestack_LowPolyTree.png"
          alt="Tree"
        />
      </div>
    </div>
    <Spacer x5 />
    <Spacer x5 />
  </div>

  <section class="bg-white">
    <div class="container mx-auto px-3 lg:px-0 md:px-0 py-10">
      <div class="section-title pb-5 font-light">
        &#171;Software Engineers should be able to make a living with open
        source software&#187;
      </div>
      <Spacer x5 />
      <div class="subtitle">
        Some individuals already live their dream of developing open source
        software and manage to make a decent living. However, those are
        exceptions rather then the norm. We want to change this!
      </div>
      <div class="flex justify-between flex-wrap mt-10">
        <div
          class="w-full md:w-2/5 p-5 bg-secondary-500 text-white shadow-lg mb-5 md:mb-0 rounded-md"
        >
          <div class="section-title justify-center">Become a Sponsor</div>
          <div class="text-lg justify-center text-secondary-800 font-bold">
            Support your favorite stack and projects.
          </div>

          <div class="justify-center mt-5">
            For a flat fee of $10 per month (or $120 per year), you can support
            as many open source project as you like. Your developer badge shows
            which projects you are passionate about.
          </div>
          <div class="flex justify-center mt-5">
            <button
              class="bg-white p-2 text-secondary-500 rounded-sm"
            >Support Open Source</button>
          </div>
        </div>
        <div
          class="w-full md:w-2/5 p-5 text-white bg-primary-500 shadow-lg rounded-md"
        >
          <div class="section-title justify-center">Monetize your Project</div>
          <div class="text-lg justify-center text-primary-800 font-bold">
            Put a value on your open source contribution.
          </div>
          <div class="justify-center mt-5 font-light">
            Earn money if your open source project is backed. Your project badge
            shows the number of sponsors. Your developer badge shows in addition
            a score based on your earnings. Compare your score with other
            developers.
          </div>
          <div class="flex justify-center mt-5">
            <button
              class="bg-white p-2 text-primary-500 rounded-sm"
            >Receive Support</button>
          </div>
        </div>
      </div>
    </div>
  </section>

  <section>
    <div class="container container mx-auto px-3 lg:px-0 md:px-0 py-10">
      <div class="section-title pb-10 font-light">Why Flatfeestack?</div>
      <div class="flex flex-wrap mb-40">
        <div class="w-full md:w-1/2">
          <div class="container">
            <div class="subtitle">Agony of choice</div>
            <Spacer x3 />
            <div>
              Donation services such as Open Collective, Patreon, or Github
              Sponsors support project or person based contributions: e.g.,
              support project X with $50 per year, or support person Y with $50
              per year. The problem is that the more projects or people you
              support, the more you pay. When coming across an interesting
              project, we often end up not supporting, as we would have to
              support at least 50 other interesting projects.
            </div>
          </div>
        </div>
        <div
          class="w-full md:w-1/2 flex justify-center order-first md:order-none mb-5"
        >
          <img
            class="object-contain w-4/5 h-40"
            src="assets/images/flat1.svg"
            alt="Support"
          />
        </div>
      </div>
      <div class="flex flex-wrap mb-40">
        <div class="w-full md:w-1/2 flex justify-center mb-5">
          <img
            class="object-contain w-4/5 h-40"
            src="assets/images/flat2.svg"
            alt="XLS"
          />
        </div>
        <div class="w-full md:w-1/2">
          <div class="container">
            <div class="subtitle">
              Flatfeestack is compatible with your accounting department
            </div>
            <Spacer x3 />
            <div>
              Sponsoring open source projects is difficult for a company. Try to
              convince your accounting department to sponsor
              &lt;insert-open-source-project&gt;. Flatfeestack is compatible
              with your accounting department. If the accounting department
              calculates $1500/per year per developer for its IT cost, with
              Flatfeestack it is $120 (12 month x $10) more: $1620. And the
              developer can support as many projects as you want.
            </div>
          </div>
        </div>
      </div>
      <div class="flex flex-wrap">
        <div class="w-full md:w-1/2">
          <div class="container">
            <div class="subtitle">Simplified flat-fee approach</div>
            <Spacer x3 />
            <div>
              Each project receives your donation equally divided. Simply donate
              and don't worry about how many projects you want to support, it is
              always $10 per month. It is a similar concept to Flattr for open
              source projects, where git commits decide how much each
              contributor earns.
            </div>
          </div>
        </div>
        <div
          class="flex w-full md:w-1/2 order-first md:order-none justify-center mb-5"
        >
          <img
            class="object-contain w-4/5 h-40"
            src="assets/images/flat3.svg"
            alt="Support"
          />
        </div>
      </div>
      <Spacer x5 />
      <Spacer x5 />
    </div>
  </section>

  <section class="mt-20 bg-white">
    <div class="container mx-auto px-3 lg:px-0">
      <div class="text-lg text-center font-light text-gray-600">
        Flatfeestack is supported by:
      </div>
      <div class="flex flex-wrap">
        <a
          href="https://axlabs.com"
          class="w-full md:w-1/3 flex justify-center grey"
        ><img
            class="object-contain w-full h-full"
            src="assets/images/axlabs.png"
            alt="AxLabs"
          /></a>
        <a
          href="https://axelra.com"
          class="w-full md:w-1/3 flex justify-center grey2"
        ><img
            class="object-contain w-full h-full"
            src="assets/images/axelra.svg"
            alt="Axelra"
          /></a>
        <a
          href="https://coinblesk.com"
          class="w-full md:w-1/3 flex justify-center grey2"
        ><img
            class="object-contain w-full h-full"
            src="assets/images/coinblesk.svg"
            alt="Coinblesk"
          /></a>
      </div>
    </div>
  </section>

</div>
