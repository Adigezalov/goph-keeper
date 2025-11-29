/* eslint-env serviceworker */
/* eslint-disable no-undef */
const CACHE_NAME = 'goph-keeper-v1'

const urlsToCache = ['/', '/index.html']

self.addEventListener('install', (event) => {
	console.log('[Service Worker] Установка...')

	event.waitUntil(
		caches.open(CACHE_NAME).then((cache) => {
			console.log('[Service Worker] Кэширование ресурсов')
			return cache.addAll(urlsToCache)
		}),
	)

	self.skipWaiting()
})

self.addEventListener('activate', (event) => {
	console.log('[Service Worker] Активация...')

	event.waitUntil(
		caches.keys().then((cacheNames) => {
			return Promise.all(
				cacheNames.map((cacheName) => {
					if (cacheName !== CACHE_NAME) {
						console.log('[Service Worker] Удаление старого кэша:', cacheName)
						return caches.delete(cacheName)
					}
				}),
			)
		}),
	)

	self.clients.claim()
})

self.addEventListener('fetch', (event) => {
	const { request } = event

	if (!request.url.startsWith('http')) {
		return
	}

	event.respondWith(
		caches.match(request).then((cachedResponse) => {
			if (cachedResponse) {
				return cachedResponse
			}

			return fetch(request).then((response) => {
				return response
			})
		}),
	)
})
