import { lazy } from 'react'

import { LazyLoader } from '@shared/ui'

const AuthPageLazy = lazy(() =>
	import('@pages/auth-page').then((module) => ({
		default: module.AuthPage,
	})),
)

const SecretsPageLazy = lazy(() =>
	import('@pages/secrets-page').then((module) => ({
		default: module.SecretsPage,
	})),
)

export const AuthPage = () => <LazyLoader Component={AuthPageLazy} />
export const SecretsPage = () => <LazyLoader Component={SecretsPageLazy} />
