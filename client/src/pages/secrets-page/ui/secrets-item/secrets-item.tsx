import { observer } from 'mobx-react-lite'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'

import { StoreContextLogic, TStoreLogic, useStoreLogic } from '@shared/store'

import { TSecret, TSecretForSave } from '../../types'
import { SecretsItemView } from './secrets-item.view.tsx'

type Props = {
	secret: TSecret
}

export const SecretsItem = observer(({ secret }: Props) => {
	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)

	const { updateSecret, decryptPassword, deleteSecret } = store.secretsPage

	const [decryptedPassword, setDecryptedPassword] = useState('')

	const { control, handleSubmit, reset, watch } = useForm<TSecretForSave>({
		mode: 'all',
		defaultValues: {
			login: '',
			password: '',
			metadata: '',
		},
	})

	const formValues = watch()

	useEffect(() => {
		const loadDecryptedPassword = async () => {
			try {
				const decrypted = await decryptPassword(secret.password)
				setDecryptedPassword(decrypted)
				reset({
					login: secret.login,
					password: decrypted,
					metadata: secret.metadata || '',
				})
			} catch (error) {
				console.error('Ошибка расшифровки пароля:', error)
			}
		}

		void loadDecryptedPassword()
	}, [secret.password, secret.login, secret.metadata, decryptPassword, reset])

	const disabledSave = (): boolean => {
		const isLoginSame = formValues.login === secret.login
		const isPasswordSame = formValues.password === decryptedPassword
		const isMetadataSame = (formValues.metadata || '') === (secret.metadata || '')

		return isLoginSame && isPasswordSame && isMetadataSame
	}

	const onDelete = () => {
		void deleteSecret(secret.localId)
	}

	const onSubmit = async (data: TSecretForSave) => {
		try {
			await updateSecret(secret.localId, data)
		} catch (error) {
			// Ошибка уже обработана в store
		}
	}

	const onSave = handleSubmit(onSubmit)

	return (
		<SecretsItemView<TSecretForSave>
			control={control}
			onSave={onSave}
			onDelete={onDelete}
			disabled={disabledSave()}
			isEditMode
		/>
	)
})
