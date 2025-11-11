import { ReactNode } from 'react'

import { AuthStore } from '@entities/auth/models'

import { InitStoreLogic, StoreContextLogic, TStoreLogic } from '@shared/store'
import { useToastNotification } from '@shared/toast-notification'

type Props = {
	children: ReactNode
}

export const StoreProvider = ({ children }: Props) => {
	const toastNotification = useToastNotification()

	const logic = InitStoreLogic<TStoreLogic>([
		{ toastNotification: toastNotification },
		{ auth: AuthStore },
	])

	return <StoreContextLogic.Provider value={logic}>{children}</StoreContextLogic.Provider>
}
