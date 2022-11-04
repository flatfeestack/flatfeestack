import { escapeInject, dangerouslySkipEscape } from "vite-plugin-ssr"

export { render }
export { passToClient }

//https://github.com/ryanweal/vite-plugin-ssr-svelte
//https://github.com/jiangfengming/svelte-vite-ssr
// See https://vite-plugin-ssr.com/data-fetching
const passToClient = ['pageProps', 'routeParams']

async function render(pageContext) {
    const app = pageContext.Page.render(pageContext)
    const appHtml = app.html
    const appCss = app.css.code
    const appHead = app.head

    // We are using Svelte's app.head variable rather than the Vite Plugin SSR
    // technique described here: https://vite-plugin-ssr.com/html-head This seems
    // easier for using data fetched from APIs and also allows us to input the
    // data using our custom MetaTags Svelte component.

    return escapeInject`<!DOCTYPE html>
    <html lang="en">
      <head>
        <meta charset="UTF-8" />
        <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg'/%3E">
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
         <link rel="preconnect" href="https://fonts.gstatic.com">
    <link rel="stylesheet" type="text/css" href="https://fonts.googleapis.com/css2?family=Roboto:wght@300;400;500&display=swap" >
        ${dangerouslySkipEscape(appHead)}
        <style>${appCss}</style>
      </head>
      <body>
        <div id="app">${dangerouslySkipEscape(appHtml)}</div>
      </body>
    </html>`
}