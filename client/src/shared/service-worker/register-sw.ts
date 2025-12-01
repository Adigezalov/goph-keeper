import i18next from 'i18next'

import { showToastNotification } from '@shared/toast-notification'
import { TOAST_SEVERITY } from '@shared/uikit/toast'

export const registerServiceWorker = async () => {
	if ('serviceWorker' in navigator) {
		try {
			const registration = await navigator.serviceWorker.register('/service-worker.js')
			console.log(i18next.t('service_worker.registered'), registration.scope)

			setInterval(() => {
				registration.update()
			}, 60000)

			registration.addEventListener('updatefound', () => {
				const newWorker = registration.installing
				if (newWorker) {
					newWorker.addEventListener('statechange', () => {
						if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
							console.log(i18next.t('service_worker.update_available'))

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
			console.error(i18next.t('service_worker.registration_error'), error)
		}
	} else {
		console.warn(i18next.t('service_worker.not_supported'))
	}
}

export const unregisterServiceWorker = async () => {
	if ('serviceWorker' in navigator) {
		try {
			const registrations = await navigator.serviceWorker.getRegistrations()
			for (const registration of registrations) {
				await registration.unregister()
				console.log(i18next.t('service_worker.removed'))
			}
		} catch (error) {
			console.error(i18next.t('service_worker.removal_error'), error)
		}
	}
}
