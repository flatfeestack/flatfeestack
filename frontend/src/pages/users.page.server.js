import { escapeInject, dangerouslySkipEscape } from "vite-plugin-ssr/server";

export { render };
export { passToClient };

//https://github.com/ryanweal/vite-plugin-ssr-svelte
//https://github.com/jiangfengming/svelte-vite-ssr
// See https://vite-plugin-ssr.com/data-fetching
const passToClient = ["pageProps", "routeParams"];

async function render(pageContext) {
  pageContext.showEmptyUser = "true";
  const app = pageContext.Page.render(pageContext);
  const appHtml = app.html;
  const appCss = app.css.code;
  const appHead = app.head;

  // We are using Svelte's app.head variable rather than the Vite Plugin SSR
  // technique described here: https://vite-plugin-ssr.com/html-head This seems
  // easier for using data fetched from APIs and also allows us to input the
  // data using our custom MetaTags Svelte component.

  return escapeInject`<!DOCTYPE html>
      <html lang="en">
        <head>
          <meta charset="UTF-8" />
          <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png" />
          <link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png" />
          <link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png" />
          <link rel="manifest" href="/site.webmanifest" />
          <link rel="mask-icon" href="/safari-pinned-tab.svg" color="#5bbad5" />
          <meta name="msapplication-TileColor" content="#da532c" />
          <meta name="theme-color" content="#ffffff" />
          <meta name="viewport" content="width=device-width, initial-scale=1.0" />
          <meta
            http-equiv="Content-Security-Policy"
            content="default-src 'self';
              connect-src 'self' https://api.stripe.com; frame-src 'self' https://js.stripe.com https://hooks.stripe.com;
              script-src 'self' 'unsafe-inline' https://js.stripe.com; img-src 'self' data: https://*.stripe.com; 
              font-src 'self' fonts.gstatic.com; style-src 'self' 'unsafe-inline' fonts.googleapis.com"
          />
          <link rel="preconnect" href="https://fonts.gstatic.com">
          <link rel="stylesheet" type="text/css" href="https://fonts.googleapis.com/css2?family=Roboto:wght@300;400;500&display=swap" >
            ${dangerouslySkipEscape(appHead)}
            <style>${appCss}</style>
          <title>FlatFeeStack</title>
        </head>
        <body>
          <div id="app">${dangerouslySkipEscape(appHtml)}</div>
        </body>
      </html>`;
}
