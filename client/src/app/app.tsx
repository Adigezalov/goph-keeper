import { PrimeReactProvider } from 'primereact/api'
import { BrowserRouter } from 'react-router-dom'

import { useRoutePreloader } from '@shared/hooks/use-route-preloader'

import {
	LocalizationProvider,
	StoreProvider,
	StyleProvider,
	ToastNotificationProvider,
} from './providers'
import { AppRouter } from './router'

export const App = () => {
	useRoutePreloader({
		routes: [() => import('@pages/auth-page'), () => import('@pages/dashboard-page')],
		delay: 3000,
		priority: 'low',
	})

	return (
		<StyleProvider>
			<LocalizationProvider>
				<BrowserRouter>
					<StoreProvider>
						<PrimeReactProvider>
							<ToastNotificationProvider>
								<AppRouter />
							</ToastNotificationProvider>
						</PrimeReactProvider>
					</StoreProvider>
				</BrowserRouter>
			</LocalizationProvider>
		</StyleProvider>
	)
}
