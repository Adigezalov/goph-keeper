import { JSX, Suspense } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'

import { type TBook, bookPrivate, bookPublic } from '@shared/book'
import { PageLoader } from '@shared/ui'

import { PrivateLayout, PublicLayout } from '../layouts'

export const AppRouter = () => {
	const privateRootPages = bookPrivate.filter((book) => book.isRoot)[0]
	const privatePages = bookPrivate.filter((book) => book.isShow)
	const publicPages = bookPublic.filter((book) => book.isShow)

	const createRoutes = (pages: TBook[], basePath = '') => {
		const routes: JSX.Element[] = []

		pages.forEach((route: TBook) => {
			if (route.externalUrl) return

			const fullPath = basePath + route.path

			if (route.page) {
				routes.push(
					<Route
						key={`${route.name}-${fullPath}`}
						path={fullPath}
						element={route.page}
					/>,
				)
			}

			if (route.pages) {
				const subRoutes = createRoutes(route.pages, fullPath)

				routes.push(...subRoutes)
			}
		})

		return routes
	}

	return (
		<Suspense fallback={<PageLoader />}>
			<Routes>
				<Route element={<PrivateLayout />}>
					{createRoutes(privatePages)}
					<Route path={'*'} element={<Navigate to={privateRootPages.path} />} />
				</Route>
				<Route element={<PublicLayout />}>
					{publicPages.map((route: TBook) => {
						return (
							<Route
								key={`${route.name}-${route.path}`}
								path={route.path}
								element={route.page}
							/>
						)
					})}
				</Route>
			</Routes>
		</Suspense>
	)
}
