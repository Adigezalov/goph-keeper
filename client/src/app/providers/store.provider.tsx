import { ReactNode } from 'react'

import { AuthStore } from '@entities/auth/models'
import { CryptoStore } from '@entities/crypto/models'
import { NetworkStatusStore } from '@entities/network-status/models'
import { RealtimeStore } from '@entities/realtime/models'
import { ServerStatusStore } from '@entities/server-status/models'

import { SecretsPageStore } from '@pages/secrets-page/models'

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
		{ cryptoKey: CryptoStore },
		{ networkStatus: NetworkStatusStore },
		{ realtime: RealtimeStore },
		{ secretsPage: SecretsPageStore },
		{ serverStatus: ServerStatusStore },
	])

	return <StoreContextLogic.Provider value={logic}>{children}</StoreContextLogic.Provider>
}
