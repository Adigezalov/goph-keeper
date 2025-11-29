import { observer } from 'mobx-react-lite'
import { ReactNode, useEffect } from 'react'

import { AuthPage } from '@pages/auth-page'

import { StoreContextLogic, TStoreLogic, useStoreLogic } from '@shared/store'
import { PageLoader } from '@shared/ui'

type TProps = {
	children: ReactNode
}

export const AuthProvider = observer(({ children }: TProps) => {
	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)

	const { auth, isCheckLoading, initAuthStore } = store.auth

	useEffect(() => {
		initAuthStore()
	}, [initAuthStore])

	if (isCheckLoading) return <PageLoader />

	if (!auth) return <AuthPage />

	return children
})
