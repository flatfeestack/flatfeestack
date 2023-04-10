import { goto } from "$app/navigation";

const showMetaMaskRequired = () => goto("/dao/metamask");
export default showMetaMaskRequired;
