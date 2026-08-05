self.addEventListener('fetch', (event) => {
    // NOVA REGRA: Se for uma chamada para a API, o Service Worker não se mete!
    if (event.request.url.includes('/api/')) {
        return; // Deixa o navegador seguir o fluxo normal pela internet
    }
});

// my-app/www/sw.js
self.addEventListener('install', (e) => {
  console.log('[Service Worker] Instalado');
});

self.addEventListener('fetch', (e) => {
  e.respondWith(fetch(e.request));
});