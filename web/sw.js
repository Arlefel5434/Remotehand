// The controller requires a live Agent. Keep pairing URLs and replies out of caches.
self.addEventListener("fetch", event => {
  if (event.request.mode !== "navigate" || event.request.method !== "GET") return;
  event.respondWith(fetch(event.request).catch(() => new Response(
    '<!doctype html><html lang="ru"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Remotehand</title><h1>Ноутбук недоступен</h1><p>Проверьте сеть и запустите Remotehand на ноутбуке.</p><a href="/">Повторить подключение</a></html>',
    {status: 503, headers: {"Content-Type": "text/html; charset=utf-8", "Cache-Control": "no-store"}}
  )));
});
