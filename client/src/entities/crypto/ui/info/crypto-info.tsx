import { observer } from 'mobx-react-lite'

import { StoreContextLogic, TStoreLogic, useStoreLogic } from '@shared/store'

import { CryptoInfoView } from './crypto-info.view'

export const CryptoInfo = observer(() => {
	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)

	const { cryptoKey, setVisibleCryptoModal } = store.cryptoKey

	const onShowCryptoModal = () => {
		setVisibleCryptoModal(true)
	}

	return (
		<CryptoInfoView
			isCryptoKeySuccess={!!cryptoKey}
			onShowCryptoModal={onShowCryptoModal}
		/>
	)
})
