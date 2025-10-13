import {goto} from "@mateothegreat/svelte5-router";

const showMetaMaskRequired = () => goto("/dao/metamask");
export default showMetaMaskRequired;
