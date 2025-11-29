import { observer } from 'mobx-react-lite'
import { useEffect } from 'react'

import { StoreContextLogic, TStoreLogic, useStoreLogic } from '@shared/store'

import { ServerStatusView } from './server-status.view'

export const ServerStatus = observer(() => {
	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)

	const { status, startStatusCheck, stopStatusCheck } = store.serverStatus

	useEffect(() => {
		startStatusCheck()

		return () => {
			stopStatusCheck()
		}
	}, [startStatusCheck, stopStatusCheck])

	return <ServerStatusView status={status} />
})
