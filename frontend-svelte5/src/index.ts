import { hydrate } from 'svelte';
import App from './Index.svelte';
import './index.css';

hydrate(App, {
    target: document.getElementById('root')!,
    props: {},
});