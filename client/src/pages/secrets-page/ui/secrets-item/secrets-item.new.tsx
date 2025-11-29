import { observer } from 'mobx-react-lite'
import { useForm } from 'react-hook-form'

import { StoreContextLogic, TStoreLogic, useStoreLogic } from '@shared/store'

import { TSecretForSave } from '../../types'
import { extractFileData } from '../../utils'
import { SecretsItemView } from './secrets-item.view.tsx'

type TFormData = Omit<TSecretForSave, 'binaryData'> & {
	binaryData?: File
}

export const SecretsItemNew = observer(() => {
	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)

	const { createSecret } = store.secretsPage

	const { control, handleSubmit, reset } = useForm<TFormData>({
		mode: 'all',
		defaultValues: {
			login: '',
			password: '',
			metadata: {},
			binaryData: undefined,
		},
	})

	const onSubmit = async (data: TFormData) => {
		if (data.binaryData) {
			const { binaryData, metadata: fileMetadata } = await extractFileData(data.binaryData)

			const secretData: TSecretForSave = {
				...data,
				metadata: {
					...data.metadata,
					...fileMetadata,
				},
				binaryData,
			}

			void createSecret({ secret: secretData, cb: () => reset() })
		} else {
			const secretData: TSecretForSave = {
				...data,
				binaryData: undefined,
			}
			void createSecret({ secret: secretData, cb: () => reset() })
		}
	}

	const onSave = handleSubmit(onSubmit)

	return <SecretsItemView<TFormData> control={control} onSave={onSave} />
})
