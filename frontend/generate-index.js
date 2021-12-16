fs = require('fs');
//https://www.base2.io/2020/12/12/svelte-ssr
const App = require("./public/server/ssr.js")

const { html } = App.render({})

const htmlStart = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset='utf-8'>
  <meta name='viewport' content='width=device-width,initial-scale=1'>
  <title>FlatFeeStack</title>
  <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg'/%3E">
  <link rel="stylesheet" type="text/css" href="/build/bundle.css">
  <link rel="preconnect" href="https://fonts.gstatic.com">
  <link rel="stylesheet" type="text/css" href="https://fonts.googleapis.com/css2?family=Roboto:wght@300;400;500&display=swap" > 
  <script type="module" defer src='/build/main.js'></script>
  <script defer src="https://js.stripe.com/v3/"></script>
  
</head>
<body>`;

const htmlEnd = `</body>
</html>`;

fs.writeFileSync('public/index.html', htmlStart+html+htmlEnd);
console.log("created index.html")
