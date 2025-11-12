import { observer } from 'mobx-react-lite'
import { useForm } from 'react-hook-form'

import { StoreContextLogic, TStoreLogic, useStoreLogic } from '@shared/store'

import { TSecretForSave } from '../../types'
import { SecretsItemView } from './secrets-item.view.tsx'

export const SecretsItemNew = observer(() => {
	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)

	const { createSecret } = store.secretsPage

	const { control, handleSubmit, reset } = useForm<TSecretForSave>({
		mode: 'all',
		defaultValues: {
			login: '',
			password: '',
			metadata: '',
		},
	})

	const onSubmit = async (data: TSecretForSave) => {
		void createSecret({ secret: data, cb: () => reset() })
	}

	const onSave = handleSubmit(onSubmit)

	return <SecretsItemView<TSecretForSave> control={control} onSave={onSave} />
})
