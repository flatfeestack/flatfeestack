<script>import { onMount, afterUpdate, onDestroy } from 'svelte';
import { Chart as ChartJS } from 'chart.js';
import { useForwardEvents } from '../utils/index.js';
function clean(props) {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    let { data, type, options, plugins, children, $$scope, $$slots, ...rest } = props;
    return rest;
}
export let type;
export let data = {
    datasets: [],
};
export let options = {};
export let plugins = [];
export let updateMode = undefined;
export let chart = null;
let canvasRef;
let props = clean($$props);
onMount(() => {
    chart = new ChartJS(canvasRef, {
        type,
        data,
        options,
        plugins,
    });
});
afterUpdate(() => {
    if (!chart)
        return;
    chart.data = data;
    Object.assign(chart.options, options);
    chart.update(updateMode);
});
onDestroy(() => {
    if (chart)
        chart.destroy();
    chart = null;
});
useForwardEvents(() => canvasRef);
</script>

<canvas bind:this={canvasRef} {...props} />
