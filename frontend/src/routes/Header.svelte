<script>
  import { faHome } from "@fortawesome/free-solid-svg-icons";
  import Fa from "svelte-fa";

  import favicon from "$lib/images/favicon.svg";
  import ffsLogo from "$lib/images/ffs-logo.svg";

  const user = { id: 1, image: "", email: "test" };

  function logout() {
    console.log("logout");
  }
</script>

<style>
  header {
    padding: 1em;
    background-color: #fff;
    border-bottom: 1px #000 solid;
    justify-content: space-between;
    flex: 0 0 auto;
  }

  header,
  nav {
    display: flex;
    align-items: center;
    font-size: 1.1rem;
  }

  nav a {
    border-radius: 5px;
    padding: 0.5rem 0.75rem 0.5rem 0.75rem;
    text-decoration: none;
    cursor: pointer;
    font-size: 1.05rem;
  }

  header :global(a),
  header :global(a:visited),
  header :global(a:active) {
    text-decoration: none;
    color: #000;
  }

  .imgSmallLogo {
    padding-right: 0.25em;
    width: 3rem;
  }
  .imgNormalLogo {
    padding-right: 0.25em;
    width: 10rem;
  }
</style>

<header>
  <a href="/">
    <img class="hide-mda imgSmallLogo" src={favicon} alt="FlatFeeStack" />
    <img class="hide-sx imgNormalLogo" src={ffsLogo} alt="FlatFeeStack" />
  </a>
  <nav>
    {#if user.id}
      <a href="/user/search"><Fa icon={faHome} size="sm" class="icon" /></a>
      {#if user.image}
        <img class="image-org-sx" src={user.image} alt="user profile img" />
      {/if}
      {user.email}
      <form on:submit|preventDefault={logout}>
        <button class="button3 center mx-2" type="submit">Sign out</button>
      </form>
    {:else}
      <a href="/login" class="button3 center mx-2">Login</a>
      <a href="/signup" class="button1 center mx-2" type="submit">Sign Up</a>
    {/if}
    <a href="/dao/home" class="button4 center mx-2">DAO</a>
  </nav>
</header>
