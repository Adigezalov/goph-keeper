import { observer } from 'mobx-react-lite'
import { useEffect } from 'react'

import { StoreContextLogic, TStoreLogic, useStoreLogic } from '@shared/store'

import { SecretsPageView } from './secrets-page.view.tsx'

export const SecretsPage = observer(() => {
	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)

	const { secrets, initStore } = store.secretsPage

	useEffect(() => {
		void initStore()
	}, [initStore])

	return <SecretsPageView secrets={secrets} />
})
