<script lang="ts">
  export let name: string;
</script>

<style>
  main {
    text-align: center;
    padding: 1em;
    max-width: 240px;
    margin: 0 auto;
  }

  @media (min-width: 640px) {
    main {
      max-width: none;
    }
  }
</style>

<main>
  <div class="container">
    <h1 class="title is-1">j0nezz Svelte Starter</h1>
    <p class="subtitle is-4">
      <strong>Svelte</strong>🚀 + Typescript + Sass + <a
        href="https://bulma.io/">Bulma</a>
    </p>
    <section class="section">
      <div class="subtitle">
        Customize colors in the <span class="tag"> styles.scss </span> file using
        Sass variables
      </div>
      <span class="tag is-black">Black</span>
      <span class="tag is-dark">Dark</span>
      <span class="tag is-light">Light</span>
      <span class="tag is-white">White</span>
      <span class="tag is-primary">Primary</span>
      <span class="tag is-link">Link</span>
      <span class="tag is-info">Info</span>
      <span class="tag is-success">Success</span>
      <span class="tag is-warning">Warning</span>
      <span class="tag is-danger">Danger</span>
    </section>
  </div>
</main>
<svelte:head>
  <style src="styles.scss">
  </style>
</svelte:head>
