import { observer } from 'mobx-react-lite'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'

import { StoreContextLogic, TStoreLogic, useStoreLogic } from '@shared/store'

import { TSecret, TSecretForSave } from '../../types'
import { extractFileData } from '../../utils'
import { SecretsItemView } from './secrets-item.view.tsx'

type TFormData = Omit<TSecretForSave, 'binaryData'> & {
	binaryData?: File | Uint8Array
}

type Props = {
	secret: TSecret
}

export const SecretsItem = observer(({ secret }: Props) => {
	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)

	const { updateSecret, decryptPassword, deleteSecret } = store.secretsPage

	const [decryptedPassword, setDecryptedPassword] = useState('')

	const { control, handleSubmit, reset, watch } = useForm<TFormData>({
		mode: 'all',
		defaultValues: {
			login: '',
			password: '',
			metadata: {},
			binaryData: undefined,
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
					metadata: secret.metadata || {},
					binaryData: secret.binaryData,
				})
			} catch (error) {
				console.error('Ошибка расшифровки пароля:', error)
			}
		}

		void loadDecryptedPassword()
	}, [
		secret.password,
		secret.login,
		secret.metadata,
		secret.binaryData,
		decryptPassword,
		reset,
	])

	const disabledSave = (): boolean => {
		const isLoginSame = formValues.login === secret.login
		const isPasswordSame = formValues.password === decryptedPassword
		const isMetadataSame =
			JSON.stringify(formValues.metadata || {}) === JSON.stringify(secret.metadata || {})
		const isBinaryDataSame = formValues.binaryData === secret.binaryData

		return isLoginSame && isPasswordSame && isMetadataSame && isBinaryDataSame
	}

	const onDelete = () => {
		void deleteSecret(secret.localId)
	}

	const onDownload = () => {
		if (!secret.binaryData) return

		const blob = new Blob([new Uint8Array(secret.binaryData)])

		const url = URL.createObjectURL(blob)
		const link = document.createElement('a')
		link.href = url

		const fileName = secret.metadata?.fileName || `file-${secret.localId}`
		const fileExtension = secret.metadata?.fileExtension || 'bin'
		link.download = `${fileName}.${fileExtension}`

		document.body.appendChild(link)
		link.click()

		document.body.removeChild(link)
		URL.revokeObjectURL(url)
	}

	const onSubmit = async (data: TFormData) => {
		try {
			if (data.binaryData instanceof File) {
				const { binaryData, metadata: fileMetadata } = await extractFileData(
					data.binaryData,
				)

				const secretData: TSecretForSave = {
					...data,
					metadata: {
						...data.metadata,
						...fileMetadata,
					},
					binaryData,
				}

				await updateSecret({
					localId: secret.localId,
					secret: secretData,
					cb: () => reset(),
				})
			} else {
				await updateSecret({
					localId: secret.localId,
					secret: data as TSecretForSave,
					cb: () => reset(),
				})
			}
		} catch (error) {
			// Ошибка уже обработана в store
		}
	}

	const onSave = handleSubmit(onSubmit)

	return (
		<SecretsItemView<TFormData>
			control={control}
			onSave={onSave}
			onDelete={onDelete}
			onDownload={!!secret.binaryData ? onDownload : undefined}
			disabled={disabledSave()}
			isEditMode
		/>
	)
})
