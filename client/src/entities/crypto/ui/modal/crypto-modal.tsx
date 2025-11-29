import { observer } from 'mobx-react-lite'
import { useEffect } from 'react'
import { useForm } from 'react-hook-form'

import { generateCryptoKey } from '@entities/crypto/utils'

import { StoreContextLogic, TStoreLogic, useStoreLogic } from '@shared/store'

import { CryptoModalView } from './crypto-modal.view.tsx'

export const CryptoModal = observer(() => {
	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)

	const { visibleCryptoModal, initCryptoStore, setVisibleCryptoModal, setCryptoKey } =
		store.cryptoKey

	useEffect(() => {
		void initCryptoStore()
	}, [initCryptoStore])

	const { control, handleSubmit, setValue } = useForm<{ key?: string }>({
		mode: 'all',
		defaultValues: {
			key: '',
		},
	})

	const onSubmit = (data: { key?: string }) => {
		void setCryptoKey(data.key)
	}

	const onSave = handleSubmit(onSubmit)

	const onGenerate = async () => {
		const generatedKey = await generateCryptoKey()
		setValue('key', generatedKey)
	}

	const onHide = () => {
		setVisibleCryptoModal(false)
	}

	return (
		<CryptoModalView
			visible={visibleCryptoModal}
			control={control}
			onHide={onHide}
			onSave={onSave}
			onGenerate={onGenerate}
		/>
	)
})
