<style>
.footer {
  background: var(--primary);
  min-height: var(--footer-height);
  color: white;
}
.footer-padding {
  padding: var(--spacing-x2) 0;
}
</style>

<script>
const date = new Date().getFullYear();
</script>

<footer class="footer">
  <div class="container flex footer-padding">© Flatfeestack {date}</div>
</footer>
