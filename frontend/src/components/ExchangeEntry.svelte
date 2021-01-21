<script lang="ts">
import type { Exchange } from "../types/exchange.type";
import { API } from "../api/api";
import { faEthereum } from "@fortawesome/free-brands-svg-icons";
import { faLongArrowAltRight } from "@fortawesome/free-solid-svg-icons";
import Fa from "svelte-fa";
import { onMount } from "svelte";
import format from "date-fns/format";
import parseISO from "date-fns/parseISO";

export let exchange: Exchange;
let error: string;

let date = new Date().toISOString();
let completed = false;

async function update(e) {
  console.log("in update");
  e.preventDefault();
  try {
    const res = await API.exchanges.update({
      ...exchange,
      date,
      price: parseFloat(exchange.price).toFixed(4),
    });
    if (res.data.success === true) {
      exchange = {
        ...exchange,
        date,
        price: parseFloat(exchange.price).toFixed(4),
      };
      console.log("new exchange", exchange);
      completed = true;
    }
  } catch (e) {
    console.log(e);
    error = String(e);
  }
}
onMount(() => {
  completed = !!exchange.date;
});
</script>

<div class="flex">
  <div
    class="flex flex-col p-5 shadow-lg mb-5 {completed ? 'bg-green-500 text-white' : ''}"
  >
    <div class="flex flex-row items-center text-lg mb-2">
      {(parseFloat(exchange.amount) / 100).toFixed(2)}$

      <div class="pl-5">
        <Fa icon="{faLongArrowAltRight}" />
      </div>
      <div class="pl-5">
        <Fa icon="{faEthereum}" />
      </div>
      <div class="flex-1"></div>
      {#if completed}
        <div>{format(parseISO(exchange.date), 'dd.MM.yyyy')}</div>
      {/if}
    </div>
    <form class="flex items-end" on:submit="{update}">
      <div>
        <label
          for="password-input"
          class="text-sm {completed ? 'text-white' : ''} text-opacity-75"
        >Exchange rate in USD</label>
        <input
          type="number"
          step="0.0001"
          class="input"
          disabled="{completed}"
          bind:value="{exchange.price}"
        />
      </div>
      <div>
        <button
          type="submit"
          class="button ml-5"
          disabled="{completed}"
        >Update</button>
      </div>
    </form>
  </div>
</div>
