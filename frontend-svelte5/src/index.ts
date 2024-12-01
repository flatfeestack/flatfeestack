import { hydrate } from 'svelte';
import App from './Index.svelte';
import 'css/index.css';
import 'css/space.css';
import 'css/text.css';
import 'css/border.css';
import 'css/button.css';
import 'css/input.css';

hydrate(App, {
    target: document.getElementById('root')!,
    props: {},
});