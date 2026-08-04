// my-app/www/sw.js
self.addEventListener('install', (e) => {
  console.log('[Service Worker] Instalado');
});

self.addEventListener('fetch', (e) => {
  e.respondWith(fetch(e.request));
});