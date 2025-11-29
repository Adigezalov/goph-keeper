import { observer } from 'mobx-react-lite'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

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
	const { t } = useTranslation()
	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)

	const { updateSecret, decryptData, decryptBinaryData, deleteSecret } = store.secretsPage

	const [decryptedLogin, setDecryptedLogin] = useState('')
	const [decryptedPassword, setDecryptedPassword] = useState('')
	const [decryptedBinaryData, setDecryptedBinaryData] = useState<Uint8Array | undefined>(
		undefined,
	)

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
		const loadDecryptedData = async () => {
			try {
				const decryptedLoginValue = await decryptData(secret.login)
				const decryptedPasswordValue = await decryptData(secret.password)
				const decryptedBinaryDataValue = secret.binaryData
					? await decryptBinaryData(secret.binaryData)
					: undefined

				setDecryptedLogin(decryptedLoginValue)
				setDecryptedPassword(decryptedPasswordValue)
				setDecryptedBinaryData(decryptedBinaryDataValue)

				reset({
					login: decryptedLoginValue,
					password: decryptedPasswordValue,
					metadata: secret.metadata || {},
					binaryData: decryptedBinaryDataValue,
				})
			} catch (error) {
				console.error(t('secrets.decrypt_error'), error)
			}
		}

		void loadDecryptedData()
	}, [
		secret.password,
		secret.login,
		secret.metadata,
		secret.binaryData,
		decryptData,
		decryptBinaryData,
		reset,
		t,
	])

	const disabledSave = (): boolean => {
		const isLoginSame = formValues.login === decryptedLogin
		const isPasswordSame = formValues.password === decryptedPassword
		const isMetadataSame =
			JSON.stringify(formValues.metadata || {}) === JSON.stringify(secret.metadata || {})
		const isBinaryDataSame = formValues.binaryData === decryptedBinaryData

		return isLoginSame && isPasswordSame && isMetadataSame && isBinaryDataSame
	}

	const onDelete = () => {
		void deleteSecret(secret.localId)
	}

	const onDownload = () => {
		if (!decryptedBinaryData) return

		const arrayBuffer = new Uint8Array(decryptedBinaryData).buffer
		const blob = new Blob([arrayBuffer])

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
		} catch (_error) {
			// Error handled in updateSecret
		}
	}

	const onSave = handleSubmit(onSubmit)

	return (
		<SecretsItemView<TFormData>
			control={control}
			onSave={onSave}
			onDelete={onDelete}
			onDownload={decryptedBinaryData ? onDownload : undefined}
			disabled={disabledSave()}
			isEditMode
		/>
	)
})
