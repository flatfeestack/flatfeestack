<script>
    export let label;
    export let secondary;
</script>

<style>
    .btn-base{
        box-shadow: 0 2px 4px 0 rgba(0,0,0,0.15);
        border-radius: 8px;
        font-family: "Open Sans", sans-serif;
        padding: var(--spacing-x2);
    }
    .btn-primary{
        background: #06623B;
        color: #FFFFFF;
    }
    .btn-secondary{
        background: rgba(248,248,248,0.79);
        color: var(--primary);
    }
</style>

<button type="button" class="btn-base {secondary?'btn-secondary':'btn-primary'}">{label}</button>
