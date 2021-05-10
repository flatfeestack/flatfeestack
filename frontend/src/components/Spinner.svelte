<style>
    .loader {
        border: 0.3rem solid var(--secondary-100);
        border-top: 0.3rem solid var(--primary-900);
        border-radius: 50%;
        width: 2rem;
        height: 2rem;
        animation: spin 2s linear infinite;
        position: fixed;
        z-index: 999;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
    }

    @keyframes spin {
        0% { transform: rotate(0deg); }
        100% { transform: rotate(360deg); }
    }
</style>

<div class="loader"></div>
