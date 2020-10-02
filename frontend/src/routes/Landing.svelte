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
  margin-top: calc(3 * var(--spacing-x5));
}
.padding {
  padding-right: calc(2 * var(--spacing-x5));
}
</style>

<script lang="ts">
import Button from "../components/UI/Button.svelte";
import Spacer from "../components/UI/Spacer.svelte";
</script>

<section class="hero-bg">
  <div class="container">
    <div class="flex spacing">
      <div class="flex-1 padding">
        <div class="lead">Show your ♥️ for Open Source Software</div>
        <div class="subtitle">
          Become a Sponsor and support your favorite stack and projects. For a
          flat fee of $5, you can support as many open source project as you
          like.
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
    <div class="flex">
      <Button label="Support Open Source" />
      <Spacer x2 />
      <Button label="Receive Support" secondary />
    </div>
    <Spacer x5 />
    <Spacer x5 />
    <Spacer x5 />
  </div>
</section>
<section class="section-padding">
  <div class="container">
    <div class="section-title">What is it?</div>
    <div class="subtitle">
      Flatfeestack is a Platform to support the software you love.
    </div>
  </div>
</section>
<section class="section-padding">
  <div class="container">
    <div class="section-title">By the way...</div>
    <div class="subtitle">
      We are also completely transparent how the money is split between
      contributors.
    </div>
  </div>
</section>
