import { PrimeReactProvider } from 'primereact/api'
import { useEffect } from 'react'
import { BrowserRouter } from 'react-router-dom'

import { useRoutePreloader } from '@shared/hooks/use-route-preloader'
import { registerServiceWorker } from '@shared/service-worker'

import {
	LocalizationProvider,
	RealtimeProvider,
	StoreProvider,
	StyleProvider,
	ToastNotificationProvider,
} from './providers'
import { AppRouter } from './router'

export const App = () => {
	useRoutePreloader({
		routes: [() => import('@pages/auth-page'), () => import('@pages/secrets-page')],
		delay: 3000,
		priority: 'low',
	})

	useEffect(() => {
		void registerServiceWorker()
	}, [])

	return (
		<StyleProvider>
			<LocalizationProvider>
				<BrowserRouter>
					<StoreProvider>
						<PrimeReactProvider>
							<ToastNotificationProvider>
								<RealtimeProvider>
									<AppRouter />
								</RealtimeProvider>
							</ToastNotificationProvider>
						</PrimeReactProvider>
					</StoreProvider>
				</BrowserRouter>
			</LocalizationProvider>
		</StyleProvider>
	)
}
