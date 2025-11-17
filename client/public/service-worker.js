const CACHE_NAME = 'goph-keeper-v1'

// Ресурсы для кэширования
const urlsToCache = ['/', '/index.html']

// Установка Service Worker
self.addEventListener('install', (event) => {
	console.log('[Service Worker] Установка...')

	event.waitUntil(
		caches.open(CACHE_NAME).then((cache) => {
			console.log('[Service Worker] Кэширование ресурсов')
			return cache.addAll(urlsToCache)
		}),
	)

	// Активируем новый Service Worker сразу
	self.skipWaiting()
})

// Активация Service Worker
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

	// Берем контроль над всеми клиентами
	self.clients.claim()
})

// Обработка запросов (Offline-First стратегия)
self.addEventListener('fetch', (event) => {
	const { request } = event

	// Пропускаем chrome-extension и другие не-http(s) запросы
	if (!request.url.startsWith('http')) {
		return
	}

	event.respondWith(
		caches.match(request).then((cachedResponse) => {
			// Если ресурс в кэше - возвращаем его
			if (cachedResponse) {
				return cachedResponse
			}

			// Если нет в кэше - запрашиваем из сети
			return fetch(request).then((response) => {
				// Кэшируем только GET запросы со статусом 200
				// if (
				// 	request.method === 'GET' &&
				// 	response.status === 200 &&
				// 	response.type === 'basic'
				// ) {
				// 	const responseToCache = response.clone()
				//
				// 	caches.open(CACHE_NAME).then((cache) => {
				// 		cache.put(request, responseToCache)
				// 	})
				// }

				return response
			})
		}),
	)
})
