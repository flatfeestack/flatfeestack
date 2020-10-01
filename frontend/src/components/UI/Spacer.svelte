<script>
    export let x1;
    export let x2;
    export let x3;
    export let x4;
    export let x5;
</script>
<style>
    .x1{
        min-width: var(--spacing-x1);
        min-height: var(--spacing-x1);
    }
    .x2{
        min-width: var(--spacing-x2);
        min-height: var(--spacing-x2);
    }
    .x3{
        min-width: var(--spacing-x3);
        min-height: var(--spacing-x3);
    }
    .x4{
        min-width: var(--spacing-x4);
        min-height: var(--spacing-x4);
    }
    .x5{
         min-width: var(--spacing-x5);
         min-height: var(--spacing-x5);
    }
    .base{
        position: relative;
        display: flex;
    }
</style>

<div class="base {x1?'x1':''} {x2?'x2':''} {x3?'x3':''} {x4?'x3':''} {x4?'x4':''} {x5?'x5':''}"></div>
