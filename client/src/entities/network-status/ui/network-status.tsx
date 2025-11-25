import { observer } from 'mobx-react-lite'
import { useEffect } from 'react'

import { StoreContextLogic, TStoreLogic, useStoreLogic } from '@shared/store'

import { NetworkStatusView } from './network-status.view'

export const NetworkStatus = observer(() => {
	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)

	const { isOnline, initNetworkStatusStore, clearNetworkStatusStore } = store.networkStatus

	useEffect(() => {
		initNetworkStatusStore()

		return () => {
			clearNetworkStatusStore()
		}
	}, [initNetworkStatusStore, clearNetworkStatusStore])

	return <NetworkStatusView isOnline={isOnline} />
})
