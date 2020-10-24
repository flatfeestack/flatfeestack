<script>
import Navigation from "./Navigation.svelte";
import Footer from "./Footer.svelte";
</script>

<div class="page min-h-screen flex flex-col">
  <Navigation />
  <div class="container mx-auto px-3 lg:px-0">
    <slot />
  </div>
  <div class="flex-1"></div>
  <Footer />
</div>
