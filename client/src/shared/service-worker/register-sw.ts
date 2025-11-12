import i18next from 'i18next'

import { showToastNotification } from '@shared/toast-notification'
import { TOAST_SEVERITY } from '@shared/uikit/toast'

export const registerServiceWorker = async () => {
	if ('serviceWorker' in navigator) {
		try {
			const registration = await navigator.serviceWorker.register('/service-worker.js')
			console.log('[SW] Service Worker зарегистрирован:', registration.scope)

			// Проверяем обновления каждые 60 секунд
			setInterval(() => {
				registration.update()
			}, 60000)

			// Обработка обновления Service Worker
			registration.addEventListener('updatefound', () => {
				const newWorker = registration.installing
				if (newWorker) {
					newWorker.addEventListener('statechange', () => {
						if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
							console.log('[SW] Доступно обновление приложения')

							showToastNotification({
								message: i18next.t('new_version_service'),
								header: i18next.t('info'),
								severity: TOAST_SEVERITY.INFO,
							})
						}
					})
				}
			})
		} catch (error) {
			console.error('[SW] Ошибка регистрации Service Worker:', error)
		}
	} else {
		console.warn('[SW] Service Worker не поддерживается браузером')
	}
}

export const unregisterServiceWorker = async () => {
	if ('serviceWorker' in navigator) {
		try {
			const registrations = await navigator.serviceWorker.getRegistrations()
			for (const registration of registrations) {
				await registration.unregister()
				console.log('[SW] Service Worker удален')
			}
		} catch (error) {
			console.error('[SW] Ошибка удаления Service Worker:', error)
		}
	}
}
