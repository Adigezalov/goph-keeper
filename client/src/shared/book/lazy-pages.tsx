import { lazy } from 'react'

import { LazyLoader } from '@shared/ui'

const AuthPageLazy = lazy(() =>
	import('@pages/auth-page').then((module) => ({
		default: module.AuthPage,
	})),
)

const DashboardPageLazy = lazy(() =>
	import('@pages/dashboard-page').then((module) => ({
		default: module.DashboardPage,
	})),
)

export const AuthPage = () => <LazyLoader Component={AuthPageLazy} />
export const DashboardPage = () => <LazyLoader Component={DashboardPageLazy} />
