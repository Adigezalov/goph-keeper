import { useEffect } from 'react'

interface RoutePreloaderConfig {
	routes: Array<() => Promise<any>>
	delay?: number
	priority?: 'high' | 'low'
}

export const useRoutePreloader = ({
	routes,
	delay = 2000,
	priority = 'low',
}: RoutePreloaderConfig) => {
	useEffect(() => {
		const preloadRoutes = async () => {
			await new Promise((resolve) => setTimeout(resolve, delay))

			if (priority === 'high') {
				await Promise.all(routes.map((route) => route().catch(() => {})))
			} else {
				for (const route of routes) {
					try {
						await route()
						await new Promise((resolve) => setTimeout(resolve, 100))
					} catch {
					}
				}
			}
		}

		if ('requestIdleCallback' in window) {
			requestIdleCallback(() => {
				void preloadRoutes()
			})
		} else {
			setTimeout(preloadRoutes, delay)
		}
	}, [routes, delay, priority])
}
