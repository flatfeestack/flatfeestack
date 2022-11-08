import "../app.css";

export { render };

//export const clientRouting = true

async function render(pageContext) {
  const app_el = document.getElementById("app");
  new pageContext.Page({
    target: app_el,
    hydrate: true,
    props: {
      pageProps: pageContext.pageProps,
    },
  });
}
