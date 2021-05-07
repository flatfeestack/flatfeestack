fs = require('fs');
//https://www.base2.io/2020/12/12/svelte-ssr
const App = require("./public/server/ssr.js")

const { html} = App.render({})

const htmlStart = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset='utf-8'>
  <meta name='viewport' content='width=device-width,initial-scale=1'>
  <title>Flatfeestack</title>
  <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg'/%3E">
  <link rel='stylesheet' href='/build/bundle.css'>
  <script type="module" defer src='/build/main.js'></script>
  <script defer src="https://js.stripe.com/v3/"></script>
</head>
<body>`;

const htmlEnd = `</body>
</html>`;

fs.writeFileSync('public/index.html', htmlStart+html+htmlEnd);
