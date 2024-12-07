<script lang="ts">
    import {goto} from "@mateothegreat/svelte5-router";
    import {onDestroy, onMount} from "svelte";
    import { initAnimation } from './backgroundAnimation';
    import HeaderLanding from "./HeaderLanding.svelte";

    let cleanup: (() => void) | undefined;

    onMount(() => {
        cleanup = initAnimation();
    });

    onDestroy(() => {
        if (cleanup) cleanup();
    });
</script>

<style>
    svg {
        width: 100%;
        height: 100%;
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

    .flex-row {
        display: flex;
        flex-direction: row;
        justify-content: center;
    }

    main {
        width: 100%;
    }

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

    .flex-col {
        display: flex;
        flex-direction: column;
    }

    .parallax {
        text-align: center;
        justify-content: center;
        background-image: url(/images/code-dark2.webp);
        color: #fff;
        background-attachment: fixed;
        background-position: center;
        background-repeat: no-repeat;
        background-size: cover;
        box-shadow: 0 0 5px 5px var(--secondary-200);
        margin-bottom: 1em;
    }

    .parallax2 {
        display: flex; /* Add this to use Flexbox */
        flex-direction: column; /* Stack children vertically */
        justify-content: center; /* Center children horizontally */
        align-items: center; /* Center children vertically */
        color: #fff;
        background-attachment: fixed;
        background-position: center;
        background-repeat: no-repeat;
        background-size: cover;
        box-shadow: 0 0 5px 5px var(--secondary-200);
        height: 32em;
    }

    .background-svg {
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 32em;
        z-index: -2; /* Ensure it stays behind other content */
    }

    .background-svg2 {
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 32em; /* Adjust height as needed */
        background-color: rgba(0, 0, 0, 0.6); /* Dark overlay with 50% opacity */
        z-index: -1; /* Ensure it stays behind other content but on top of the SVG */
    }

</style>


<main>
    <HeaderLanding></HeaderLanding>
    <div class="background-svg">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="-40 -250 800 900" id="floor" style="width: 100%">
        <g id="roll2-part2" transform="translate(425,-39)">
            <path d="M55.086 67.746c-4.089.086-8.675 1.458-13.518 4.254-17.218 9.941-31.175 34.118-31.175 54 0 9.942 3.49 16.929 9.132 20.186L268.068 289.64l44.96-75.827L64.68 70.494c-2.69-1.875-5.944-2.824-9.594-2.748"
                  style="fill:#939393"/>
            <path d="M290.985 288c-17.219 9.941-31.177 1.882-31.177-18s13.958-44.059 31.177-54c17.218-9.941 31.176-1.882 31.176 18s-13.958 44.059-31.176 54"
                  style="fill:#b9b9b9"/>
        </g>

        <g id="scrollingFloor"></g>

        <g transform="translate(425,-39)">
            <path d="M20.785 108V36L0 24v72z" style="fill:#a3a3a3"/>
            <path d="M62.354 84V12l-41.57 24v72z" style="fill:#e1e1e1"/>
        </g>

        <g id="dots"></g>

        <g id="press1" transform="translate(-40,230)">
            <path d="M55.086 67.746c-4.089.086-8.675 1.458-13.518 4.254-17.218 9.941-31.175 34.118-31.175 54 0 9.942 3.49 16.929 9.132 20.186L268.068 289.64l44.96-75.827L64.68 70.494c-2.69-1.875-5.944-2.824-9.594-2.748"
                  style="fill:#939393"/>
            <path d="M290.985 288c-17.219 9.941-31.177 1.882-31.177-18s13.958-44.059 31.177-54c17.218-9.941 31.176-1.882 31.176 18s-13.958 44.059-31.176 54"
                  style="fill:#b9b9b9"/>
            <path d="M20.785 108V36L0 24v72z" style="fill:#a3a3a3"/>
            <path d="M62.354 84V12l-41.57 24v72z" style="fill:#e1e1e1"/>
            <path d="m20.785 36 41.569-24 249.415 144v24l-41.57 24v-24z" style="fill:#939393"/>
            <path d="m20.785 60 311.769 180v-24L20.784 36Z" style="fill:#606060"/>
            <path d="m290.985 192 41.569-24-20.785-12-41.57 24z" style="fill:#bebebe"/>
            <path d="M290.985 264v-72L270.2 180v72z" style="fill:#a3a3a3"/>
            <path d="m332.555 168-41.57 24v72l15.587-9v24l10.393-6v-24l15.59-9z" style="fill:#e1e1e1"/>
            <path d="m20.785 36 41.569-24L41.569 0 0 24Z" style="fill:#bebebe"/>
            <path d="m306.573 279-20.785-12v-6l5.197 3 15.588-9z" style="fill:#a3a3a3"/>
        </g>

        <g transform="translate(425,-39)">
            <path d="m20.785 36 41.569-24 249.415 144v24l-41.57 24v-24z" style="fill:#939393"/>
            <path d="m20.785 60 311.769 180v-24L20.784 36Z" style="fill:#606060"/>
            <path d="m290.985 192 41.569-24-20.785-12-41.57 24z" style="fill:#bebebe"/>
            <path d="M290.985 264v-72L270.2 180v72z" style="fill:#a3a3a3"/>
            <path d="m332.555 168-41.57 24v72l15.587-9v24l10.393-6v-24l15.59-9z" style="fill:#e1e1e1"/>
            <path d="m20.785 36 41.569-24L41.569 0 0 24Z" style="fill:#bebebe"/>
            <path d="m306.573 279-20.785-12v-6l5.197 3 15.588-9z" style="fill:#a3a3a3"/>
        </g>
      </svg>
    </div>
    <div class="background-svg2"></div>

      <div class="flex-col">
        <div class="parallax2">
            <img
                    class="hide-sx logoFFSWhiteLarge"
                    src="/images/ffs-logo-white.svg"
                    alt="FlatFeeStack Logo"
            />
            <h1>
                A flat fee sponsoring platform for <br />
                open-source projects based on git metrics
            </h1>
        </div>
      </div>


    <div class="flex-row py-300">
        <div class="card">
            <div>
                <img
                        class="icon bottom"
                        src="/images/holding-world-bw.svg"
                        alt="Hands holding world icon bw"
                />
                <img
                        class="icon top"
                        src="/images/holding-world.svg"
                        alt="Hands holding world icon color"
                />
            </div>
            <h2>On the Shoulders of Giants</h2>
            <div class="max-w20">
                Without open-source software, your software project would not exist
            </div>
        </div>
        <div class="card">
            <div>
                <img
                        class="icon bottom"
                        src="/images/house-bw.svg"
                        alt="House icon bw"
                />
                <img class="icon top" src="/images/house.svg" alt="House icon bw" />
            </div>
            <h2>Our Mission</h2>
            <div class="max-w20">
                &#171;Open-Source contributors should be able to make a living with open
                source software&#187;
            </div>
        </div>
    </div>

    <div class="flex-col">
        <div class="flex-row py-300 parallax">
            <div class="p-300 max-w20 rightBorder">
                <h3>
                    For a flat fee of $125/year, sponsor any open-source projects. Show
                    your supported projects badge.
                </h3>
                <form on:submit|preventDefault={() => goto("/signup")}>
                    <button class="button1">Sign Up</button>
                </form>
            </div>
            <div class="p-300 max-w20">
                <h3>
                    Earn money if your open-source project is sponsored. All you need is a
                    public git repository.
                </h3>
                <form on:submit|preventDefault={() => goto("/signup")}>
                    <button class="button1">Sign Up</button>
                </form>
            </div>
        </div>
    </div>

    <div class="flex-row">
        <div class="flex-row card max-w50 cardLeft">
            <div>
                <img
                        class="icon bottom"
                        src="/images/running-bw.svg"
                        alt="person running icon bw"
                />
                <img
                        class="icon top"
                        src="/images/running.svg"
                        alt="person running icon color"
                />
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
                <img
                        class="icon bottom"
                        src="/images/metrics-bw.svg"
                        alt="metric icon bw"
                />
                <img class="icon top" src="/images/metrics.svg" alt="metric icon bw" />
            </div>
        </div>
    </div>

    <div class="flex-row">
        <div class="flex-row card max-w50 cardLeft">
            <div>
                <img class="icon bottom" src="/images/seat-bw.svg" alt="seat icon bw" />
                <img class="icon top" src="/images/seat.svg" alt="metric icon color" />
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
                <img
                        class="icon bottom"
                        src="/images/cc-bw.svg"
                        alt="crypto currencies icon bw"
                />
                <img
                        class="icon top"
                        src="/images/cc.svg"
                        alt="crypto currencies icon bw"
                />
            </div>
        </div>
    </div>
</main>
